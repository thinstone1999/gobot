package btree

import "github.com/Gonewithmyself/gobot/pkg/util"

type Tick struct {
	Store    util.Map
	UserData interface{}
}

func (tick *Tick) GetOpenNodes() map[string]Status {
	openNodes := tick.Store.Get("_openNodes")
	if openNodes == nil {
		openNodes = make(map[string]Status)
		tick.Store.Set("_openNodes", openNodes)
	}
	return openNodes.(map[string]Status)
}

func (tick *Tick) GetRunningNodes() map[INode]struct{} {
	openNodes := tick.Store.Get("_runningNodes")
	if openNodes == nil {
		openNodes = make(map[INode]struct{})
		tick.Store.Set("_runningNodes", openNodes)
	}
	return openNodes.(map[INode]struct{})
}

func (tick *Tick) Reset() {
	tick.Store.Set("_openNodes", nil)
	tick.Store.Set("_runningNodes", nil)
}

func (tick *Tick) GetIndexMap() map[string]int {
	data := tick.Store.Get("_indexMap")
	if data == nil {
		data = make(map[string]int)
		tick.Store.Set("_indexMap", data)
	}
	return data.(map[string]int)
}

func (tick *Tick) GetRandNode() map[string]map[int]struct{} {
	data := tick.Store.Get("_randNode")
	if data == nil {
		data = make(map[string]map[int]struct{})
		tick.Store.Set("_randNode", data)
	}
	return data.(map[string]map[int]struct{})
}
