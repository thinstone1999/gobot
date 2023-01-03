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

// 逻辑或
type Priority struct {
	branch
}

// 叶子节点成功一个就返回
func (br *Priority) OnTick(tick *Tick) Status {
	for _, leaf := range br.leafs {
		status := leaf.execute(tick)
		if status == SUCCESS {
			return SUCCESS
		}
	}
	return ERROR
}

type MemPriority struct {
	branch
}

func (br *MemPriority) OnEnter(tick *Tick) {
	tick.GetIndexMap()[br.Id] = 0
}

// 叶子节点成功一个就返回
func (br *MemPriority) OnTick(tick *Tick) Status {
	indexMap := tick.GetIndexMap()
	for i := indexMap[br.Id]; i < len(br.leafs); i++ {
		status := br.leafs[i].execute(tick)
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
	branch
}

// 叶子节点失败一个就返回
func (br *Sequence) OnTick(tick *Tick) Status {
	for _, leaf := range br.leafs {
		status := leaf.execute(tick)
		if status != SUCCESS {
			return ERROR
		}
	}
	return SUCCESS
}

type MemSequence struct {
	branch
}

func (br *MemSequence) OnEnter(tick *Tick) {
	tick.GetIndexMap()[br.Id] = 0
}

// 叶子节点失败一个就返回
func (br *MemSequence) OnTick(tick *Tick) Status {
	indexMap := tick.GetIndexMap()
	for i := indexMap[br.Id]; i < len(br.leafs); i++ {
		status := br.leafs[i].execute(tick)
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
	branch
}

// 随机执行一个叶子节点
func (br *Rand) OnTick(tick *Tick) Status {
	n := len(br.leafs)
	if n == 0 {
		return SUCCESS
	}

	idx := rand.Intn(n)
	return br.leafs[idx].execute(tick)
}

// 执行顺序打乱，但每个节点都有机会执行一次
type MemRand struct {
	branch
}

// 随机执行一个叶子节点
func (br *MemRand) OnTick(tick *Tick) Status {
	n := len(br.leafs)
	if n == 0 {
		return SUCCESS
	}

	rand := tick.GetRandNode()
	toTick := rand[br.Id]
	if len(toTick) == 0 {
		toTick = map[int]struct{}{}
		rand[br.Id] = toTick

		for i := range br.leafs {
			toTick[i] = struct{}{}
		}
	}

	idx := 0
	for idx = range toTick {
		delete(toTick, idx)
		break
	}

	return br.leafs[idx].execute(tick)
}

func init() {
	Register(&Sequence{})
	Register(&MemSequence{})
	Register(&Priority{})
	Register(&MemPriority{})
	Register(&Rand{})
	Register(&MemRand{})
}
