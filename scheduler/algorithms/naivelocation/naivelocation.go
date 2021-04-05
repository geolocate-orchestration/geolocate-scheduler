package naivelocation

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"strings"
)

type naivelocation struct {
	nodes      nodes.INodes
	pod        *v1.Pod
	queryType  string // required or preferred
	cities     []string
	countries  []string
	continents []string
}

// New creates new naivelocation struct
func New(nodes nodes.INodes) algorithms.Algorithm {
	return &naivelocation{
		nodes:      nodes,
		pod:        nil,
		queryType:  "",
		cities:     make([]string, 0),
		countries:  make([]string, 0),
		continents: make([]string, 0),
	}
}

func (geo *naivelocation) GetName() string {
	return "naivelocation"
}

// GetNode select the best node matching the given constraints labels
// It returns error if there are no nodes available and if no node matches an existing 'requiredLocation' label
func (geo *naivelocation) GetNode(pod *v1.Pod) (*nodes.Node, error) {
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
		node, err = nodes.GetRandomFromList(geo.nodes.GetAllNodes())
	}

	return node, err
}

// Locations

func (geo *naivelocation) getNodeByLocation() (*nodes.Node, error) {
	locations := geo.pod.Labels["deployment.edge.aida.io/"+geo.queryType+"Location"]
	klog.Infoln(geo.queryType, "location:", locations)

	// fill location info from labels in the geo struct
	geo.parseLocations(locations)

	if node, err := geo.getRequestedLocation(); err == nil {
		return node, nil
	} else if geo.queryType == "required" {
		// if location is "required" but there are no matching nodes, throw error
		return nil, err
	}

	if node, err := geo.getSimilarToRequestedLocation(); err == nil {
		return node, nil
	}

	// when location is "preferred" and there are no matching nodes, return random node
	return nodes.GetRandomFromList(geo.nodes.GetAllNodes())
}

func (geo *naivelocation) getRequestedLocation() (*nodes.Node, error) {
	if node, err := geo.getByCity(); err == nil {
		return node, nil
	}

	if node, err := geo.getByCountry(); err == nil {
		return node, nil
	}

	if node, err := geo.getByContinent(); err == nil {
		return node, nil
	}

	return nil, errors.New("no nodes match given locations")
}

func (geo *naivelocation) getSimilarToRequestedLocation() (*nodes.Node, error) {
	if options := geo.nodes.FindNodesByCityCountry(geo.cities, nil); len(options) > 0 {
		return nodes.GetRandomFromMap(options)
	}

	if options := geo.nodes.FindNodesByCityContinent(geo.cities, nil); len(options) > 0 {
		return nodes.GetRandomFromMap(options)
	}

	if options := geo.nodes.FindNodesByCountryContinent(geo.countries, nil); len(options) > 0 {
		return nodes.GetRandomFromMap(options)
	}

	return nil, errors.New("no nodes match similar location to given locations")
}

// GetBy

func (geo *naivelocation) getByCity() (*nodes.Node, error) {
	options := geo.nodes.FindNodesByCity(geo.cities, nil)
	return nodes.GetRandomFromMap(options)
}

func (geo *naivelocation) getByCountry() (*nodes.Node, error) {
	options := geo.nodes.FindNodesByCountry(geo.countries, nil)
	return nodes.GetRandomFromMap(options)
}

func (geo *naivelocation) getByContinent() (*nodes.Node, error) {
	options := geo.nodes.FindNodesByContinent(geo.continents, nil)
	return nodes.GetRandomFromMap(options)
}

// Helpers

func (geo *naivelocation) getLocationLabelType() string {
	if geo.pod.Labels["deployment.edge.aida.io/requiredLocation"] != "" {
		return "required"
	}

	if geo.pod.Labels["deployment.edge.aida.io/preferredLocation"] != "" {
		return "preferred"
	}

	return ""
}

func (geo *naivelocation) parseLocations(locations string) {
	divisions := strings.Split(locations, "-")
	geo.cities = strings.Split(divisions[0], "_")
	geo.countries = strings.Split(divisions[1], "_")
	geo.continents = strings.Split(divisions[2], "_")
}
