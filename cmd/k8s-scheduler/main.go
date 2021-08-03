package main

import (
	k8smanager "github.com/geolocate-orchestration/geolocate-scheduler/internal/k8s-manager"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	algorithm := os.Getenv("ALGORITHM")

	if algorithm == "" {
		algorithm = "random"
	}

	err := k8smanager.Run(algorithm)
	klog.Errorln(err)
}
