package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	informersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"math/rand"
	"time"
)

const schedulerName = "random-scheduler"

func emitEvent(clientset *kubernetes.Clientset, pod *v1.Pod, node *v1.Node, err error) {
	reason, message, eventType := "", "", ""

	if err == nil {
		eventType = "Normal"
		reason = "Scheduled"
		message = fmt.Sprintf("Pod %s scheduled in node %s", pod.Name, node.Name)
	} else if node == nil {
		eventType = "Warning"
		reason = "ScheduleNodeError"
		message = fmt.Sprintf("Failed to get Node information to schedule Pod %s", pod.Name)
	} else {
		eventType = "Warning"
		reason = "ScheduleError"
		message = fmt.Sprintf("Failed to schedule Pod %s in node %s", pod.Name, node.Name)
	}

	timestamp := time.Now().UTC()
	_, _ = clientset.CoreV1().Events(pod.Namespace).Create(
		context.TODO(),
		&v1.Event{
			Count:          1,
			Message:        message,
			Reason:         reason,
			LastTimestamp:  metav1.NewTime(timestamp),
			FirstTimestamp: metav1.NewTime(timestamp),
			Type:           eventType,
			Source: v1.EventSource{
				Component: schedulerName,
			},
			InvolvedObject: v1.ObjectReference{
				Kind:      "Pod",
				Name:      pod.Name,
				Namespace: pod.Namespace,
				UID:       pod.UID,
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: pod.Name + "-",
			},
		},
		metav1.CreateOptions{},
	)
}

func bind(clientset *kubernetes.Clientset, nodeLister informersv1.NodeLister, pod *v1.Pod) {
	nodes, err := nodeLister.List(labels.Everything())

	if err != nil {
		fmt.Printf("ERROR: Could not list nodes")
		emitEvent(clientset, pod, nil, err)
		return
	}

	node := nodes[rand.Intn(len(nodes))]
	fmt.Printf("Assigned pod %s/%s to node %s\n", pod.Namespace, pod.Name, node.Name)

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

	emitEvent(clientset, pod, node, err)
}

func watch(clientset *kubernetes.Clientset, nodeLister informersv1.NodeLister) {
	watch, err := clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metav1.ListOptions {
			FieldSelector: fmt.Sprintf("spec.schedulerName=%s,spec.nodeName=", schedulerName),
		},
	)

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Watching for new pods to schedule...\n")

	for event := range watch.ResultChan() {
		if event.Type != "ADDED" {
			continue
		}
		pod := event.Object.(*v1.Pod)
		fmt.Printf("Found a pod to schedule: %s/%s\n", pod.Namespace, pod.Name)
		bind(clientset, nodeLister, pod)
	}
}

func getNodeLister(clientset *kubernetes.Clientset) informersv1.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	nodeInformer := factory.Core().V1().Nodes()

	stopper := make(chan struct{})
	defer close(stopper)

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			fmt.Printf("New Node Added to Store: %s\n", node.Name)
		},
	})

	factory.Start(stopper)

	return nodeInformer.Lister()
}

func main() {
	fmt.Printf("Starting %s...\n", schedulerName)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Creating clientset...\n")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodeLister := getNodeLister(clientset)
	watch(clientset, nodeLister)
}
