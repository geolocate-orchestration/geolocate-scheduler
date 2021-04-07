package k8smanager

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// EmitEvent logs the binding of a pod to a node to kubernetes cluster
func EmitEvent(name string, clientset *kubernetes.Clientset, pod *v1.Pod, nodeID string, err error) {
	reason, message, eventType := getMessage(pod, nodeID, err)
	timestamp := time.Now().UTC()

	_, _ = clientset.CoreV1().Events(pod.Namespace).Create(
		context.TODO(),
		&v1.Event{
			Count:          1,
			Message:        message,
			Reason:         reason,
			LastTimestamp:  metaV1.NewTime(timestamp),
			FirstTimestamp: metaV1.NewTime(timestamp),
			Type:           eventType,
			Source:         eventSource(name),
			InvolvedObject: involvedObject(pod),
			ObjectMeta:     objectMeta(pod),
		},
		metaV1.CreateOptions{},
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

func objectMeta(pod *v1.Pod) metaV1.ObjectMeta {
	return metaV1.ObjectMeta{
		GenerateName: pod.Name + "-",
	}
}

func getMessage(pod *v1.Pod, nodeID string, err error) (string, string, string) {
	reason, message, eventType := "", "", ""

	if err == nil {
		eventType = "Normal"
		reason = "Scheduled"
		message = fmt.Sprintf("Pod %s scheduled in node %s", pod.Name, nodeID)
	} else if nodeID == "" {
		eventType = "Warning"
		reason = "ScheduleNodeError"
		message = fmt.Sprintf("Failed to get Node information to schedule Pod %s", pod.Name)
	} else {
		eventType = "Warning"
		reason = "ScheduleError"
		message = fmt.Sprintf("Failed to schedule Pod %s in node %s", pod.Name, nodeID)
	}

	return reason, message, eventType
}
