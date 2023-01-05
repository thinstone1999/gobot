package btree

import "math/rand"

// 分支节点，叶子节点的容器
type branch struct {
	baseNode
	leafs []INode
}

var _ IBranch = (*branch)(nil)

type IBranch interface {
	INode
	AddLeaf(INode)
}

func (br *branch) AddLeaf(node INode) {
	br.leafs = append(br.leafs, node)
}

// 可包含多个子节点
type Composite struct {
	branch
}

func (node *Composite) GetChild(i int) INode {
	return node.branch.leafs[i]
}

func (node *Composite) Len() int {
	return len(node.branch.leafs)
}

// 只能包含一个子节点
type Decorator struct {
	branch
}

func (node *Decorator) GetChild() INode {
	return node.branch.leafs[0]
}

// 逻辑或
type Priority struct {
	Composite
}

// 叶子节点成功一个就返回
func (br *Priority) OnTick(tick *Tick) Status {
	for i := 0; i < br.Len(); i++ {
		child := br.GetChild(i)
		status := child.Execute(tick)
		if status == SUCCESS {
			return SUCCESS
		}
	}
	return ERROR
}

type MemPriority struct {
	Composite
}

func (br *MemPriority) OnEnter(tick *Tick) {
	tick.GetIndexMap()[br.Id] = 0
}

// 叶子节点成功一个就返回
func (br *MemPriority) OnTick(tick *Tick) Status {
	indexMap := tick.GetIndexMap()
	for i := indexMap[br.Id]; i < br.Len(); i++ {
		status := br.GetChild(i).Execute(tick)
		if status == SUCCESS {
			return status
		}

		if status == RUNNING {
			indexMap[br.Id] = i
			return status
		}
	}
	return ERROR
}

// 逻辑与
type Sequence struct {
	Composite
}

// 叶子节点失败一个就返回
func (br *Sequence) OnTick(tick *Tick) Status {
	for i := 0; i < br.Len(); i++ {
		child := br.GetChild(i)
		status := child.Execute(tick)
		if status == ERROR {
			return ERROR
		}
	}
	return SUCCESS
}

type MemSequence struct {
	Composite
}

func (br *MemSequence) OnEnter(tick *Tick) {
	tick.GetIndexMap()[br.Id] = 0
}

// 叶子节点失败一个就返回
func (br *MemSequence) OnTick(tick *Tick) Status {
	indexMap := tick.GetIndexMap()
	for i := indexMap[br.Id]; i < br.Len(); i++ {
		status := br.GetChild(i).Execute(tick)
		if status == ERROR {
			return status
		}

		if status == RUNNING {
			indexMap[br.Id] = i
			return status
		}
	}
	return SUCCESS
}

// 随机节点
type Rand struct {
	Composite
}

// 随机执行一个叶子节点
func (br *Rand) OnTick(tick *Tick) Status {
	n := br.Len()
	if n == 0 {
		return SUCCESS
	}

	idx := rand.Intn(n)
	return br.GetChild(idx).Execute(tick)
}

// 执行顺序打乱，但每个节点都有机会执行一次
type MemRand struct {
	Composite
}

// 随机执行一个叶子节点
func (br *MemRand) OnTick(tick *Tick) Status {
	n := br.Len()
	if n == 0 {
		return SUCCESS
	}

	rand := tick.GetRandNode()
	toTick := rand[br.Id]
	if len(toTick) == 0 {
		toTick = map[int]struct{}{}
		rand[br.Id] = toTick

		for i := 0; i < n; i++ {
			toTick[i] = struct{}{}
		}
	}

	idx := 0
	for idx = range toTick {
		delete(toTick, idx)
		break
	}

	return br.GetChild(idx).Execute(tick)
}

func init() {
	Register(&Sequence{})
	Register(&MemSequence{})
	Register(&Priority{})
	Register(&MemPriority{})
	Register(&Rand{})
	Register(&MemRand{})
}
