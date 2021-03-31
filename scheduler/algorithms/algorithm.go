package algorithms

import (
	"aida-scheduler/scheduler/nodes"
	v1 "k8s.io/api/core/v1"
)

type Algorithm interface {
	GetNode(pod *v1.Pod) (*nodes.Node, error)
}