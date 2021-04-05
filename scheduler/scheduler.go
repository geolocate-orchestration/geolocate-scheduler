package scheduler

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/algorithms/naivelocation"
	"aida-scheduler/scheduler/algorithms/location"
	"aida-scheduler/scheduler/algorithms/random"
	"aida-scheduler/scheduler/nodes"
	"aida-scheduler/utils"
	"context"
	"errors"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
)

var availableAlgorithms = []string{"location", "naivelocation", "random"}

func bind(clientset *kubernetes.Clientset, algorithm algorithms.Algorithm, pod *v1.Pod) {
	node, err := algorithm.GetNode(pod)

	if err != nil {
		klog.Errorln(err)
		utils.EmitEvent(algorithm.GetName(), clientset, pod, "", err)
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

	utils.EmitEvent(algorithm.GetName(), clientset, pod, node.Name, err)
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

func algorithmExists(algorithmName string) bool {
	for _, algorithm := range availableAlgorithms {
		if algorithm == algorithmName {
			return true
		}
	}

	return false
}

func initAlgorithm(algorithmName string, nodesStruct nodes.INodes) algorithms.Algorithm {
	var algorithm algorithms.Algorithm

	switch algorithmName {
	case "random":
		algorithm = random.New(nodesStruct)
	case "naivelocation":
		algorithm = naivelocation.New(nodesStruct)
	case "location":
		algorithm = location.New(nodesStruct)
	default:
		algorithm = random.New(nodesStruct)
	}

	return algorithm
}

// Run init the scheduler service
func Run(algorithmName string) error {
	if !algorithmExists(algorithmName) {
		return errors.New("selected algorithm does not exist")
	}

	klog.Infof("starting aida-scheduler - %s algorithm...\n", algorithmName)

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
	algorithm := initAlgorithm(algorithmName, nodesStruct)

	watch(clientset, algorithm)

	return nil
}
