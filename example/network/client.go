package network

import (
	"context"
	"sync"
	"time"

	"github.com/Gonewithmyself/gobot/pkg/logger"

	"github.com/fish-tennis/gnet"
	"github.com/golang/protobuf/proto"
)

type Callbacks interface {
	OnConnected()
	OnDisconnected()
	OnRecvPacket(*Message)
	// CreateHeartBeatPacket() gnet.Packet
}

type Client struct {
	conn    gnet.Connection
	donec   chan struct{}
	cancel  context.CancelFunc
	addr    string // 连接地址
	gamer   string // 连接对应的玩家标识
	handler Callbacks
}

const (
	ClientSendPacketCacheCap = uint32(32)
	ClientSendBufferSize     = uint32(1024 * 16)  // 16K
	ClientRecvBufferSize     = uint32(1024 * 8)   // 8K
	ClientMaxPacketSize      = uint32(1024 * 128) // 128K
	ClientRecvTimeout        = uint32(60)         // 60s
	retryCount               = 10
)

func NewClient(addr, gamer string, cb Callbacks) *Client {
	c := &Client{
		donec:   make(chan struct{}),
		addr:    addr,
		gamer:   gamer,
		handler: cb,
	}
	connectionConfig := gnet.ConnectionConfig{
		SendPacketCacheCap: ClientSendPacketCacheCap,
		SendBufferSize:     ClientSendBufferSize, // 16K
		RecvBufferSize:     ClientRecvBufferSize, // 8K
		MaxPacketSize:      ClientMaxPacketSize,  // 128K
		RecvTimeout:        ClientRecvTimeout,
	}
	conn := gnet.NewTcpConnector(&connectionConfig,
		NewCodec(), c)
	conn.SetTag(gamer)
	c.conn = conn
	return c
}

func (c *Client) Connect() bool {
	var connectOK bool
	for i := 0; i < retryCount; i++ {
		if c.conn.Connect(c.addr) {
			connectOK = true
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	return connectOK
}

func (c *Client) OnConnected(connection gnet.Connection, success bool) {
	if c.handler != nil {
		c.handler.OnConnected()
	}
	// logger.Debug("cOnConnected ")
}

func (c *Client) OnDisconnected(connection gnet.Connection) {
	if c.handler != nil {
		c.handler.OnDisconnected()
	}
	// logger.Debug("cOnDisconnected ")
}

func (c *Client) OnRecvPacket(connection gnet.Connection, packet gnet.Packet) {
	if c.handler != nil {
		c.handler.OnRecvPacket(packet.(*Message))
	}
}

func (c *Client) CreateHeartBeatPacket(connection gnet.Connection) gnet.Packet {
	return nil
}

func (c *Client) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	wg := sync.WaitGroup{}
	c.conn.Start(ctx, &wg, func(connection gnet.Connection) {})

	logger.Info("gnet start", "gamer", c.gamer)

	wg.Wait()

	logger.Info("gnet exit", "gamer", c.gamer)
}

func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Client) Send(msg proto.Message) bool {
	return c.conn.SendPacket(&Message{
		Pkt: msg,
	})
}
