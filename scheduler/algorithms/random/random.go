package random

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"math/rand"
)

type random struct {
	nodes nodes.INodes
}

// New creates new random struct
func New(nodes nodes.INodes) algorithms.Algorithm {
	return &random{
		nodes: nodes,
	}
}

func (r random) GetNode(*v1.Pod) (*nodes.Node, error) {
	klog.Infoln("getting cached nodes")
	allNodes := r.nodes.GetAllNodes()
	return getRandomNode(allNodes)
}

func getRandomNode(allNodes []*nodes.Node) (*nodes.Node, error) {
	if len(allNodes) == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	klog.Infof("will get 1 node from the %d available\n", len(allNodes))
	node := allNodes[rand.Intn(len(allNodes))]
	klog.Infof("returning node %s\n", node.Name)

	return node, nil
}
