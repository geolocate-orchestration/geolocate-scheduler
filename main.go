package scheduler

import (
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"math/rand"
	"time"
)

const schedulerName = "random-scheduler"

func emitEvent(clientset *kubernetes.Clientset, pod *v1.Pod, node *v1.Node, err error) {
	reason, message, eventType := "", "", ""

	if err != nil {
		eventType = "Normal"
		reason = "Scheduled"
		message = fmt.Sprintf("Pod %s scheduled in node %s", pod.Name, node.Name)
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

func randomNode(clientset *kubernetes.Clientset) *v1.Node {
	nodes, _ := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	return &nodes.Items[rand.Intn(len(nodes.Items))]
}

func bind(clientset *kubernetes.Clientset, pod *v1.Pod) {
	node :=randomNode(clientset)

	err := clientset.CoreV1().Pods(pod.Namespace).Bind(
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

func watch(clientset *kubernetes.Clientset) {
	watch, err := clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metav1.ListOptions {
			FieldSelector: fmt.Sprintf("spec.schedulerName=%s,spec.nodeName=", schedulerName),
		},
	)

	if err != nil {
		panic(err.Error())
	}

	for event := range watch.ResultChan() {
		if event.Type != "ADDED" {
			continue
		}
		pod := event.Object.(*v1.Pod)
		fmt.Println("found a pod to schedule:", pod.Namespace, "/", pod.Name)
		bind(clientset, pod)
	}
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watch(clientset)
}
