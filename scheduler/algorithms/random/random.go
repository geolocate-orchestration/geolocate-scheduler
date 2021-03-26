package random

import (
	"aida-scheduler/scheduler/utils"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"math/rand"
)

func GetNode(nodes *utils.Nodes, pod *v1.Pod) (*utils.Node, error) {
	klog.Infoln("getting cached nodes")
	allNodes := nodes.GetAllNodes()

	if len(allNodes) == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	klog.Infof("will get 1 node from the %d available\n", len(allNodes))
	node := allNodes[rand.Intn(len(allNodes))]
	klog.Infof("returning node %s\n", node.Name)

	return node, nil
}
