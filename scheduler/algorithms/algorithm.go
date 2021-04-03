package algorithms

import (
	"aida-scheduler/scheduler/nodes"
	v1 "k8s.io/api/core/v1"
)

// Algorithm interface that exposes GetNode method
type Algorithm interface {
	GetName() string
	GetNode(pod *v1.Pod) (*nodes.Node, error)
}
