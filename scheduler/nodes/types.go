package nodes

import (
	"github.com/aida-dos/gountries"
	"k8s.io/client-go/kubernetes"
)

// INodes exports all node controller public methods
type INodes interface {
	CountNodes() int
	GetAllNodes() []*Node
	GetNodes(filter *NodeFilter) []*Node
	FindNodesByCity(cities []string, filter *NodeFilter) map[string][]*Node
	FindNodesByCityCountry(cities []string, filter *NodeFilter) map[string][]*Node
	FindNodesByCityContinent(cities []string, filter *NodeFilter) map[string][]*Node
	FindNodesByCountry(countries []string, filter *NodeFilter) map[string][]*Node
	FindNodesByCountryContinent(countries []string, filter *NodeFilter) map[string][]*Node
	FindNodesByContinent(continents []string, filter *NodeFilter) map[string][]*Node
}

// Nodes controls in-cache nodes
type Nodes struct {
	ClientSet *kubernetes.Clientset

	Query          *gountries.Query
	ContinentsList gountries.Continents

	Nodes      []*Node
	Cities     map[string][]*Node
	Countries  map[string][]*Node
	Continents map[string][]*Node
}

// Node represents a cluster Node
type Node struct {
	Name   string
	Labels map[string]string
	CPU    int64
	Memory int64
}

// NodeFilter states the params which nodes must match to be returned
type NodeFilter struct {
	Labels []string
	CPU    int64
	Memory int64
}

const (
	cityLabel      = "node.edge.aida.io/city"
	countryLabel   = "node.edge.aida.io/country"
	continentLabel = "node.edge.aida.io/continent"
)
