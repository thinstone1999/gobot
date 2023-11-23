package gamer

import (
	"fmt"
	"reflect"

	"github.com/Gonewithmyself/gobot/example/network"
	"github.com/Gonewithmyself/gobot/example/types"

	"github.com/Gonewithmyself/gobot/back"
	"github.com/Gonewithmyself/gobot/pkg/logger"
)

// 一个玩家
type Gamer struct {
	*back.Gamer
	Seq       int32
	authData  *types.AuthData
	authRsp   *network.Authrsp
	status    state
	conn      *network.Client
	sentlogin bool
	Role
	rttmap map[uint16]int64 // msgid -> 请求发送时间
}

type (
	Role struct {
		Name     string
		Gid      int64
		Levle    int32
		ZoneName string
	}
)

func NewGamer(info *types.LoginConfig) *Gamer {
	account := fmt.Sprintf("%v", info.Account)
	auth := &types.AuthData{
		SdkAccount: account,
		Conf:       info,
	}

	return &Gamer{
		authData: auth,
		Gamer:    back.NewGamer(),
		rttmap:   map[uint16]int64{},
	}
}

func NewGamerBySeq(seq int32, info *types.LoginConfig) *Gamer {
	auth := &types.AuthData{
		SdkAccount: fmt.Sprintf("%v_%v", types.SdkConf.Prefix, seq),
		Conf:       info,
	}
	return &Gamer{
		authData: auth,
		Seq:      seq,
		Gamer:    back.NewGamer(),
		rttmap:   map[uint16]int64{},
	}
}

// 玩家唯一标识
func (g *Gamer) GetUid() string {
	return fmt.Sprintf("%v.%v", types.SdkConf.Name, g.authData.SdkAccount)
}

// 独立的玩家协程中处理网络消息
func (g *Gamer) ProcessMsg(data interface{}) {
	msg := data.(*network.Message)
	g.RecordRsp(msg.Hd.CmdAct())
	msgName := ""
	defer func() {
		r := recover()
		if r != nil {
			g.LogError(msgName, fmt.Sprintf("%v", msg))
			logger.Error("procMsg", "name", msgName, "msg", msg)
		}
	}()
	if msg.Pkt == nil {
		// 第一个包 或者 新加了协议没编译
		logger.Debug("nilPkt", "cmd", msg.Hd.CmdAct())
		return
	}

	msgName = reflect.TypeOf(msg.Pkt).Elem().Name()
	// logger.Debug("recv", "msg", msgName)
	router.Handle(msg, g)
}

func (g *Gamer) OnExit() {
	g.conn.Close()
	g.status = stateIdle
	g.conn = nil
	g.sentlogin = false
	g.changeStatus("error")
}

func (g *Gamer) changeStatus(status string) {
	str := "离线"
	if g.IsOnline() {
		str = "在线"
	}
	g.ChangeStatus(g.Role.Name,
		status,
		fmt.Sprintf("gid(%v) level(%v) zone(%v) | %v",
			g.Gid, g.Levle, g.ZoneName, str))
}

func (g *Gamer) IsOnline() bool {
	return g.status >= stateLoginOk
}
