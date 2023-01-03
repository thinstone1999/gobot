package back

import (
	"sync"

	"github.com/Gonewithmyself/gobot/front"
	"github.com/Gonewithmyself/gobot/pkg/logger"
)

// 一个玩家
type IGamer interface {
	GetUid() string // 玩家唯一标识，登录前就要确定
	Close()         //
	OnExit()

	MsgChan() <-chan interface{} // 消息channel
	ExitChan() <-chan struct{}   // exit channel
	ProcessMsg(interface{})      // 网络消息处理函数

	GetTickMs() int64
	IsStopped() bool
	Stop()

	SetUI(ui *front.UI)
	SetTab(tab int32)
}

type Gamer struct {
	MsgCh  chan interface{}
	ExitCh chan struct{}
	UI     *front.UI
	Silent bool
	// stop后不再发送消息，但继续收消息
	// 提升时延统计的精度 避免有的消息刚发出 没收到回包就停了
	Stopped   bool
	Tab       int32
	closeOnce sync.Once
}

const (
	msgChanCap = 64
)

func NewGamer() *Gamer {
	return &Gamer{
		MsgCh:  make(chan interface{}, msgChanCap),
		ExitCh: make(chan struct{}),
	}
}

//
func (g *Gamer) GetUid() string {
	return ""
}

func (g *Gamer) Close() {
	g.closeOnce.Do(func() {
		close(g.ExitCh)
	})

}

func (g *Gamer) OnExit() {
}

func (g *Gamer) MsgChan() <-chan interface{} {
	return g.MsgCh
}

func (g *Gamer) ExitChan() <-chan struct{} {
	return g.ExitCh
}

func (g *Gamer) ProcessMsg(interface{}) {

}

func (g *Gamer) GetTickMs() int64 {
	return 5000
}

func (g *Gamer) IsStopped() bool {
	return g.Stopped
}

func (g *Gamer) Stop() {
	g.Stopped = true
}

func (g *Gamer) SetUI(ui *front.UI) {
	g.UI = ui
}

func (g *Gamer) SetTab(tab int32) {
	g.Tab = tab
}

func (g *Gamer) LogReq(name string, info interface{}) {
	if g.canLog() {
		g.UI.LogReq(name, g.Tab, info)
	}
}

func (g *Gamer) LogRsp(name string, info interface{}) {
	if g.canLog() {
		g.UI.LogRsp(name, g.Tab, info)
	}
}

func (g *Gamer) LogNtf(name string, info interface{}) {
	if g.canLog() {
		g.UI.LogNtf(name, g.Tab, info)
	}
}

func (g *Gamer) LogError(name string, info interface{}) {
	if g.canLog() {
		g.UI.LogError(name, g.Tab, info)
	}
}

func (g *Gamer) ChangeStatus(name, status, info string) {
	if g.canLog() {
		g.UI.UIChangeStatus(g.Tab, name, status, info)
	}
	logger.Debug("ChangeStatus", "n", name, "s", status, "i", info)
}

func (g *Gamer) canLog() bool {
	return g.UI != nil && !g.Silent
}
