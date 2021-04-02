package nodes

import (
	"aida-scheduler/utils"
	"errors"
	"fmt"
	"github.com/aida-dos/gountries"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// New create a new Nodes struct
func New(clientSet *kubernetes.Clientset) INodes {
	nodes := Nodes{
		ClientSet: clientSet,

		Query:          gountries.New(),
		ContinentsList: gountries.NewContinents(),

		Nodes:      make([]*Node, 0),
		Cities:     make(map[string][]*Node),
		Countries:  make(map[string][]*Node),
		Continents: make(map[string][]*Node),
	}
	utils.StartNodeInformerHandler(clientSet, nodes.addHandler, nodes.updateHandler, nodes.deleteHandler)
	return &nodes
}

func (nodes *Nodes) addHandler(obj interface{}) {
	nodes.addNode(obj.(*v1.Node))
}

func (nodes *Nodes) updateHandler(oldObj interface{}, newObj interface{}) {
	oldNode := oldObj.(*v1.Node)
	newNode := newObj.(*v1.Node)

	_, oldHasEdgeLabel := oldNode.Labels["node-role.kubernetes.io/edge"]
	_, newHasEdgeLabel := newNode.Labels["node-role.kubernetes.io/edge"]

	if !oldHasEdgeLabel && newHasEdgeLabel {
		// If node wasn't an edge node but now it is, create it in cache
		nodes.addNode(newNode)
	} else if oldHasEdgeLabel && newHasEdgeLabel {
		// If the node is an edge node and has significant update it in cache
		nodes.updateNode(oldNode, newNode)
	} else if oldHasEdgeLabel && !newHasEdgeLabel {
		// If node was an edge node but now it isn't, remove it from cache
		nodes.deleteNode(oldNode)
	}
}

func (nodes *Nodes) deleteHandler(obj interface{}) {
	nodes.deleteNode(obj.(*v1.Node))
}

// Private

func (nodes *Nodes) addNode(objNode *v1.Node) {
	if _, ok := objNode.Labels["node-role.kubernetes.io/edge"]; !ok {
		// Don't add new node if it doesn't have the edge role
		return
	}

	node := &Node{
		Name:   objNode.Name,
		Labels: objNode.Labels,
		CPU:    objNode.Status.Allocatable.Cpu().MilliValue(),
		Memory: objNode.Status.Allocatable.Memory().MilliValue(),
	}

	nodes.Nodes = append(nodes.Nodes, node)
	nodes.addToCities(node)
	nodes.addToCountries(node)
	nodes.addToContinents(node)

	klog.Infof("new node added to cache: %s\n", node.Name)
}

func (nodes *Nodes) addToCities(node *Node) {
	cityValue := node.Labels[cityLabel]

	if cityValue != "" {
		if city, err := nodes.Query.FindSubdivisionByName(cityValue); err == nil {
			cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
			nodes.Cities[cityCode] = append(nodes.Cities[cityCode], node)
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *Nodes) addToCountries(node *Node) {
	countryValue := node.Labels[countryLabel]

	if countryValue != "" {
		if country, err := nodes.findCountry(countryValue); err == nil {
			nodes.Countries[country.Alpha2] = append(nodes.Countries[country.Alpha2], node)
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *Nodes) addToContinents(node *Node) {
	continentValue := node.Labels[continentLabel]

	if continentValue != "" {
		if continent, err := nodes.ContinentsList.FindContinent(continentValue); err == nil {
			nodes.Continents[continent.Code] = append(nodes.Continents[continent.Code], node)
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *Nodes) updateNode(oldNode *v1.Node, newNode *v1.Node) {
	if nodeHasSignificantChanges(oldNode, newNode) {
		klog.Infof("node will be replaced in cache: %s\n", oldNode.Name)
		nodes.deleteNode(oldNode)
		nodes.addNode(newNode)
	} else {
		klog.Infof("node will be updated in cache: %s\n", oldNode.Name)
		node, _ := nodes.findNodeByName(newNode.Name)
		node.Labels = newNode.Labels
		node.CPU = newNode.Status.Allocatable.Cpu().MilliValue()
		node.Memory = newNode.Status.Allocatable.Memory().MilliValue()
	}
}

func (nodes *Nodes) findNodeByName(name string) (*Node, error) {
	for _, node := range nodes.Nodes {
		if node.Name == name {
			return node, nil
		}
	}

	return nil, errors.New("node with given name not found")
}

func (nodes *Nodes) deleteNode(objNode *v1.Node) {
	if _, ok := objNode.Labels["node-role.kubernetes.io/edge"]; !ok {
		// Don't try to remove the node if it doesn't have the edge role
		return
	}

	nodes.removeNodeFromNodes(objNode)
	nodes.removeNodeFromCities(objNode)
	nodes.removeNodeFromCountries(objNode)
	nodes.removeNodeFromContinents(objNode)

	klog.Infof("node deleted from cache: %s\n", objNode.Name)
}

func (nodes *Nodes) removeNodeFromNodes(objNode *v1.Node) {
	for i, v := range nodes.Nodes {
		if v.Name == objNode.Name {
			nodes.Nodes = append(nodes.Nodes[:i], nodes.Nodes[i+1:]...)
		}
	}
}

func (nodes *Nodes) removeNodeFromCities(objNode *v1.Node) {
	cityValue := objNode.Labels[cityLabel]

	if cityValue != "" {
		if city, err := nodes.Query.FindSubdivisionByName(cityValue); err == nil {
			cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
			for i, v := range nodes.Cities[cityCode] {
				if v.Name == objNode.Name {
					nodes.Cities[cityCode] = append(nodes.Cities[cityCode][:i], nodes.Cities[cityCode][i+1:]...)
				}
			}
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *Nodes) removeNodeFromCountries(objNode *v1.Node) {
	countryValue := objNode.Labels[countryLabel]
	if countryValue != "" {
		if country, err := nodes.findCountry(countryValue); err == nil {
			for i, v := range nodes.Countries[country.Alpha2] {
				if v.Name == objNode.Name {
					nodes.Countries[country.Alpha2] =
						append(nodes.Countries[country.Alpha2][:i], nodes.Countries[country.Alpha2][i+1:]...)
				}
			}
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *Nodes) removeNodeFromContinents(objNode *v1.Node) {
	continentValue := objNode.Labels[continentLabel]

	if continentValue != "" {
		if continent, err := nodes.ContinentsList.FindContinent(continentValue); err == nil {
			for i, v := range nodes.Continents[continent.Code] {
				if v.Name == objNode.Name {
					nodes.Continents[continent.Code] =
						append(nodes.Continents[continent.Code][:i], nodes.Continents[continent.Code][i+1:]...)
				}
			}
		} else {
			klog.Errorln(err)
		}
	}
}
