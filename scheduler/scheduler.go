package scheduler

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/algorithms/geographiclocation"
	"aida-scheduler/scheduler/algorithms/random"
	"aida-scheduler/scheduler/nodes"
	"aida-scheduler/utils"
	"context"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
)

const algorithmName = "geographiclocation"

// const algorithmName = "random"

func bind(clientset *kubernetes.Clientset, algorithm algorithms.Algorithm, pod *v1.Pod) {
	node, err := algorithm.GetNode(pod)

	if err != nil {
		klog.Errorln(err)
		utils.EmitEvent(algorithmName, clientset, pod, "", err)
		return
	}

	klog.Infof("assigned pod %s/%s to node %s\n", pod.Namespace, pod.Name, node.Name)

	err = clientset.CoreV1().Pods(pod.Namespace).Bind(
		context.TODO(),
		&v1.Binding{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			Target: v1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Node",
				Name:       node.Name,
			},
		},
		metaV1.CreateOptions{},
	)

	utils.EmitEvent(algorithmName, clientset, pod, node.Name, err)
}

func watch(clientset *kubernetes.Clientset, algorithm algorithms.Algorithm) {
	watch, err := clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metaV1.ListOptions{
			FieldSelector: "spec.schedulerName=aida-scheduler,spec.nodeName=",
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
		bind(clientset, algorithm, pod)
	}
}

// Run init the scheduler service
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

	nodesStruct := nodes.New(clientset)
	var algorithm algorithms.Algorithm

	switch algorithmName {
	case "random":
		algorithm = random.New(nodesStruct)
	case "geographiclocation":
		algorithm = geographiclocation.New(nodesStruct)
	default:
		algorithm = random.New(nodesStruct)
	}

	watch(clientset, algorithm)
}
