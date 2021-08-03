package k8smanager

import (
	"github.com/geolocate-orchestration/scheduler"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
)

// Run init the scheduler service
func Run(algorithmName string) error {
	ischeduler, err := scheduler.NewScheduler(algorithmName)
	if err != nil {
		klog.Errorln(err)
		os.Exit(1)
	}

	k := k8smanager{
		algorithm:  algorithmName,
		ischeduler: ischeduler,
	}

	klog.Infof("starting geolocate-scheduler - %s algorithm...\n", k.algorithm)

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorln(err)
		os.Exit(2)
	}

	klog.Infoln("creating clientset...")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorln(err)
		os.Exit(3)
	}

	k.clientset = clientset

	k.startNodeInformerHandler()
	k.watch()

	return nil
}
