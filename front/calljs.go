package front

import (
	"fmt"
	"sync/atomic"
)

// GO 调用 js提供的函数
func (app *UI) CallJS(req *Request) *Response {
	app.UI.Eval(fmt.Sprintf("GOAgent.Invoke(%s)", req))
	return nil
}

func (app *UI) UIChangeStatus(tab int32, name, status, info string) {
	data := map[string]interface{}{
		"name":   name,
		"status": status, // error 红色 success 绿色
		"info":   info,
		"tab":    tab,
	}

	app.CallJS(NewRequest("info.status", data))
}

type UILog struct {
	Name string `json:"name"`
	Cate string `json:"cate"` // hall pvp
	Id   int64  `json:"id"`   //
	Type string `json:"type"` // ntf req rsp error
	Tab  int32  `json:"tab"`  // 所属页签
}

func (app *UI) Log(meta UILog, info interface{}, err ...interface{}) {
	if app.silent {
		return
	}
	app.doLog(meta, info, err...)
}

func (app *UI) doLog(meta UILog, info interface{}, err ...interface{}) {
	meta.Id = atomic.AddInt64(&app.logId, 1)
	l := map[string]interface{}{
		"meta": meta,
		"data": info,
	}

	app.CallJS(NewRequest("info.log", l))
}

func (app *UI) LogError(name string, tab int32, info interface{}) {
	app.Log(UILog{
		Name: name,
		Cate: "hall",
		Tab:  tab,
		Type: "error",
	}, info, "error")
}

func (app *UI) LogRsp(name string, tab int32, info interface{}) {
	app.Log(UILog{
		Name: name,
		Cate: "hall",
		Tab:  tab,
		Type: "rsp",
	}, info)
}

func (app *UI) LogReq(name string, tab int32, info interface{}) {
	app.Log(UILog{
		Name: name,
		Cate: "hall",
		Tab:  tab,
		Type: "req",
	}, info)
}

func (app *UI) LogNtf(name string, tab int32, info interface{}) {
	app.Log(UILog{
		Name: name,
		Cate: "hall",
		Tab:  tab,
		Type: "ntf",
	}, info)
}

// ignore silent
func (app *UI) Print(name string, info interface{}) {
	app.doLog(UILog{
		Name: name,
		Cate: "stress",
		// Title: name,
		Type: "ntf",
		Tab:  0,
	}, info)
}
