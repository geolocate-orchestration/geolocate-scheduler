package scheduler

import (
	"aida-scheduler/scheduler/algorithms/geographicLocation"
	"aida-scheduler/scheduler/utils"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
)

const algorithmName = "geographic_location"

func bind(clientset *kubernetes.Clientset, geo geographicLocation.GeographicLocation, pod *v1.Pod) {
	node, err := geo.GetNode(pod)

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

func watch(clientset *kubernetes.Clientset, geo geographicLocation.GeographicLocation) {
	watch, err := clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metav1.ListOptions{
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
		bind(clientset, geo, pod)
	}
}

func Run() {
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

	nodes := utils.New(clientset)
	geo := geographicLocation.New(nodes)

	watch(clientset, geo)
}
