package network

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/Gonewithmyself/gobot/pkg/util"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// 预处理pb协议
type NbaPb struct {
	CsList []string               // 所有cs消息名
	CsMap  map[string]interface{} // 每个cs消息的默认值
	sync.Once
	CsType2Cmd map[reflect.Type]uint16
	CsCmd2Type map[uint16]reflect.Type
	ScCmd2Type map[uint16]reflect.Type
	Special    map[string]*util.SpecialInfo
}

func CmdAct(cmd, act uint8) int {
	return int(cmd)<<8 + int(act)
}

func ParseCmd(cmd uint16) (uint8, uint8) {
	return uint8((cmd & 0xff00) >> 8), uint8(cmd & 0x00ff)
}

func GetName(cmd uint16) string {
	tp, ok := NbaPbInfo.ScCmd2Type[cmd]
	if !ok {
		return ""
	}
	return tp.Name()
}

var NbaPbInfo NbaPb

func GetCmdActByProto(protoType reflect.Type) uint16 {
	typeName := protoType.Name()
	enumTypeName := fmt.Sprintf("%s_MsgID", typeName)
	enumValues := proto.EnumValueMap(enumTypeName)
	if enumValues != nil {
		enumValue, ok := enumValues["eMsgID"]
		if ok {
			return (uint16(enumValue))
		}
	}
	return 0
}

func (info *NbaPb) ProcessOneMsg(name string, tp reflect.Type) {
	dft := util.JsonDefault(tp)
	info.Special[name] = dft

	if strings.Contains(name, "C2S") {
		cmdAct := GetCmdActByProto(tp)
		info.CsCmd2Type[cmdAct] = tp
		info.CsType2Cmd[tp] = cmdAct
		info.CsMap[name] = dft.DefaultData
		info.CsList = append(info.CsList, name)
	}

	if strings.Contains(name, "S2C") {
		cmdAct := GetCmdActByProto(tp)
		info.ScCmd2Type[cmdAct] = tp
	}
}

func (info *NbaPb) ListMsg() []string {
	return info.CsList
}

func (info *NbaPb) GetMsgDefault(name string) interface{} {
	return info.CsMap[name]
}

func (info *NbaPb) Init() {
	info.CsMap = map[string]interface{}{}
	info.CsCmd2Type = make(map[uint16]reflect.Type)
	info.CsType2Cmd = make(map[reflect.Type]uint16)
	info.ScCmd2Type = make(map[uint16]reflect.Type)
	info.Special = make(map[string]*util.SpecialInfo)
}

func (info *NbaPb) DecodeScMsg(cmd uint16, data []byte) (proto.Message, error) {
	tp, ok := info.ScCmd2Type[cmd]
	if !ok {
		return nil, fmt.Errorf("%v scmsgNotFound", cmd)
	}
	msg := reflect.New(tp).Interface().(proto.Message)
	err := proto.Unmarshal(data, msg)
	return msg, err
}

func (info *NbaPb) EncodeCsMsg(msg proto.Message) (*CSHeader, []byte) {
	body, er := proto.Marshal(msg)
	if er != nil {
		panic(er)
	}

	tp := reflect.TypeOf(msg)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	cmdact, ok := info.CsType2Cmd[tp]
	if !ok {
		panic(fmt.Errorf("%v notRegister", tp.Name()))
	}
	cmd, act := ParseCmd(cmdact)
	return &CSHeader{
		Len: uint32(len(body)),
		Cmd: cmd,
		Act: act,
	}, body

}

func (info *NbaPb) GetCsMsgByJSON(name string, js string) proto.Message {
	tp := proto.MessageType(name)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}

	msg := reflect.New(tp).Interface()

	sp := info.Special[name]
	if !sp.HasSpecial {
		json.Unmarshal([]byte(js), msg)
		return msg.(proto.Message)
	}

	var tmp map[string]interface{}
	json.Unmarshal([]byte(js), &tmp)
	util.DecodeJSONmap(tmp, sp.SpecialData)
	data, _ := json.Marshal(tmp)
	json.Unmarshal(data, msg)
	return msg.(proto.Message)
}

func InitPbInfo() {
	NbaPbInfo.Once.Do(func() {
		NbaPbInfo.Init()
		protoregistry.GlobalFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
			for i := 0; i < fileDescriptor.Messages().Len(); i++ {
				desc := fileDescriptor.Messages().Get(i)
				name := string(desc.Name())
				if strings.HasSuffix(name, "C2S") || strings.HasSuffix(name, "S2C") {
					tp := proto.MessageType(name)
					if tp.Kind() == reflect.Ptr {
						tp = tp.Elem()
					}
					NbaPbInfo.ProcessOneMsg(name, tp)
				}
			}

			return true
		})
	})
}
