package btree

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/Gonewithmyself/gobot/pkg/util"
)

var mgr *TreeManager

type TreeManager struct {
	trees  sync.Map // id -> tree
	titles sync.Map // title -> tree
}

func NewTreeMgr() *TreeManager {
	mgr = &TreeManager{}
	return mgr
}

func (mgr *TreeManager) Get(id string) *Tree {
	v, ok := mgr.trees.Load(id)
	if ok {
		return v.(*Tree)
	}
	return nil
}

func (mgr *TreeManager) GetByTitle(title string) *Tree {
	var tr *Tree
	mgr.titles.Range(func(key, value interface{}) bool {
		tree := value.(*Tree)
		if strings.Contains(tree.Title, title) {
			tr = tree
			return false
		}
		return true
	})
	return tr
}

func (mgr *TreeManager) LoadProject(path string) error {
	pj, err := LoadProject(path)
	if err != nil {
		return err
	}

	for _, tree := range pj.Data.Trees {
		tree.buildRoot()
		mgr.trees.Store(tree.ID, tree)
		mgr.titles.Store(tree.Title, tree)
	}

	return nil
}

type Tree struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Root        string          `json:"root"`
	Properties  util.Map        `json:"properties"`
	Nodes       map[string]Node `json:"nodes"`

	root INode
}

func Load(jsonfile string) (*Tree, error) {
	data, err := ioutil.ReadFile(jsonfile)
	if err != nil {
		return nil, err
	}

	var tree Tree
	err = json.Unmarshal(data, &tree)
	if err != nil {
		return nil, err
	}

	tree.buildRoot()
	return &tree, nil
}

// 形成树结构
func (tree *Tree) buildRoot() {
	nodes := make(map[string]INode)
	for id, node := range tree.Nodes {
		inode := FactoryNode(&node)
		inode.init(&node, inode)
		nodes[id] = inode
	}

	for _, node := range nodes {
		cfg := node.getCfg()
		switch cfg.Category {
		case COMPOSITE:
			for _, cid := range cfg.Children {
				node.(IBranch).AddLeaf(nodes[cid])
			}

		case DECORATOR:
			node.(IBranch).AddLeaf(nodes[cfg.Child])
		}
	}

	tree.root = nodes[tree.Root]
	tree.Nodes = nil
}

// 遍历树
func (tree *Tree) Tick(userdata interface{}, store util.Map) bool {
	if tree.Properties.GetInt32("Run") == 0 {
		return false
	}
	tick := &Tick{
		UserData: userdata,
		Store:    store,
	}

	lastRunning := tick.GetRunningNodes()
	tick.Reset()

	tree.root.Execute(tick)

	running := tick.GetRunningNodes()
	for node := range lastRunning {
		if _, ok := running[node]; !ok {
			node.OnLeave(tick)
		}
	}
	return true
}
