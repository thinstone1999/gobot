package gobot

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Gonewithmyself/gobot/back"
	"github.com/Gonewithmyself/gobot/pkg/btree"
	"github.com/Gonewithmyself/gobot/pkg/util"
)

type Worker struct {
	btree.Action
}

type actionHandler map[string]reflect.Value

var actionRouter = actionHandler{}

func RegisterAction(gamer back.IGamer) {
	rt := reflect.TypeOf(gamer)
	for i := 0; i < rt.NumMethod(); i++ {
		method := rt.Method(i)
		if strings.HasSuffix(method.Name, "Action") {
			fn := method.Func
			// 检查参数
			// for i := 0; i < fn.Type().NumIn(); i++ {
			// 	fn.Type().In()
			// }
			actionRouter[strings.TrimSuffix(method.Name, "Action")] = fn
		}
	}
}

func (node *Worker) OnTick(tick *btree.Tick) btree.Status {
	gamer := tick.UserData.(back.IGamer)
	hd, ok := actionRouter[node.Title]
	if !ok {
		panic(fmt.Errorf("action(%v) NotImplement", node.Title))
	}

	if node.inCd(tick.Store) {
		return btree.SUCCESS
	}

	out := hd.Call([]reflect.Value{reflect.ValueOf(gamer), reflect.ValueOf(node), reflect.ValueOf(tick)})
	status := out[0].Interface().(btree.Status)
	return status
}

func (node *Worker) inCd(store util.Map) bool {
	cd := node.Properties.GetInt64("cd")
	if cd == 0 {
		return false
	}

	k := node.Id + "_ticktime"
	lastTick := store.GetInt64(k)
	now := time.Now().Unix()
	if lastTick != 0 && now-lastTick < cd {
		return true
	}

	store.Set(k, now)
	return false
}
