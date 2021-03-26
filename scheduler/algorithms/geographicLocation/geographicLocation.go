package geographicLocation

import (
	"aida-scheduler/scheduler/utils"
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"math/rand"
	"strings"
)

type GeographicalLocation struct {
	Continents []string
	Countries []string
	Cities []string
}

func GetNode(nodes *utils.Nodes, pod *v1.Pod) (*utils.Node, error) {
	var node *utils.Node
	var err error

	if pod.Labels["deployment.edge.aida.io/requiredLocation"] != "" {
		node, err = getRequiredLocationNode(nodes, pod)
	} else if pod.Labels["deployment.edge.aida.io/preferredLocation"] != "" {
		node, err = getPreferredLocationNode(nodes, pod)
	} else {
		node, err = getRandomNode(nodes)
	}

	return node, err
}

func getRequiredLocationNode(nodes *utils.Nodes,  pod *v1.Pod) (*utils.Node, error) {
	locations := pod.Labels["deployment.edge.aida.io/requiredLocation"]

	klog.Infoln("requiredLocation:", locations)
	// geo := parseLocations(locations)

	return getRandomNode(nodes)
}

func getPreferredLocationNode(nodes *utils.Nodes,  pod *v1.Pod) (*utils.Node, error) {
	locations := pod.Labels["deployment.edge.aida.io/preferredLocation"]

	klog.Infoln("preferredLocation:", locations)
	// geo := parseLocations(locations)

	return getRandomNode(nodes)
}

func getRandomNode(nodes *utils.Nodes) (*utils.Node, error) {
	allNodes := nodes.GetAllNodes()

	if len(allNodes) == 0 {
		errMessage := "no nodes are available"
		return nil, errors.New(errMessage)
	}

	klog.Infof("will randomly get 1 node from the %d available\n", len(allNodes))
	return allNodes[rand.Intn(len(allNodes))], nil
}

func parseLocations(locations string) *GeographicalLocation {
	geo := &GeographicalLocation{}

	divisions := strings.Split(locations, "-")
	geo.Cities = strings.Split(divisions[0], "_")
	geo.Countries = strings.Split(divisions[1], "_")
	geo.Continents = strings.Split(divisions[2], "_")

	return geo
}
