package front

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/Gonewithmyself/gobot/pkg/logger"
)

// js调用GO 提供的函数
func FromJS(req *Request) (rsp *Response) {
	rsp = &Response{
		Seq: req.Seq,
	}
	defer func() {
		r := recover()
		if r != nil {
			rsp.Code = Failed
			rsp.Message = fmt.Sprintf("%v", r)
			logger.Error("FromJS", "r", r, "req", req)
		}

		// logger.Debug("FromJS", "req", req, "rsp", rsp)
	}()

	var (
		args []interface{}
		ok   bool
	)
	if args, ok = req.Arg.([]interface{}); !ok {
		args = []interface{}{req.Arg}
	}

	ret, err := Call(req.Method, args...)
	if err != nil {
		rsp.Code = Failed
		rsp.Message = err.Error()
		return
	}

	rsp.Data = ret
	return
}

// 注册函数供js调用
func Register(method string, fn reflect.Value) {
	if fn.Kind() != reflect.Func {
		panic("must be func")
	}
	router[method] = fn
}

func RegisterStruct(v interface{}) {
	rtv := reflect.ValueOf(v)
	rtp := rtv.Type()
	for i := 0; i < rtv.NumMethod(); i++ {
		method := rtv.Method(i)
		name := rtp.Method(i).Name
		if !strings.HasPrefix(name, APIPrifix) {
			continue
		}
		name = strings.Replace(name, APIPrifix, "", 1)
		Register(name, method)
	}
}

const (
	APIPrifix = "Js"
)

type Code int

const (
	Success Code = iota
	Failed
)

type (
	Request struct {
		Seq     int64       `json:"seq"`
		Method  string      `json:"method"`
		Message string      `json:"msg"`
		Arg     interface{} `json:"arg"`
	}
	Response struct {
		Seq     int64       `json:"seq"`
		Code    Code        `json:"code"`
		Message string      `json:"msg"`
		Data    interface{} `json:"data"`
	}
)

func (req *Request) String() string {
	data, _ := json.Marshal(req)
	return string(data)
}

func NewRequest(method string, data interface{}) *Request {
	return &Request{
		Method: method,
		// Message: msg,
		Arg: data,
	}
}

var router = map[string]reflect.Value{}

func Call(method string, args ...interface{}) (ret interface{}, err error) {
	fn, ok := router[method]
	if !ok {
		err = fmt.Errorf("method:(%s) notFound", method)
		return
	}
	typ := fn.Type()
	numIn := typ.NumIn()

	if numIn != 0 && numIn != len(args) {
		err = fmt.Errorf("wrong inParam num want(%v) give(%v)",
			numIn, len(args))
		return
	}

	inArgs := make([]reflect.Value, 0, numIn)
	for i := 0; i < numIn; i++ {
		inArgs = append(inArgs, reflect.ValueOf(args[i]))
	}
	vals := fn.Call(inArgs)
	outs := make([]interface{}, 0, len(vals))
	for i := 0; i < len(vals); i++ {
		outs = append(outs, vals[i].Interface())
	}

	ret = outs[0]
	err, _ = outs[1].(error)

	return
}
