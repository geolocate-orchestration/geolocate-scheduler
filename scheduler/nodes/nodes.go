package nodes

import (
	"aida-scheduler/utils"
	"errors"
	"fmt"
	"github.com/aida-dos/gountries"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
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

// GetAllNodes list all cluster edge nodes
func (nodes *Nodes) GetAllNodes() []*Node {
	return nodes.Nodes
}

// CountNodes returns the number of cluster edge nodes
func (nodes *Nodes) CountNodes() int {
	return len(nodes.Nodes)
}

// FindAnyNodeByCity returns one node in given city if exists
func (nodes *Nodes) FindAnyNodeByCity(cities []string) (*Node, error) {
	for _, city := range cities {
		if node, err := nodes.getNodeByCity(city); err == nil {
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities")
}

// FindAnyNodeByCityCountry returns one node in given city country if exists
func (nodes *Nodes) FindAnyNodeByCityCountry(cities []string) (*Node, error) {
	for _, city := range cities {
		countryResult, err := nodes.Query.FindSubdivisionCountryByName(city)
		if err != nil {
			// If subdivision name does not exists return error
			return nil, err
		}
		// If any node exists in the given city country return it
		if node, err := nodes.getNodeByCountry(countryResult.Alpha2); err == nil {
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

// FindAnyNodeByCityContinent returns one node in given city continent if exists
func (nodes *Nodes) FindAnyNodeByCityContinent(cities []string) (*Node, error) {
	for _, city := range cities {
		countryResult, err := nodes.Query.FindSubdivisionCountryByName(city)
		if err != nil {
			// If subdivision name does not exists return error
			return nil, err
		}
		// If any node exists in the given city continent return it
		if node, err := nodes.getNodeByContinent(countryResult.Continent); err == nil {
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

// FindAnyNodeByCountry returns one node in given country if exists
func (nodes *Nodes) FindAnyNodeByCountry(countries []string) (*Node, error) {
	for _, countryID := range countries {
		if country, err := nodes.findCountry(countryID); err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		} else if node, err := nodes.getNodeByCountry(country.Alpha2); err == nil {
			// If any node exists in the given country return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries")
}

// FindAnyNodeByCountryContinent returns one node in given country continent if exists
func (nodes *Nodes) FindAnyNodeByCountryContinent(countries []string) (*Node, error) {
	for _, countryID := range countries {
		if country, err := nodes.findCountry(countryID); err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		} else if node, err := nodes.getNodeByContinent(country.Continent); err == nil {
			// If any node exists in the given continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries continents")
}

// FindAnyNodeByContinent returns one node in given continent if exists
func (nodes *Nodes) FindAnyNodeByContinent(continents []string) (*Node, error) {
	for _, continentsID := range continents {
		if node, err := nodes.getNodeByContinent(continentsID); err == nil {
			// If any node exists in the given continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given continents")
}

// Private

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
	} else if oldHasEdgeLabel && newHasEdgeLabel && nodeHasSignificantChanges(oldNode, newNode) {
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

func (nodes *Nodes) addNode(objNode *v1.Node) {
	if _, ok := objNode.Labels["node-role.kubernetes.io/edge"]; !ok {
		// Don't add new node if it doesn't have the edge role
		return
	}

	node := &Node{Name: objNode.Name}
	cityValue := objNode.Labels[cityLabel]
	countryValue := objNode.Labels[countryLabel]
	continentValue := objNode.Labels[continentLabel]

	nodes.Nodes = append(nodes.Nodes, node)

	if cityValue != "" {
		if city, err := nodes.Query.FindSubdivisionByName(cityValue); err == nil {
			cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
			nodes.Cities[cityCode] = append(nodes.Cities[cityCode], node)
		} else {
			klog.Errorln(err)
		}
	}

	if countryValue != "" {
		if country, err := nodes.findCountry(countryValue); err == nil {
			nodes.Countries[country.Alpha2] = append(nodes.Countries[country.Alpha2], node)
		} else {
			klog.Errorln(err)
		}
	}

	if continentValue != "" {
		if continent, err := nodes.ContinentsList.FindContinent(continentValue); err == nil {
			nodes.Continents[continent.Code] = append(nodes.Continents[continent.Code], node)
		} else {
			klog.Errorln(err)
		}
	}

	klog.Infof("new node added to cache: %s\n", node.Name)
}

func (nodes *Nodes) updateNode(oldNode *v1.Node, newNode *v1.Node) {
	klog.Infof("node will be updated in cache: %s\n", oldNode.Name)
	// TODO: I know it could be more efficient
	nodes.deleteNode(oldNode)
	nodes.addNode(newNode)
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

func (nodes *Nodes) getNodeByCity(cityName string) (*Node, error) {
	city, err := nodes.Query.FindSubdivisionByName(cityName)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
	if options, ok := nodes.Cities[cityCode]; ok {
		return getRandom(options), nil
	}

	return nil, errors.New("no node matches city")
}

func (nodes *Nodes) getNodeByCountry(countryCode string) (*Node, error) {
	if options, ok := nodes.Countries[countryCode]; ok {
		return getRandom(options), nil
	}

	return nil, errors.New("no nodes match given country")
}

func (nodes *Nodes) getNodeByContinent(continentName string) (*Node, error) {
	continent, err := nodes.ContinentsList.FindContinent(continentName)

	if err != nil {
		klog.Errorln(err)
	}

	if options, ok := nodes.Continents[continent.Code]; ok {
		return getRandom(options), nil
	}

	return nil, errors.New("no node matches given continent")
}

func getRandom(options []*Node) *Node {
	return options[rand.Intn(len(options))]
}

func (nodes *Nodes) findCountry(countryID string) (gountries.Country, error) {
	if country, err := nodes.Query.FindCountryByName(countryID); err == nil {
		return country, nil
	}

	if country, err := nodes.Query.FindCountryByAlpha(countryID); err == nil {
		return country, nil
	}

	return gountries.Country{}, errors.New("given country identifier does not match any country")
}

func nodeHasSignificantChanges(oldNode *v1.Node, newNode *v1.Node) bool {
	return oldNode.Name != newNode.Name ||
		oldNode.Labels[cityLabel] != newNode.Labels[cityLabel] ||
		oldNode.Labels[countryLabel] != newNode.Labels[countryLabel] ||
		oldNode.Labels[continentLabel] != newNode.Labels[continentLabel]
}
