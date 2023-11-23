package gamer

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Gonewithmyself/gobot/example/network"

	"github.com/Gonewithmyself/gobot/pkg/logger"
	"github.com/Gonewithmyself/gobot/pkg/ratelimit"

	"github.com/golang/protobuf/proto"
)

func (g *Gamer) OnConnected() {
	g.status = stateConnectedLogic
	logger.Debug("OnConnected")
}

func (g *Gamer) OnDisconnected() {
	g.Close()
	logger.Debug("OnDisconnected")
}

// 收到网络层包 转发到玩家协程
func (g *Gamer) OnRecvPacket(msg *network.Message) {
	g.MsgCh <- msg
}

func (g *Gamer) SendMsg(msg proto.Message) bool {
	if g.conn == nil {
		return false
	}

	tp := reflect.TypeOf(msg).Elem()
	if !ratelimit.Consume(tp.Name()) {
		return false
	}

	if g.conn.Send(msg) {
		g.RecordReq(network.NbaPbInfo.CsType2Cmd[tp])
		g.LogReq(tp.Name(), msg)
		return true
	}

	return false
}

type (
	Handler    map[reflect.Type]func(*Gamer, interface{}) // 消息回调
	ErrHandler map[string]func(*Gamer, string, string)    // 错误码回调
)

func (r Handler) Handle(msg *network.Message, g *Gamer) {
	// sc := msg.Pkt
	// tp := reflect.TypeOf(sc).Elem()
	// msgName := tp.Name()
	// if code := msg.Hd.Error; code != 0 {
	// 	// 错误码处理
	// 	codeName := pb.ErrorCode_name[int32(code)]
	// 	if h, ok := errRouter[codeName]; ok {
	// 		h(g, codeName, msgName)
	// 	}
	// 	return
	// }

	// h, ok := r[tp]
	// if ok {
	// 	h(g, sc)
	// }

	// idx := strings.Index(msgName, "S2C")
	// if proto.MessageType(msgName[:idx]+"C2S") == nil {
	// 	g.LogNtf(msgName, sc)
	// } else {
	// 	g.LogRsp(msgName, sc)
	// }
}

var (
	router    = Handler{}
	errRouter = ErrHandler{}
)

func init() {
	autoRegisterHandler()

}

func autoRegisterHandler() {
	tp := reflect.TypeOf(&Gamer{})
	for i := 0; i < tp.NumMethod(); i++ {
		method := tp.Method(i)
		fn := method.Func.Interface()
		name := method.Name
		switch {
		case strings.HasSuffix(name, "S2C"):
			// sc消息回调
			msgtp := proto.MessageType(name)
			if msgtp == nil {
				panic(fmt.Sprintf("msg(%v)NotFound", name))
			}
			router[msgtp.Elem()] = fn.(func(*Gamer, interface{}))

		case strings.HasPrefix(name, "Err"):
			// 错误码回调
			errRouter[name] = fn.(func(*Gamer, string, string))
		}
	}
}

func (g *Gamer) RecordReq(cmd uint16) {
	g.rttmap[cmd] = time.Now().UnixNano()
	rec := network.GetRecorder(cmd)
	rec.UpCounter.Inc(1)
}

func (g *Gamer) RecordRsp(cmd uint16) {
	now := time.Now().UnixNano()
	rec := network.GetRecorder(cmd)
	rec.DownCounter.Inc(1)
	reqTs := g.rttmap[cmd]
	if reqTs == 0 {
		return
	}

	rtt := now - reqTs
	if rtt < 0 {
		return
	}
	rec.RTTRecorder.Update(time.Nanosecond * time.Duration(rtt))
	g.rttmap[cmd] = 0
}
