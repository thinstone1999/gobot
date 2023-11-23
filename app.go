package gobot

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Gonewithmyself/gobot/back"
	"github.com/Gonewithmyself/gobot/front"
	"github.com/Gonewithmyself/gobot/pkg/btree"
	"github.com/Gonewithmyself/gobot/pkg/logger"
	"github.com/Gonewithmyself/gobot/pkg/util"
)

type IApp interface {
	Init(*Options)
	Run(context.Context)
	Close()
	OnExit(reason string)

	// 根据配置文件创建一个玩家
	CreateGamer(confJson string, seq int32) (back.IGamer, error)
	RunGamer(g back.IGamer, tree *btree.Tree)

	ResetConfig(js string) // 前端传入部分配置
	OnClickStop()

	StressStart(start, count int32, treeID, confJs string) // 压测开始
	PrintStressStatus()                                    // 打印压测状态

	setInstance(IApp)
}

type RunMode int

const (
	NormalMode = iota // 普通模式
	StressMode        // 压测模式
)

type RunOs int

const (
	OsLinux RunOs = iota
	OsWindows
)

// 基础实现，默认linux
type BaseApp struct {
	*Options
	UI       *front.UI
	Fs       *front.FileServer
	SignalCh chan os.Signal
	ExitCh   chan struct{}
	sync.Once
	Trees  *btree.TreeManager
	Gamers map[string]back.IGamer
	Mode   RunMode
	Os     RunOs // 运行环境 linux/windows
	State  int32

	ins IApp // 实例
}

type App struct {
	BaseApp
}

func (app *BaseApp) setInstance(ins IApp) {
	app.ins = ins
}

func (app *BaseApp) PrintStressStatus() {

}

func (app *BaseApp) ResetConfig(js string) {

}

func (app *BaseApp) OnClickStop() {
	atomic.CompareAndSwapInt32(&app.State, 1, 0)
}

// 压测开始
func (app *BaseApp) StressStart(start, count int32, treeID, confJs string) {
	tree := app.Trees.Get(treeID)
	if tree == nil {
		tree = app.Trees.GetByTitle(treeID)
	}

	logger.Info("StressStart", "tree", tree.Title, "start", start, "count", count, "conf", confJs)

	for i := int32(0); i < count; i++ {
		gamer, err := app.ins.CreateGamer(confJs, start+i)
		if err != nil {
			logger.Error("CreateGamer", "err", err)
			continue
		}

		uid := gamer.GetUid()
		old, ok := app.Gamers[uid]
		if ok {
			old.Close()
		}
		app.Gamers[uid] = gamer
		go app.ins.RunGamer(gamer, tree)
	}
}

func (app *BaseApp) OnExit(reason string) {
	logger.Info("appExit", "reason", reason)
}

func (app *BaseApp) Close() {
	app.Do(func() {
		close(app.ExitCh)
	})
}

func (app *BaseApp) Run(ctx context.Context) {
	var reason string
	defer func() {
		app.ins.OnExit(reason)
	}()

	var uiDone <-chan struct{}
	if app.Os == OsWindows {
		// windows
		app.Fs.Start()
		app.UI.Load(app.Fs.Addr()) // access index.html
		uiDone = app.UI.Done()
	} else {
		// linux 环境
		app.ins.StressStart(0, 0, "", "")
	}

	logger.Info("appRun")
	tick := time.NewTicker(time.Duration(app.TickMs) * time.Millisecond)
	ctxDone := ctx.Done()
	for {
		select {
		case <-app.SignalCh:
			reason = "signal"
			return

		case <-ctxDone:
			ctxDone = nil
			reason = "ctxDone"
			logger.Info("ctxDone", "wait", app.StopWaitSec)
			// 让gamer停止发消息
			for _, gamer := range app.Gamers {
				gamer.Stop()
			}

			// 等待一段时间再退出
			time.AfterFunc(time.Second*time.Duration(app.StopWaitSec), func() {
				app.Close()
			})

		case <-app.ExitCh:
			reason = "appclose"
			return

		case <-uiDone:
			reason = "uiclose"
			return

		case <-tick.C:
			if app.Mode == StressMode {
				app.ins.PrintStressStatus()
			}
		}
	}
}

// 阻塞运行
func RunApp(ctx context.Context, app IApp, ops ...Option) {
	rand.Seed(time.Now().UnixNano())
	op := &Options{
		StaticDir:   StaticDir,
		TreeFile:    TreeFile,
		TickMs:      TickMs,
		StopWaitSec: StopWaitSec,
	}
	op.Apply(ops...)

	btree.Register(&Worker{})

	app.Init(op)
	app.setInstance(app)

	front.RegisterStruct(app)

	app.Run(ctx)
}

func (app *BaseApp) Init(opt *Options) {
	app.Os = OsLinux
	app.SignalCh = util.Listen()
	app.ExitCh = make(chan struct{})
	app.Options = opt
	app.Gamers = make(map[string]back.IGamer)
	app.Trees = btree.NewTreeMgr()
	if err := app.Trees.LoadProject(opt.TreeFile); err != nil {
		panic(err)
	}
}

func (app *BaseApp) CreateGamer(confJson string, seq int32) (back.IGamer, error) {
	return nil, errors.New("notImplement")
}

func (app *BaseApp) RunGamer(gamer back.IGamer, tree *btree.Tree) {
	var reason string
	timer := time.NewTimer(time.Millisecond * time.Duration(gamer.GetTickMs()))
	store := util.NewMap()
	defer func() {
		logger.Info("gamer exit", "uid", gamer.GetUid(), "reason", reason)
		gamer.OnExit()
	}()

	gamer.SetUI(app.UI)
	logger.Info("gamer start", "uid", gamer.GetUid())
	tree.Tick(gamer, store)

	for {
		select {
		case <-timer.C:
			if gamer.IsStopped() {
				// 停止后只收消息不发消息
				// 避免只统计到消息发送，服务还没回包就停了
				continue
			}

			if !tree.Tick(gamer, store) {
				logger.Warn("treeDisabled", tree.Title)
			}
			timer.Reset(time.Millisecond * time.Duration(gamer.GetTickMs()))

		case msg := <-gamer.MsgChan():
			gamer.ProcessMsg(msg)

		case <-gamer.ExitChan():
			return
		}
	}

}
