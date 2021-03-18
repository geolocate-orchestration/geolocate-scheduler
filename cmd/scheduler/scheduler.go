package main

import (
	"aida-scheduler/algorithms/location_sorted"
	"aida-scheduler/utils"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	informersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"os"
)

const algorithmName = "location_sorted"
var GetNode = location_sorted.GetNode

func bind(clientset *kubernetes.Clientset, nodeLister informersv1.NodeLister, pod *v1.Pod) {
	node, err := GetNode(nodeLister, pod)

	if err != nil {
		klog.Errorln(err)
		utils.EmitEvent(algorithmName, clientset, pod, nil, err)
		return
	}

	klog.Infof("assigned pod %s/%s to node %s\n", pod.Namespace, pod.Name, node.Name)

	err = clientset.CoreV1().Pods(pod.Namespace).Bind(
		context.TODO(),
		&v1.Binding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			Target: v1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Node",
				Name:       node.Name,
			},
		},
		metav1.CreateOptions{},
	)

	utils.EmitEvent(algorithmName, clientset, pod, node, err)
}

func watch(clientset *kubernetes.Clientset, nodeLister informersv1.NodeLister) {
	watch, err := clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metav1.ListOptions {
			FieldSelector: fmt.Sprintf("spec.schedulerName=aida-scheduler,spec.nodeName="),
		},
	)

	if err != nil {
		klog.Errorln(err)
		os.Exit(3)
	}

	klog.Infof("watching for new pods to schedule...\n")

	for event := range watch.ResultChan() {
		if event.Type != "ADDED" {
			continue
		}
		pod := event.Object.(*v1.Pod)
		klog.Infof("found a pod to schedule: %s/%s\n", pod.Namespace, pod.Name)
		bind(clientset, nodeLister, pod)
	}
}

func getNodeLister(clientset *kubernetes.Clientset) informersv1.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	nodeInformer := factory.Core().V1().Nodes()

	stopper := make(chan struct{})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			klog.Infof("new node added to cache: %s\n", node.Name)
		},
		DeleteFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			klog.Infof("node deleted from cache: %s\n", node.Name)
		},
	})

	factory.Start(stopper)

	return nodeInformer.Lister()
}

func main() {
	klog.Infof("starting %s aida-scheduler...\n", algorithmName)

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorln(err)
		os.Exit(1)
	}

	klog.Infoln("creating clientset...")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorln(err)
		os.Exit(2)
	}

	nodeLister := getNodeLister(clientset)
	watch(clientset, nodeLister)
}
