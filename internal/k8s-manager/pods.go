package k8smanager

import (
	"context"
	"os"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/geolocate-orchestration/scheduler/algorithms"
)

func (k *k8smanager) watch() {
	watch, err := k.clientset.CoreV1().Pods("").Watch(
		context.TODO(),
		metaV1.ListOptions{
			FieldSelector: "spec.schedulerName=geolocate-scheduler,spec.nodeName=",
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
	cpu := int64(0)
	mem := int64(0)

	for _, v := range pod.Spec.Containers {
		cpu += v.Resources.Requests.Cpu().MilliValue()
		mem += v.Resources.Requests.Memory().MilliValue()
	}

	w := &algorithms.Workload{
		Name:   pod.Name,
		Labels: pod.Labels,
		CPU:    cpu,
		Memory: mem,
	}

	node, err := k.ischeduler.ScheduleWorkload(w)

	if err != nil {
		klog.Errorln(err)
		EmitEvent("geolocate-scheduler", k.clientset, pod, "", err)
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

	EmitEvent("geolocate-scheduler", k.clientset, pod, node.Name, err)
}
