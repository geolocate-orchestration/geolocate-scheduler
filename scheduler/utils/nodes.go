package utils

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Nodes struct {
	ClientSet *kubernetes.Clientset
	Nodes []*Node
}

type Node struct {
	Name string
}

func (nodes *Nodes) StartNodeInformerHandler() {
	factory := informers.NewSharedInformerFactory(nodes.ClientSet, 0)
	nodeInformer := factory.Core().V1().Nodes()

	stopper := make(chan struct{})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := &Node{Name: obj.(*v1.Node).Name}
			nodes.Nodes = append(nodes.Nodes, node)

			klog.Infof("new node added to cache: %s\n", node.Name)
		},
		DeleteFunc: func(obj interface{}) {
			for i, v := range nodes.Nodes {
				if v.Name == obj.(*v1.Node).Name {
					nodes.Nodes = append(nodes.Nodes[:i], nodes.Nodes[i+1:]...)
				}
			}

			klog.Infof("node NOT deleted from cache: %s - TODO!\n", obj.(*v1.Node).Name)
		},
	})

	factory.Start(stopper)
}

func (nodes *Nodes) GetAllNodes() []*Node {
	return nodes.Nodes
}
