package k8smanager

import (
	"github.com/mv-orchestration/scheduler/nodes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (k *k8smanager) startNodeInformerHandler() {
	if k.clientset == nil {
		return
	}

	factory := informers.NewSharedInformerFactory(k.clientset, 0)
	nodeInformer := factory.Core().V1().Nodes()

	stopper := make(chan struct{})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    k.addHandler,
		UpdateFunc: k.updateHandler,
		DeleteFunc: k.deleteHandler,
	})

	factory.Start(stopper)
}

func (k *k8smanager) addHandler(obj interface{}) {
	objNode := obj.(*v1.Node)

	node := &nodes.Node{
		Name:   objNode.Name,
		Labels: objNode.Labels,
		CPU:    objNode.Status.Allocatable.Cpu().MilliValue(),
		Memory: objNode.Status.Allocatable.Memory().MilliValue(),
	}

	k.ischeduler.AddNode(node)
}

func (k *k8smanager) updateHandler(oldObj interface{}, newObj interface{}) {
	oldObjNode := oldObj.(*v1.Node)
	objObjNode := newObj.(*v1.Node)

	oldNode := &nodes.Node{
		Name:   oldObjNode.Name,
		Labels: oldObjNode.Labels,
		CPU:    oldObjNode.Status.Allocatable.Cpu().MilliValue(),
		Memory: oldObjNode.Status.Allocatable.Memory().MilliValue(),
	}

	newNode := &nodes.Node{
		Name:   objObjNode.Name,
		Labels: objObjNode.Labels,
		CPU:    objObjNode.Status.Allocatable.Cpu().MilliValue(),
		Memory: objObjNode.Status.Allocatable.Memory().MilliValue(),
	}

	k.ischeduler.UpdateNode(oldNode, newNode)
}

func (k *k8smanager) deleteHandler(obj interface{}) {
	objNode := obj.(*v1.Node)

	node := &nodes.Node{
		Name:   objNode.Name,
	}

	k.ischeduler.DeleteNode(node)
}
