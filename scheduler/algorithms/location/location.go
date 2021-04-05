package location

import (
	"aida-scheduler/scheduler/algorithms"
	"aida-scheduler/scheduler/nodes"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"strings"
)

type location struct {
	nodes      nodes.INodes
	pod        *v1.Pod
	queryType  string // required or preferred
	cities     []string
	countries  []string
	continents []string
	filter     *nodes.NodeFilter
}

// New creates new location struct
func New(nodes nodes.INodes) algorithms.Algorithm {
	return &location{
		nodes:      nodes,
		pod:        nil,
		queryType:  "",
		cities:     make([]string, 0),
		countries:  make([]string, 0),
		continents: make([]string, 0),
	}
}

func (geo *location) GetName() string {
	return "location"
}

// GetNode select the best node matching the given constraints labels
// It returns error if there are no nodes available and if no node matches an existing 'requiredLocation' label
func (geo *location) GetNode(pod *v1.Pod) (*nodes.Node, error) {
	var node *nodes.Node
	var err error

	if geo.nodes.CountNodes() == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	geo.pod = pod

	geo.buildFilters()

	if queryType := geo.getLocationLabelType(); queryType != "" {
		geo.queryType = queryType
		node, err = geo.getNodeByLocation()
	} else {
		// Node location labels were set so returning a random node
		node, err = nodes.GetRandomFromList(geo.nodes.GetNodes(geo.filter))
	}

	return node, err
}

// Locations

func (geo *location) buildFilters() {
	cpu, memory := geo.getResourceSum()

	geo.filter = &nodes.NodeFilter{
		Labels: nil,
		CPU:    cpu,
		Memory: memory,
	}
}

func (geo *location) getResourceSum() (int64, int64) {
	cpu := int64(0)
	memory := int64(0)

	for _, container := range geo.pod.Spec.Containers {
		cpu += container.Resources.Requests.Cpu().MilliValue()
		memory += container.Resources.Requests.Memory().MilliValue()
	}

	return cpu, memory
}

func (geo *location) getNodeByLocation() (*nodes.Node, error) {
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
	return nodes.GetRandomFromList(geo.nodes.GetNodes(geo.filter))
}

func (geo *location) getRequestedLocation() (*nodes.Node, error) {
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

func (geo *location) getSimilarToRequestedLocation() (*nodes.Node, error) {
	if options := geo.nodes.FindNodesByCityCountry(geo.cities, geo.filter); len(options) > 0 {
		return nodes.GetRandomFromMap(options)
	}

	if options := geo.nodes.FindNodesByCityContinent(geo.cities, geo.filter); len(options) > 0 {
		return nodes.GetRandomFromMap(options)
	}

	if options := geo.nodes.FindNodesByCountryContinent(geo.countries, geo.filter); len(options) > 0 {
		return nodes.GetRandomFromMap(options)
	}

	return nil, errors.New("no nodes match similar location to given locations")
}

// GetBy

func (geo *location) getByCity() (*nodes.Node, error) {
	options := geo.nodes.FindNodesByCity(geo.cities, geo.filter)
	return nodes.GetRandomFromMap(options)
}

func (geo *location) getByCountry() (*nodes.Node, error) {
	options := geo.nodes.FindNodesByCountry(geo.countries, geo.filter)
	return nodes.GetRandomFromMap(options)
}

func (geo *location) getByContinent() (*nodes.Node, error) {
	options := geo.nodes.FindNodesByContinent(geo.continents, geo.filter)
	return nodes.GetRandomFromMap(options)
}

// Helpers

func (geo *location) getLocationLabelType() string {
	if geo.pod.Labels["deployment.edge.aida.io/requiredLocation"] != "" {
		return "required"
	}

	if geo.pod.Labels["deployment.edge.aida.io/preferredLocation"] != "" {
		return "preferred"
	}

	return ""
}

func (geo *location) parseLocations(locations string) {
	divisions := strings.Split(locations, "-")
	geo.cities = strings.Split(divisions[0], "_")
	geo.countries = strings.Split(divisions[1], "_")
	geo.continents = strings.Split(divisions[2], "_")
}
