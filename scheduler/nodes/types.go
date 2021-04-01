package nodes

import (
	"github.com/aida-dos/gountries"
	"k8s.io/client-go/kubernetes"
)

// INodes exports all node controller public methods
type INodes interface {
	GetAllNodes() []*Node
	CountNodes() int
	FindAnyNodeByCity(cities []string) (*Node, error)
	FindAnyNodeByCityCountry(cities []string) (*Node, error)
	FindAnyNodeByCityContinent(cities []string) (*Node, error)
	FindAnyNodeByCountry(countries []string) (*Node, error)
	FindAnyNodeByCountryContinent(countries []string) (*Node, error)
	FindAnyNodeByContinent(continents []string) (*Node, error)
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
	Name string
}

const (
	cityLabel      = "node.edge.aida.io/city"
	countryLabel   = "node.edge.aida.io/country"
	continentLabel = "node.edge.aida.io/continent"
)
