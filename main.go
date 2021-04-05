package main

import (
	"aida-scheduler/scheduler"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	algorithm := os.Getenv("ALGORITHM")

	if algorithm == "" {
		algorithm = "location"
	}

	err := scheduler.Run(algorithm)
	klog.Errorln(err)
}
