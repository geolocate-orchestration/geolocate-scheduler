package random

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	informersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	"math/rand"
)

func GetNode(nodeLister informersv1.NodeLister, pod *v1.Pod) (*v1.Node, error) {
	klog.Infoln("getting cached nodes")
	nodes, err := nodeLister.List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, err
	} else if len(nodes) == 0 {
		errMessage := "no nodes are available"
		klog.Errorln(errMessage)
		return nil, errors.New(errMessage)
	}

	klog.Infof("will get 1 node from the %d available\n", len(nodes))
	node := nodes[rand.Intn(len(nodes))]
	klog.Infof("returning node %s\n", node.Name)

	return node, nil
}
