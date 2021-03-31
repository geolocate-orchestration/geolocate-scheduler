package geographiclocation

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"math/rand"
	"strings"
)

type geographiclocation struct {
	nodes      nodes.INodes
	pod        *v1.Pod
	queryType  string // required or preferred
	cities     []string
	countries  []string
	continents []string
}

// New creates new geographiclocation struct
func New(nodes nodes.INodes) algorithms.Algorithm {
	return &geographiclocation{
		nodes:      nodes,
		pod:        nil,
		queryType:  "",
		cities:     make([]string, 0),
		countries:  make([]string, 0),
		continents: make([]string, 0),
	}
}

// GetNode select the best node matching the given constraints labels
// It returns error if there are no nodes available and if no node matches an existing 'requiredLocation' label
func (geo *geographiclocation) GetNode(pod *v1.Pod) (*nodes.Node, error) {
	var node *nodes.Node
	var err error

	if geo.nodes.CountNodes() == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	geo.pod = pod
	if queryType := geo.getLocationLabelType(); queryType != "" {
		geo.queryType = queryType
		node, err = geo.getNodeByLocation()
	} else {
		// Node location labels were set so returning a random node
		node, err = getRandomNode(geo.nodes)
	}

	return node, err
}

// Locations

func (geo *geographiclocation) getNodeByLocation() (*nodes.Node, error) {
	locations := geo.pod.Labels["deployment.edge.aida.io/"+geo.queryType+"Location"]
	klog.Infoln(geo.queryType, "location:", locations)

	// fill location info from labels in the geo struct
	geo.parseLocations(locations)

	if node, err := geo.getByCity(); err == nil {
		return node, nil
	}

	if node, err := geo.getByCountry(); err == nil {
		return node, nil
	}

	if node, err := geo.getByContinent(); err == nil {
		return node, nil
	}

	if geo.queryType == "required" {
		// if location is "required" but there are no matching nodes, throw error
		return nil, errors.New("no nodes match given locations")
	}

	// when location is "preferred" and there are no matching nodes, return random node
	return getRandomNode(geo.nodes)
}

// GetBy

func (geo *geographiclocation) getByCity() (*nodes.Node, error) {
	if node, err := geo.nodes.FindAnyNodeByCity(geo.cities); err == nil {
		return node, nil
	} else if geo.queryType == "required" {
		// because location is "required" will not search nodes matching selected cities countries not continents
		return nil, errors.New("no nodes matched selected cities")
	}

	// because location is "preferred" will search nodes in the selected cities countries
	if node, err := geo.nodes.FindAnyNodeByCityCountry(geo.cities); err == nil {
		return node, nil
	}

	// because location is "preferred" will search nodes in the selected cities continents
	if node, err := geo.nodes.FindAnyNodeByCityContinent(geo.cities); err == nil {
		return node, nil
	}

	return nil, errors.New("no nodes matched selected cities or their countries/continents")
}

func (geo *geographiclocation) getByCountry() (*nodes.Node, error) {
	if node, err := geo.nodes.FindAnyNodeByCountry(geo.countries); err == nil {
		return node, nil
	} else if geo.queryType == "required" {
		// because location is "required" will not search nodes matching selected countries continents
		return nil, errors.New("no nodes matched selected countries")
	}

	// because location is "preferred" will search nodes in the selected countries continents
	if node, err := geo.nodes.FindAnyNodeByCountryContinent(geo.countries); err == nil {
		return node, nil
	}

	return nil, errors.New("no nodes matched selected countries or their continents")
}

func (geo *geographiclocation) getByContinent() (*nodes.Node, error) {
	if node, err := geo.nodes.FindAnyNodeByContinent(geo.continents); err == nil {
		return node, nil
	}

	return nil, errors.New("no nodes matched selected continents")
}

// Helpers

func (geo *geographiclocation) getLocationLabelType() string {
	if geo.pod.Labels["deployment.edge.aida.io/requiredLocation"] != "" {
		return "required"
	}

	if geo.pod.Labels["deployment.edge.aida.io/preferredLocation"] != "" {
		return "preferred"
	}

	return ""
}

func getRandomNode(nodes nodes.INodes) (*nodes.Node, error) {
	allNodes := nodes.GetAllNodes()

	if len(allNodes) == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	klog.Infof("will randomly get 1 node from the %d available\n", len(allNodes))
	return allNodes[rand.Intn(len(allNodes))], nil
}

func (geo *geographiclocation) parseLocations(locations string) {
	divisions := strings.Split(locations, "-")
	geo.cities = strings.Split(divisions[0], "_")
	geo.countries = strings.Split(divisions[1], "_")
	geo.continents = strings.Split(divisions[2], "_")
}
