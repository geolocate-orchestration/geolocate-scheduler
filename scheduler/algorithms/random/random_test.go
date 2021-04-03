package random

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"github.com/aida-dos/gountries"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestRandom() algorithms.Algorithm {
	return New(nodes.New(nil))
}

func newTestRandomWithNode() *nodes.Nodes {
	return &nodes.Nodes{
		ClientSet: nil,

		Query:          gountries.New(),
		ContinentsList: gountries.NewContinents(),

		Nodes:      []*nodes.Node{{Name: "Node0"}},
		Cities:     make(map[string][]*nodes.Node),
		Countries:  make(map[string][]*nodes.Node),
		Continents: make(map[string][]*nodes.Node),
	}
}

func TestGetNodeEmpty(t *testing.T) {
	randomStruct := newTestRandom()
	_, err := randomStruct.GetNode(nil)
	assert.Error(t, err)
}

func TestGetNode(t *testing.T) {
	node, _ := getRandomNode(newTestRandomWithNode())
	assert.Equal(t, "Node0", node.Name)
}
