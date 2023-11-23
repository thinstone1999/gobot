package gamer

import (
	"strings"

	"github.com/Gonewithmyself/gobot/example/network"

	"github.com/Gonewithmyself/gobot"
	"github.com/Gonewithmyself/gobot/pkg/btree"
	"github.com/Gonewithmyself/gobot/pkg/logger"
)

type state int8

func (s state) String() string {
	if s < stateAuthOk {
		return "offline"
	}
	return "online"
}

const (
	stateIdle           state = iota // 离线
	stateAuthOk                      // 认证
	stateConnectedLogic              // 连接上logic
	stateLoginOk                     // 登录成功
)

// 认证
func (g *Gamer) AuthAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	if g.status > stateIdle {
		return btree.SUCCESS
	}

	// logger.Debug("begin Auth", "uid", g.GetUid())
	// rsp, err := g.Auth()
	// if err != nil {
	// 	logger.Error("authErr", "err", err)
	// 	return btree.ERROR
	// }

	// for _, zone := range rsp.Zones {
	// 	if zone.GetId() == g.authData.Conf.Zone {
	// 		g.authRsp = rsp
	// 		g.status = stateAuthOk
	// 		g.ZoneName = *zone.Name
	// 		return btree.SUCCESS
	// 	}
	// }
	// g.Close()
	// logger.Error("zoneNotFound", "want", g.authData.Conf.Zone, "got", rsp.Zones)
	return btree.ERROR
}

func (g *Gamer) Auth() (*network.Authrsp, error) {
	auth := network.NewAuther(g.authData)
	sdkdata, err := auth.SdkAuth(nil)
	if err != nil {
		return nil, err
	}

	rsp, err := auth.GameAuth(sdkdata)
	if err != nil {
		return nil, err
	}
	return rsp.(*network.Authrsp), nil
}

// 连接logic
func (g *Gamer) ConnectAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	if g.status != stateAuthOk {
		return btree.SUCCESS
	}

	addr := strings.Split(g.authRsp.Addr, "://")[1]
	conn := network.NewClient(addr, g.authRsp.SdkUID, g)
	logger.Debug("Connect")
	if !conn.Connect() {
		return btree.ERROR
	}

	go conn.Run()
	g.conn = conn
	return btree.SUCCESS
}

// 发送登录请求
func (g *Gamer) LoginAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	if g.status != stateConnectedLogic || g.sentlogin {
		return btree.SUCCESS
	}

	return btree.SUCCESS
}

func (g *Gamer) HeartBeatAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	// g.SendMsg(&pb.GamerHeartC2S{})
	return btree.SUCCESS
}

// func (g *Gamer) GamerLoginS2C(val interface{}) {
// 	// sc := val.(*pb.GamerLoginS2C)
// 	// g.Gid = int64(sc.GetId())

// 	// g.status = stateLoginOk
// 	// logger.Info("loginOk")
// 	// g.changeStatus("success")
// 	// if sc.TeamInfo == nil {
// 	// 	// 创建角色
// 	// 	g.SendMsg(&pb.GamerLoginCreateRoleC2S{
// 	// 		Name: pb.String(util.GenChineseName(4)),
// 	// 	})
// 	// }
// }
