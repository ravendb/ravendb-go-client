package ravendb

type CurrentIndexAndNode struct {
	currentIndex int
	currentNode  *ServerNode
}

func NewCurrentIndexAndNode(currentIndex int, currentNode *ServerNode) *CurrentIndexAndNode {
	return &CurrentIndexAndNode{
		currentIndex: currentIndex,
		currentNode:  currentNode,
	}
}
