package network

import (
	"bytes"
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/fish-tennis/gnet"
	"github.com/golang/protobuf/proto"
	gproto "google.golang.org/protobuf/proto"
)

type Header struct {
	Len   uint32 //数据长度,不包含head的长度
	Cmd   uint8  //命令
	Act   uint8  //动作
	Error uint16 //错误码
}

func (hd Header) CmdAct() uint16 {
	return uint16(CmdAct(hd.Cmd, hd.Act))
}

func (hd *Header) Copy() *Header {
	var nhd Header
	nhd = *hd
	return &nhd
}

type CSHeader struct {
	Len uint32 //数据长度,不包含head的长度
	Cmd uint8  //命令
	Act uint8  //动作
}

func (hd *CSHeader) Bytes() []byte {
	buf := make([]byte, csHeaderLen)
	copy(buf, (*(*[1024]byte)(unsafe.Pointer(hd)))[:csHeaderLen])
	return buf
}

func FromBytes(data []byte) *CSHeader {
	if len(data) != 6 {
		panic("notEnough")
	}
	return (*CSHeader)(unsafe.Pointer(&data[0]))
}

type Message struct {
	Hd  *Header
	Pkt proto.Message
	Err error
}

func (m *Message) Command() gnet.PacketCommand {
	return gnet.PacketCommand(m.Hd.CmdAct())
}

func (m *Message) Message() gproto.Message {
	//proto.Message
	return nil
}

func (m *Message) GetStreamData() []byte {
	return nil
}

func (m *Message) Clone() gnet.Packet {
	return &Message{}
}

const (
	headLen         = int(unsafe.Sizeof(Header{}))
	csHeaderLen     = 6
	scHeaderLen     = 8
	csbodyHeaderLen = csHeaderLen - uint32(unsafe.Sizeof(gnet.DefaultPacketHeader{}))
	scbodyHeaderLen = scHeaderLen - uint32(unsafe.Sizeof(gnet.DefaultPacketHeader{}))
)

type Codec struct {
	gnet.RingBufferCodec
	encriptor *EnDecriptor
}

func NewCodec() *Codec {
	codec := &Codec{
		RingBufferCodec: gnet.RingBufferCodec{},
	}
	codec.DataEncoder = codec.EncodePacket
	codec.DataDecoder = codec.DecodePacket
	codec.HeaderDecoder = codec.DecodeHeader
	codec.HeaderEncoder = codec.EncodeHeader
	return codec
}

/*
gnetMsg = 4byte header + body
sc = 8byte header + body
cs = 6byte header + body
*/

// 已读取4byte
func (c *Codec) DecodeHeader(conn gnet.Connection, buf []byte) {
	header := &gnet.DefaultPacketHeader{}
	header.ReadFrom(buf)                                // gnet默认前32字节作为包体长度
	header.LenAndFlags = header.Len() + scbodyHeaderLen // 项目的实际包体长度
	header.WriteTo(buf)
}

func (c *Codec) DecodePacket(conn gnet.Connection, header gnet.PacketHeader, buf []byte) gnet.Packet {
	// body header
	bodyHeader := &Header{
		Cmd: buf[0],
		Act: buf[1],
	}

	bodyHeader.Error = binary.LittleEndian.Uint16(buf[2:4])

	if c.encriptor == nil {
		// 第一个包
		c.encriptor = NewEnDecriptor(binary.BigEndian.Uint32(buf[4:8]), binary.BigEndian.Uint32(buf[8:]))
		return &Message{
			Hd: bodyHeader,
		}
	}

	cmdact := bodyHeader.CmdAct()
	body := buf[scbodyHeaderLen:]
	body, _ = c.encriptor.Decript(body)
	msg, err := NbaPbInfo.DecodeScMsg(cmdact, body)
	return &Message{
		Hd:  bodyHeader,
		Pkt: msg,
		Err: err,
	}
}

func (c *Codec) EncodePacket(conn gnet.Connection, packet gnet.Packet) [][]byte {
	pkt := packet.(*Message)
	csheader, body := NbaPbInfo.EncodeCsMsg(pkt.Pkt)
	header := csheader.Bytes()[4:]
	pkt.Hd = &Header{
		Cmd: csheader.Cmd,
		Act: csheader.Act,
	}

	for c.encriptor == nil {
		time.Sleep(time.Millisecond * 50)
	}

	body, _ = c.encriptor.Encript(body)
	// 这里的 header+body = gnetbody
	return [][]byte{header, body}
}

func (c *Codec) EncodeHeader(conn gnet.Connection, packet gnet.Packet, buf []byte) {
	header := &gnet.DefaultPacketHeader{}
	header.ReadFrom(buf)                                        // len =  EncodePacket中 len(header + body)
	header.LenAndFlags = header.Len() - uint32(csbodyHeaderLen) // 减去bodyheader
	header.WriteTo(buf)
}

// 加解密
type EnDecriptor struct {
	eseed uint32
	dseed uint32
}

const (
	// 伪随机数因子
	cryptA uint32 = 214013
	cryptB uint32 = 2531011
)

func NewEnDecriptor(eseed, dseed uint32) *EnDecriptor {
	return &EnDecriptor{eseed: eseed, dseed: dseed}
}

func (e *EnDecriptor) Decript(data []byte) ([]byte, error) {
	e.dseed = e.dseed*cryptA + cryptB
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, e.dseed)
	key := buf.Bytes()
	k := int32(0)
	c := byte(0)

	for i := 0; i < len(data); i++ {
		k %= 4
		x := (data[i] - c) ^ key[k]
		k++
		c = data[i]
		data[i] = x
	}

	return data, nil
}

func (e *EnDecriptor) Encript(data []byte) ([]byte, error) {
	e.eseed = e.eseed*cryptA + cryptB
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, e.eseed)
	key := buf.Bytes()
	k := int32(0)
	c := byte(0)

	for i := 0; i < len(data); i++ {
		k %= 4
		x := (data[i] ^ key[k]) + c
		k++
		c = x
		data[i] = c
	}

	return data, nil
}
