package random

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type random struct {
	inodes nodes.INodes
}

// New creates new random struct
func New(inodes nodes.INodes) algorithms.Algorithm {
	return &random{
		inodes: inodes,
	}
}

func (r random) GetNode(*v1.Pod) (*nodes.Node, error) {
	klog.Infoln("getting cached nodes")
	return GetRandomNode(r.inodes)
}

// GetRandomNode returns a random
func GetRandomNode(inodes nodes.INodes) (*nodes.Node, error) {
	allNodes := inodes.GetAllNodes()

	if len(allNodes) == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	klog.Infof("will randomly get 1 node from the %d available\n", len(allNodes))
	return nodes.GetRandom(allNodes), nil
}
