package k8smanager

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
)

func (k *k8smanager) watch() {
	watch, err := k.clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metaV1.ListOptions{
			FieldSelector: "spec.schedulerName=k8s-scheduler,spec.nodeName=",
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
		k.bind(pod)
	}
}

func (k *k8smanager) bind(pod *v1.Pod) {
	node, err := k.ischeduler.ScheduleWorkload(nil)

	if err != nil {
		klog.Errorln(err)
		EmitEvent("k8s-scheduler", k.clientset, pod, "", err)
		return
	}

	klog.Infof("assigned pod %s/%s to node %s\n", pod.Namespace, pod.Name, node.Name)

	err = k.clientset.CoreV1().Pods(pod.Namespace).Bind(
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

	EmitEvent("k8s-scheduler", k.clientset, pod, node.Name, err)
}
