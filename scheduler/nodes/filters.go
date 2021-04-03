package nodes

func nodeMatchesFilters(node *Node, filter *NodeFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Labels != nil && !nodeHasLabels(node, filter.Labels) {
		return false
	}

	if filter.CPU != 0 && !nodeHasAllocatableCPU(node, filter.CPU) {
		return false
	}

	if filter.Memory != 0 && !nodeHasAllocatableMemory(node, filter.Memory) {
		return false
	}

	return true
}

func nodeHasLabels(node *Node, labels []string) bool {
	for _, label := range labels {
		if _, ok := node.Labels[label]; !ok {
			return false
		}
	}

	return true
}

func nodeHasAllocatableCPU(node *Node, cpu int64) bool {
	return node.CPU >= cpu
}

func nodeHasAllocatableMemory(node *Node, memory int64) bool {
	return node.Memory >= memory
}
