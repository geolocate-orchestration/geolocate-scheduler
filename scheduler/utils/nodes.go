package utils

import (
	"errors"
	"fmt"
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

	query          *gountries.Query
	continentsList gountries.Continents

	nodes      []*Node
	cities     map[string][]*Node
	countries  map[string][]*Node
	continents map[string][]*Node
}

type Node struct {
	Name string
}

func New(clientSet *kubernetes.Clientset) Nodes {
	nodes := nodes{
		clientSet: clientSet,

		query:          gountries.New(),
		continentsList: gountries.NewContinents(),

		nodes:      make([]*Node, 0),
		cities:     make(map[string][]*Node),
		countries:  make(map[string][]*Node),
		continents: make(map[string][]*Node),
	}
	nodes.startNodeInformerHandler()
	return &nodes
}

const (
	cityLabel      = "node.edge.aida.io/city"
	countryLabel   = "node.edge.aida.io/country"
	continentLabel = "node.edge.aida.io/continent"
)

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

func (nodes *nodes) FindAnyNodeCityCountry(cities []string) (*Node, error) {
	for _, city := range cities {
		if countryResult, err := nodes.query.FindSubdivisionCountryByName(city); err != nil {
			// If subdivision name does not exists return error
			return nil, err
		} else {
			// If any node exists in the given city country return it
			if node, err := nodes.getNodeByCountry(countryResult.Alpha2); err == nil {
				return node, nil
			}
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

func (nodes *nodes) FindAnyNodeCityContinent(cities []string) (*Node, error) {
	for _, city := range cities {
		if countryResult, err := nodes.query.FindSubdivisionCountryByName(city); err != nil {
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
		if country, err := nodes.findCountry(countryId); err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		} else if node, err := nodes.getNodeByCountry(country.Alpha2); err != nil {
			// If any node exists in the given country return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries")
}

func (nodes *nodes) FindAnyNodeByCountryContinent(countries []string) (*Node, error) {
	for _, countryId := range countries {
		if country, err := nodes.findCountry(countryId); err != nil {
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
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldNode := oldObj.(*v1.Node)
			newNode := newObj.(*v1.Node)

			_, oldHasEdgeLabel := oldNode.Labels["node-role.kubernetes.io/edge"]
			_, newHasEdgeLabel := newNode.Labels["node-role.kubernetes.io/edge"]

			if !oldHasEdgeLabel && newHasEdgeLabel {
				// If node wasn't an edge node but now it is, create it in cache
				nodes.newNode(newNode)
			} else if oldHasEdgeLabel && newHasEdgeLabel && nodeHasSignificantChanges(oldNode, newNode) {
				// If the node is an edge node and has significant update it in cache
				nodes.updateNode(oldNode, newNode)
			} else if oldHasEdgeLabel && !newHasEdgeLabel {
				// If node was an edge node but now it isn't, remove it from cache
				nodes.deleteNode(oldNode)
			}
		},
		DeleteFunc: func(obj interface{}) {
			nodes.deleteNode(obj.(*v1.Node))
		},
	})

	factory.Start(stopper)
}

func (nodes *nodes) newNode(objNode *v1.Node) {
	if _, ok := objNode.Labels["node-role.kubernetes.io/edge"]; !ok {
		// Don't add new node if it doesn't have the edge role
		return
	}

	node := &Node{Name: objNode.Name}
	cityValue := objNode.Labels[cityLabel]
	countryValue := objNode.Labels[countryLabel]
	continentValue := objNode.Labels[continentLabel]

	nodes.nodes = append(nodes.nodes, node)

	if cityValue != "" {
		if city, err := nodes.query.FindSubdivisionByName(cityValue); err == nil {
			cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
			nodes.cities[cityCode] = append(nodes.cities[cityCode], node)
		} else {
			klog.Errorln(err)
		}
	}

	if countryValue != "" {
		if country, err := nodes.findCountry(countryValue); err == nil {
			nodes.countries[country.Alpha2] = append(nodes.countries[country.Alpha2], node)
		} else {
			klog.Errorln(err)
		}
	}

	if continentValue != "" {
		if continent, err := nodes.continentsList.FindContinent(continentValue); err == nil {
			nodes.continents[continent.Code] = append(nodes.continents[continent.Code], node)
		} else {
			klog.Errorln(err)
		}
	}

	klog.Infof("new node added to cache: %s\n", node.Name)
}

func (nodes *nodes) updateNode(oldNode *v1.Node, newNode *v1.Node) {
	klog.Infof("node will be updated in cache: %s\n", oldNode.Name)
	// TODO: I know it could be more efficient
	nodes.deleteNode(oldNode)
	nodes.newNode(newNode)
}

func (nodes *nodes) deleteNode(objNode *v1.Node) {
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

func (nodes *nodes) removeNodeFromNodes(objNode *v1.Node) {
	for i, v := range nodes.nodes {
		if v.Name == objNode.Name {
			nodes.nodes = append(nodes.nodes[:i], nodes.nodes[i+1:]...)
		}
	}
}

func (nodes *nodes) removeNodeFromCities(objNode *v1.Node) {
	cityValue := objNode.Labels[cityLabel]

	if cityValue != "" {
		if city, err := nodes.query.FindSubdivisionByName(cityValue); err == nil {
			cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
			for i, v := range nodes.cities[cityCode] {
				if v.Name == objNode.Name {
					nodes.cities[cityCode] = append(nodes.cities[cityCode][:i], nodes.cities[cityCode][i+1:]...)
				}
			}
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *nodes) removeNodeFromCountries(objNode *v1.Node) {
	countryValue := objNode.Labels[countryLabel]
	if countryValue != "" {
		if country, err := nodes.findCountry(countryValue); err == nil {
			for i, v := range nodes.countries[country.Alpha2] {
				if v.Name == objNode.Name {
					nodes.countries[country.Alpha2] =
						append(nodes.countries[country.Alpha2][:i], nodes.countries[country.Alpha2][i+1:]...)
				}
			}
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *nodes) removeNodeFromContinents(objNode *v1.Node) {
	continentValue := objNode.Labels[continentLabel]

	if continentValue != "" {
		if continent, err := nodes.continentsList.FindContinent(continentValue); err == nil {
			for i, v := range nodes.continents[continent.Code] {
				if v.Name == objNode.Name {
					nodes.continents[continent.Code] =
						append(nodes.continents[continent.Code][:i], nodes.continents[continent.Code][i+1:]...)
				}
			}
		} else {
			klog.Errorln(err)
		}
	}
}

func (nodes *nodes) getNodeByCity(cityName string) (*Node, error) {
	city, err := nodes.query.FindSubdivisionByName(cityName)
	if err != nil {
		return nil, errors.New("no node matches city")
	}

	if options, ok := nodes.cities[city.Code]; ok {
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

func (nodes *nodes) findCountry(countryId string) (gountries.Country, error) {
	if country, err := nodes.query.FindCountryByName(countryId); err == nil {
		return country, nil
	}

	if country, err := nodes.query.FindCountryByAlpha(countryId); err == nil {
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
