package k8smanager

import (
	"context"
	"github.com/geolocate-orchestration/scheduler/nodes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
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
	newObjNode := newObj.(*v1.Node)

	oldNode := &nodes.Node{
		Name: oldObjNode.Name,
	}

	cpu, memory := k.getNodeAvailableResources(newObjNode)

	newNode := &nodes.Node{
		Name:   newObjNode.Name,
		Labels: newObjNode.Labels,
		CPU:    cpu,
		Memory: memory,
	}

	k.ischeduler.UpdateNode(oldNode, newNode)
}

func (k *k8smanager) getNodeAvailableResources(node *v1.Node) (int64, int64) {
	allocCPU := node.Status.Allocatable.Cpu().MilliValue()
	allocMemory := node.Status.Allocatable.Memory().MilliValue()

	cpu, memory, err := k.getNodeResourcesLimits(node)
	if err == nil {
		return allocCPU - cpu, allocMemory - memory
	}

	klog.Errorln(err)
	return 0, 0
}

func (k *k8smanager) getNodeResourcesLimits(node *v1.Node) (int64, int64, error) {
	pods, err := k.clientset.CoreV1().Pods("").List(
		context.TODO(),
		metav1.ListOptions{FieldSelector: "spec.nodeName=" + node.Name},
	)

	if err != nil {
		return 0, 0, err
	}

	cpu, memory := int64(0), int64(0)

	for _, item := range pods.Items {
		for _, container := range item.Spec.Containers {
			c, m := getEstimatedContainerResourceUsage(&container)
			cpu += c
			memory += m
		}
	}

	return cpu, memory, nil
}

func getEstimatedContainerResourceUsage(container *v1.Container) (int64, int64) {
	cpu := container.Resources.Limits.Cpu().MilliValue()
	memory := container.Resources.Limits.Memory().MilliValue()

	if cpu == 0 {
		cpu = container.Resources.Requests.Cpu().MilliValue()
	}

	if memory == 0 {
		memory = container.Resources.Requests.Memory().MilliValue()
	}

	return cpu, memory
}

func (k *k8smanager) deleteHandler(obj interface{}) {
	objNode := obj.(*v1.Node)

	node := &nodes.Node{
		Name: objNode.Name,
	}

	k.ischeduler.DeleteNode(node)
}
