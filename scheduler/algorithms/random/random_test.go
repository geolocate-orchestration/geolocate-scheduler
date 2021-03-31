package random

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestRandom() algorithms.Algorithm {
	return New(nodes.New(nil))
}

func newTestNode(name string) *nodes.Node {
	return &nodes.Node{
		Name: name,
	}
}

func TestGetNodeEmpty(t *testing.T) {
	randomStruct := newTestRandom()
	_, err := randomStruct.GetNode(nil)
	assert.Error(t, err)
}

func TestGetNode(t *testing.T) {
	allNodes := []*nodes.Node{
		newTestNode("Node0"),
	}

	node, _ := getRandomNode(allNodes)
	assert.Equal(t, "Node0", node.Name)
}
