package utils

import (
	"errors"
	"github.com/aida-dos/gountries"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Nodes interface {
	GetAllNodes() []*Node
	CountNodes() int
	FindAnyNodeByCity(cities []string) (*Node, error)
	FindAnyNodeCityCountry(cities []string) (*Node, error)
	FindAnyNodeCityContinent(cities []string) (*Node, error)
	FindAnyNodeByCountry(countries []string) (*Node, error)
	FindAnyNodeByCountryContinent(countries []string) (*Node, error)
	FindAnyNodeByContinent(continents []string) (*Node, error)
}

type nodes struct {
	clientSet *kubernetes.Clientset
	nodes []*Node
	cities map[string][]*Node
	countries map[string][]*Node
	continents map[string][]*Node
}

type Node struct {
	Name string
}

func New(clientSet *kubernetes.Clientset) Nodes {
	nodes := nodes{
		clientSet : clientSet,
		nodes     : make([]*Node, 0),
		cities    : make(map[string][]*Node),
		countries : make(map[string][]*Node),
		continents: make(map[string][]*Node),
	}
	nodes.startNodeInformerHandler()
	return &nodes
}

func (nodes *nodes) GetAllNodes() []*Node {
	return nodes.nodes
}

func (nodes *nodes) CountNodes() int {
	return len(nodes.nodes)
}

func (nodes *nodes) FindAnyNodeByCity(cities []string) (*Node, error) {
	for _, city := range cities {
		if node, err := nodes.getNodeByCity(city); err != nil {
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities")
}

func (nodes *nodes) FindAnyNodeCityCountry(cities []string) (*Node, error)  {
	query := gountries.New()

	for _, city := range cities {
		if countryResult, err := query.FindSubdivisionCountryByName(city); err != nil {
			// If subdivision name does not exists return error
			return nil, err
		} else {
			// If any node exists in the given city country return it
			if node, err := nodes.getNodeByCountry(countryResult.Alpha3); err == nil {
				return node, nil
			}
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

func (nodes *nodes) FindAnyNodeCityContinent(cities []string) (*Node, error)  {
	query := gountries.New()

	for _, city := range cities {
		if countryResult, err := query.FindSubdivisionCountryByName(city); err != nil {
			// If subdivision name does not exists return error
			return nil, err
		} else {
			// If any node exists in the given city continent return it
			if node, err := nodes.getNodeByContinent(countryResult.Continent); err == nil {
				return node, nil
			}
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

func (nodes *nodes) FindAnyNodeByCountry(countries []string) (*Node, error) {
	for _, countryId := range countries {
		if country, err := findCountry(countryId); err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		} else if node, err := nodes.getNodeByCountry(country.Alpha3); err != nil {
			// If any node exists in the given country return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries")
}

func (nodes *nodes) FindAnyNodeByCountryContinent(countries []string) (*Node, error) {
	for _, countryId := range countries {
		if country, err := findCountry(countryId); err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		} else if node, err := nodes.getNodeByContinent(country.Continent); err != nil {
			// If any node exists in the given continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries continents")
}


func (nodes *nodes) FindAnyNodeByContinent(continents []string) (*Node, error) {
	for _, continentsId := range continents {
		 if node, err := nodes.getNodeByContinent(continentsId); err != nil {
			// If any node exists in the given continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given continents")
}


// Private

func (nodes *nodes) startNodeInformerHandler() {
	factory := informers.NewSharedInformerFactory(nodes.clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes()

	stopper := make(chan struct{})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			nodes.newNode(obj.(*v1.Node))
		},
		DeleteFunc: func(obj interface{}) {
			nodes.deleteNode(obj.(*v1.Node))
		},
	})

	factory.Start(stopper)
}

func (nodes *nodes) newNode(objNode *v1.Node) {
	node := &Node{Name: objNode.Name}
	nodes.nodes = append(nodes.nodes, node)

	city := objNode.Labels["deployment.edge.aida.io/city"]
	nodes.cities[city] = append(nodes.cities[city], node)

	country := objNode.Labels["deployment.edge.aida.io/country"]
	nodes.countries[country] = append(nodes.countries[country], node)

	continent := objNode.Labels["deployment.edge.aida.io/continent"]
	nodes.continents[continent] = append(nodes.continents[continent], node)

	klog.Infof("new node added to cache: %s\n", node.Name)
}

func (nodes *nodes) deleteNode(objNode *v1.Node) {
	nodes.removeNodeFromNodes(objNode)
	nodes.removeNodeFromCities(objNode)
	nodes.removeNodeFromCountries(objNode)
	nodes.removeNodeFromContinents(objNode)

	klog.Infof("node deleted from cache: %s\n", objNode.Name)
}

func (nodes *nodes) removeNodeFromNodes(objNode *v1.Node) {
	for i, v := range nodes.nodes {
		if v.Name == objNode.Name {
			nodes.nodes = append(nodes.nodes[:i], nodes.nodes[i+1:]...)
		}
	}
}

func (nodes *nodes) removeNodeFromCities(objNode *v1.Node) {
	city := objNode.Labels["deployment.edge.aida.io/city"]
	for i, v := range nodes.cities[city] {
		if v.Name == objNode.Name {
			nodes.cities[city] = append(nodes.cities[city][:i], nodes.cities[city][i+1:]...)
		}
	}
}

func (nodes *nodes) removeNodeFromCountries(objNode *v1.Node) {
	country := objNode.Labels["deployment.edge.aida.io/country"]
	for i, v := range nodes.countries[country] {
		if v.Name == objNode.Name {
			nodes.countries[country] = append(nodes.countries[country][:i], nodes.countries[country][i+1:]...)
		}
	}
}

func (nodes *nodes) removeNodeFromContinents(objNode *v1.Node) {
	continent := objNode.Labels["deployment.edge.aida.io/continent"]
	for i, v := range nodes.continents[continent] {
		if v.Name == objNode.Name {
			nodes.continents[continent] = append(nodes.continents[continent][:i], nodes.continents[continent][i+1:]...)
		}
	}
}

func (nodes *nodes) getNodeByCity(cityName string) (*Node, error) {
	if options, ok := nodes.cities[cityName]; ok {
		return getRandom(options), nil
	}

	return nil, errors.New("no node matches city")
}

func (nodes *nodes) getNodeByCountry(countryCode string) (*Node, error) {
	if options, ok := nodes.countries[countryCode]; ok {
		return getRandom(options), nil
	}

	return nil, errors.New("no nodes match given country")
}

func (nodes *nodes) getNodeByContinent(continentName string) (*Node, error) {
	if options, ok := nodes.continents[continentName]; ok {
		return getRandom(options), nil
	}

	return nil, errors.New("no node matches given continent")
}

func getRandom(options []*Node) *Node {
	return options[rand.Intn(len(options))]
}

func findCountry(countryId string) (gountries.Country, error) {
	query := gountries.New()

	if country, err := query.FindCountryByName(countryId); err == nil {
		return country, nil
	}

	if country, err := query.FindCountryByAlpha(countryId); err == nil {
		return country, nil
	}

	return gountries.Country{}, errors.New("given country identifier does not match any country")
}
