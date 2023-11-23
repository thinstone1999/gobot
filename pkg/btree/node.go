package btree

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/Gonewithmyself/gobot/pkg/util"
)

type Status uint8

const (
	SUCCESS Status = iota
	RUNNING
	ERROR
)

type NodeCategory string

const (
	COMPOSITE NodeCategory = "composite" // 可包含多个叶子节点
	DECORATOR NodeCategory = "decorator" // 只有一个叶子节点
	ACTION    NodeCategory = "action"    // 叶子节点
	TREE      NodeCategory = "tree"      // 子树 指向具体的树
)

type Node struct {
	Id       string       `json:"id"`       // 唯一ID
	Name     string       `json:"name"`     // 需对应一个INode的具体实现 (类)
	Title    string       `json:"title"`    // 对外显示的名字 （实例）
	Category NodeCategory `json:"category"` // 类别 主要用于区分中间节点

	Description string   `json:"description"`
	Children    []string `json:"children"`
	Child       string   `json:"child"`
	Properties  util.Map `json:"properties"` // 参数
}

// 控制某个用例开关
func (node *Node) Disabled() bool {
	if node.Category != TREE {
		return false
	}

	return node.Properties.GetInt("Run") == 0
}

type INode interface {
	OnEnter(*Tick)       // 进入
	OnTick(*Tick) Status // 主要业务逻辑
	OnLeave(*Tick)       // 离开

	execute(*Tick) Status // 节点控制逻辑

	getCfg() *Node
	init(node *Node, worker INode)
}

type baseNode struct {
	Node         // 配置数据
	worker INode // 方法集合
}

func (node *baseNode) getCfg() *Node {
	return &node.Node
}

func (node *baseNode) init(cfg *Node, worker INode) {
	node.Node = *cfg
	node.worker = worker
}

func (node *baseNode) OnEnter(*Tick) {

}

func (node *baseNode) OnTick(tick *Tick) Status {
	log.Println("tick", node.Title)
	return SUCCESS
}

func (node *baseNode) OnLeave(*Tick) {

}

func (node *baseNode) execute(tick *Tick) Status {
	nodesData := tick.GetOpenNodes()

	status, ok := nodesData[node.Id]
	if ok {
		return status
	}
	defer func() {
		nodesData[node.Id] = status
	}()

	node.worker.OnEnter(tick)
	status = node.worker.OnTick(tick)
	if status == RUNNING {
		tick.GetRunningNodes()[node] = struct{}{}
	} else {
		node.worker.OnLeave(tick)
	}

	return status
}

// 叶子节点 暴露给外部节点继承
type Action struct {
	baseNode
}

// 子树
type SubTree struct {
	baseNode
}

func (node *SubTree) OnTick(tick *Tick) Status {
	tree, ok := mgr.trees.Load(node.Name)
	if !ok {
		return ERROR
	}

	if node.Disabled() {
		return SUCCESS
	}

	return tree.(*Tree).root.execute(tick)
}

// 节点名对应的实现
type implNodes map[string]reflect.Type

var nodeTypes = make(implNodes)

// 注册节点名及其实现
func Register(node INode) {
	tp := reflect.TypeOf(node).Elem()
	nodeTypes[tp.Name()] = reflect.TypeOf(node).Elem()
}

// 根据配置 构建实际节点
func FactoryNode(node *Node) INode {
	if node.Category == TREE {
		func() {
			defer func() {
				r := recover()
				if r != nil {
					fmt.Println(node.Name)
					panic(r)
				}
			}()
			v, _ := mgr.trees.Load(node.Name)
			treeCfg := v.(*Tree)
			node.Properties = treeCfg.Properties
		}()

		return &SubTree{
			baseNode: baseNode{Node: *node},
		}
	}

	tp, ok := nodeTypes[node.Name]
	if !ok {
		panic(fmt.Sprintf("%v not register", node.Name))
	}

	return reflect.New(tp).Interface().(INode)
}

// x分钟后开始运行后续节点
type StartAfter struct {
	Action
}

func (node *StartAfter) OnTick(tick *Tick) Status {
	k := node.Id + "_startAt"
	startAt := tick.Store.GetInt64(k)
	now := time.Now()
	if startAt == 0 {
		waitMin := node.Properties.GetInt64("min")
		y, month, d := now.Date()
		h, min := now.Hour(), now.Minute()
		startM := min + int(waitMin)
		start := time.Date(y, month, d, h, startM, 0, 0, time.Local)
		startAt = start.Unix()
		tick.Store.Set(k, startAt)
	}
	if now.Unix() >= startAt {
		return SUCCESS
	}
	return ERROR
}

func init() {
	Register(&StartAfter{})
}
