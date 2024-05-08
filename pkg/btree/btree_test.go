package btree

import (
	"fmt"
	"gobot/pkg/util"
	"reflect"
	"strings"
	"testing"
)

func TestNewTreeMgr(t *testing.T) {
	RegisterAction(&Gamer{})
	t.Log(len(actionRouter))
	mgr := NewTreeMgr()
	err := mgr.LoadProject("demo.b3")
	if err != nil {
		panic(err)
	}

	tr := mgr.GetByTitle("root")
	t.Log(tr == nil)

	tr.Properties.Set("Run", "1")
	g := &Gamer{}

	st := util.NewMap()
	for i := 1; i < 20; i++ {
		st.Set("_term", i)
		tr.Tick(g, st)
	}

}

type Worker struct {
	Action
}

func (node *Worker) OnTick(tick *Tick) Status {
	gamer := tick.UserData.(*Gamer)
	hd, ok := actionRouter[node.Title]
	if !ok {
		panic(fmt.Errorf("action(%v) NotImplement", node.Title))
	}

	return hd(gamer, node, tick)
}

type Gamer struct {
	running int32
}

func (g *Gamer) LoginAction(node *Worker, tick *Tick) Status {
	fmt.Println(tick.Store.GetInt("_term"), "Login")
	return SUCCESS
}

func (g *Gamer) CreateFightAction(node *Worker, tick *Tick) Status {
	fmt.Println(tick.Store.GetInt("_term"), "CreateFightAction")
	return SUCCESS
}

func (g *Gamer) FightingAction(node *Worker, tick *Tick) Status {
	fmt.Println(tick.Store.GetInt("_term"), "FightingAction")

	g.running++
	if g.running%3 == 0 {
		g.running = 0
	}

	if g.running != 0 {
		return RUNNING
	}
	return SUCCESS
}

func (g *Gamer) CleanAction(node *Worker, tick *Tick) Status {
	fmt.Println(tick.Store.GetInt("_term"), "CleanAction")

	g.running++
	if g.running%3 == 0 {
		g.running = 0
	}

	if g.running != 0 {
		return RUNNING
	}
	return SUCCESS
}

type actionHandler map[string]func(g *Gamer, node *Worker, tick *Tick) Status

var actionRouter = actionHandler{}

func RegisterAction(gamer *Gamer) {
	Register(&Worker{})

	rt := reflect.TypeOf(gamer)
	if rt.Kind() != reflect.Ptr && rt.Elem().Kind() != reflect.Struct {
		panic("mustbe struct pointer")
	}
	suffix := "Action"
	for i := 0; i < rt.NumMethod(); i++ {
		name := rt.Method(i).Name
		if strings.HasSuffix(name, suffix) {
			method := rt.Method(i).Func.Interface()
			fn := method.(func(*Gamer, *Worker, *Tick) Status)
			actionRouter[name] = fn
		}
	}
}
