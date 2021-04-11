package k8smanager

import (
	"github.com/mv-orchestration/scheduler"
	"k8s.io/client-go/kubernetes"
)

type k8smanager struct {
	clientset  *kubernetes.Clientset
	algorithm  string
	ischeduler scheduler.IScheduler
}
