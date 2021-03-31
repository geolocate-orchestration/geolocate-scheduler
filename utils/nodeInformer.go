package utils

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// StartNodeInformerHandler begins listening for cluster nodes changes
func StartNodeInformerHandler(
	clientSet *kubernetes.Clientset, addHandler func(obj interface{}),
	updateHandler func(oldObj interface{}, newObj interface{}), deleteHandler func(obj interface{}),
) {
	if clientSet == nil {
		return
	}

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes()

	stopper := make(chan struct{})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addHandler,
		UpdateFunc: updateHandler,
		DeleteFunc: deleteHandler,
	})

	factory.Start(stopper)
}
