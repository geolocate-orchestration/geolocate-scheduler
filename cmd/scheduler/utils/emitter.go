package utils

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

func EmitEvent(algorithmName string, clientset *kubernetes.Clientset, pod *v1.Pod, node *v1.Node, err error) {
	reason, message, eventType := getMessage(pod, node, err)
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
			Source:         eventSource(algorithmName),
			InvolvedObject: involvedObject(pod),
			ObjectMeta:     objectMeta(pod),
		},
		metav1.CreateOptions{},
	)
}

func involvedObject(pod *v1.Pod) v1.ObjectReference {
	return v1.ObjectReference{
		Kind:      "Pod",
		Name:      pod.Name,
		Namespace: pod.Namespace,
		UID:       pod.UID,
	}
}

func eventSource(algorithmName string) v1.EventSource {
	return v1.EventSource{
		Component: algorithmName,
	}
}

func objectMeta(pod *v1.Pod) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		GenerateName: pod.Name + "-",
	}
}

func getMessage(pod *v1.Pod, node *v1.Node, err error) (string, string, string) {
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

	return reason, message, eventType
}
