package gobot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/Gonewithmyself/gobot/pkg/btree"
)

// 热更配置
func (app *App) JsReloadConfig() (ret interface{}, err error) {
	// parseAppConf()
	// robot.LoadTree(filepath.Join(CachePath, types.Conf.Tree))
	// logger.Debug("reload")
	return
}

// 读文件
func (app *App) JsReadFile(fname string) (ret interface{}, err error) {
	_, err = os.Stat(fname)
	if err != nil {
		dir, _ := filepath.Split(fname)
		os.MkdirAll(dir, os.ModeDir)
		os.Create(fname)
	}
	var buf []byte
	buf, err = ioutil.ReadFile(fname)
	ret = string(buf)
	return
}

// 写文件
func (app *App) JsWriteFile(fname, data string) (ret interface{}, err error) {
	err = ioutil.WriteFile(fname, []byte(data), os.ModePerm|os.ModeTemporary)
	return
}

// 发送消息
func (app *App) JsSendReq(req, data string) (ret interface{}, err error) {
	return
}

// 拉取区服
func (app *App) JsFetchZones(confJs string) (ret interface{}, err error) {
	return
}

// 拉取所有消息列表
func (app *App) JsGetMsgList() (ret interface{}, err error) {
	return
}

// 获取消息详情
func (app *App) JsGetMsgDetail(msg string) (ret interface{}, err error) {
	return
}

// 普通模式运行
func (app *App) JsStartRobot(treeID string, tab float64, confJs string) (ret interface{}, err error) {
	ins := app.ins
	gamer, err := ins.CreateGamer(confJs, -1)
	if err != nil {
		return
	}
	gamer.SetTab(int32(tab))

	app.Mode = NormalMode
	app.UI.SetSilent(false)

	uid := gamer.GetUid()
	old, ok := app.Gamers[uid]
	if ok {
		old.Close()
	}

	go ins.RunGamer(gamer, app.Trees.Get(treeID))
	app.Gamers[uid] = gamer
	return
}

type stressWeight struct {
	Id     string `json:"id,omitempty"`
	Weight int32  `json:"weight,omitempty"`
}

// 运行压测
func (app *App) JsStressRobot(confJs, stressJS, weightJS, appJs string) (ret interface{}, err error) {
	var stressInfo struct {
		TreeID string `json:"tree_id,omitempty"`
		Start  int32  `json:"start,omitempty"`
		Count  int32  `json:"count,omitempty"`
	}
	err = json.Unmarshal([]byte(stressJS), &stressInfo)
	if err != nil {
		return
	}

	err = app.ResetWeight(weightJS)
	if err != nil {
		return
	}

	if !atomic.CompareAndSwapInt32(&app.State, 0, 1) {
		err = fmt.Errorf("already Start")
		return
	}
	app.Mode = StressMode
	app.UI.SetSilent(true)

	app.ins.ResetConfig(appJs)
	app.ins.StressStart(stressInfo.Start, stressInfo.Count, stressInfo.TreeID, confJs)
	return
}

func (app *App) ResetWeight(js string) error {
	var weights []stressWeight
	err := json.Unmarshal([]byte(js), &weights)
	for _, wt := range weights {
		tr := app.Trees.Get(wt.Id)
		if tr == nil {
			continue
		}
		tr.Properties.Set("weight", wt.Weight)
	}
	return err
}

// 停止运行
func (app *App) JsStop() (ret interface{}, err error) {
	trees := btree.NewTreeMgr()
	if er := trees.LoadProject(app.Options.TreeFile); er == nil {
		app.Trees = trees
	}

	for _, gamer := range app.Gamers {
		if app.Mode == StressMode {
			gamer.Stop()
		} else {
			gamer.Close()
		}
		delete(app.Gamers, gamer.GetUid())
	}
	app.ins.OnClickStop()
	return
}
