package utils

import (
	"github.com/aida-dos/gountries"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func newTestNode(name string, edge bool, city string, country string, continent string) *v1.Node {
	labels := map[string]string{
		cityLabel:      city,
		countryLabel:   country,
		continentLabel: continent,
	}

	if edge {
		labels["node-role.kubernetes.io/edge"] = ""
	}

	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func newTestNodes() *nodes {
	return &nodes{
		clientSet: nil,

		query:          gountries.New(),
		continentsList: gountries.NewContinents(),

		nodes:      make([]*Node, 0),
		cities:     make(map[string][]*Node),
		countries:  make(map[string][]*Node),
		continents: make(map[string][]*Node),
	}
}

func TestNew(t *testing.T) {
	nodes := New(nil)
	assert.Equal(t, 0, nodes.CountNodes())
}

func TestGetAndCountNodes(t *testing.T) {
	nodes := newTestNodes()
	assert.Equal(t, 0, nodes.CountNodes())
	assert.Equal(t, 0, len(nodes.GetAllNodes()))
}

func TestAddEdgeNode(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	cityNode, _ := nodes.FindAnyNodeByCity([]string{"Braga"})
	assert.Equal(t, "Node0", cityNode.Name)

	countryNode, _ := nodes.FindAnyNodeByCountry([]string{"Portugal"})
	assert.Equal(t, "Node0", countryNode.Name)

	continentNode, _ := nodes.FindAnyNodeByContinent([]string{"Europe"})
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestAddEdgeNodeError(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "RANDOM_C_123", "RANDOM_C_123", "RANDOM_C_123")

	nodes.addNode(node)
	assert.Equal(t, 1, nodes.CountNodes())
	assert.Equal(t, 0, len(nodes.cities))
	assert.Equal(t, 0, len(nodes.countries))
	assert.Equal(t, 0, len(nodes.continents))
}

func TestAddNormalNode(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", false, "", "", "")

	nodes.addNode(node)
	assert.Equal(t, 0, nodes.CountNodes())
}

func TestUpdateNode(t *testing.T) {
	nodes := newTestNodes()
	oldNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(oldNode)

	assert.Equal(t, 1, len(nodes.cities["PT-03"]))
	assert.Equal(t, 0, len(nodes.cities["PT-13"]))

	newNode := newTestNode("Node0", true, "Porto", "Portugal", "Europe")
	nodes.updateNode(oldNode, newNode)

	assert.Equal(t, 0, len(nodes.cities["PT-03"]))
	assert.Equal(t, 1, len(nodes.cities["PT-13"]))
}

func TestDeleteEdgeNode(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)
	nodes.deleteNode(node)
	assert.Equal(t, 0, nodes.CountNodes())
}

func TestDeleteNormalNode(t *testing.T) {
	nodes := newTestNodes()
	addNode := newTestNode("Node0", true, "", "", "")
	nodes.addNode(addNode)

	deleteNode := newTestNode("Node0", false, "", "", "")
	nodes.deleteNode(deleteNode)

	assert.Equal(t, 1, nodes.CountNodes())
}

func TestDeleteErrorExitsGracefully(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "RANDOM_C_123", "RANDOM_C_123", "RANDOM_C_123")
	nodes.deleteNode(node)
}

func TestFindByCity(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	cityNode, _ := nodes.FindAnyNodeByCity([]string{"Braga"})
	assert.Equal(t, "Node0", cityNode.Name)

	countryNode, _ := nodes.FindAnyNodeByCityCountry([]string{"Porto"})
	assert.Equal(t, "Node0", countryNode.Name)

	continentNode, _ := nodes.FindAnyNodeByCityContinent([]string{"Madrid"})
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestFindByCityError(t *testing.T) {
	nodes := newTestNodes()

	_, err0 := nodes.FindAnyNodeByCity([]string{"Braga"})
	assert.Error(t, err0)

	_, err1 := nodes.FindAnyNodeByCityCountry([]string{"Braga"})
	assert.Error(t, err1)

	_, err2 := nodes.FindAnyNodeByCityCountry([]string{"RANDOM_C_123"})
	assert.Error(t, err2)

	_, err3 := nodes.FindAnyNodeByCityContinent([]string{"Braga"})
	assert.Error(t, err3)

	_, err4 := nodes.FindAnyNodeByCityContinent([]string{"RANDOM_C_123"})
	assert.Error(t, err4)

	_, err5 := nodes.getNodeByCity("Braga")
	assert.Error(t, err5)

	_, err6 := nodes.getNodeByCity("RANDOM_C_123")
	assert.Error(t, err6)
}

func TestFindByCountry(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	countryNode, _ := nodes.FindAnyNodeByCountry([]string{"PT"})
	assert.Equal(t, "Node0", countryNode.Name)

	continentNode, _ := nodes.FindAnyNodeByCountryContinent([]string{"Spain"})
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestFindByCountryError(t *testing.T) {
	nodes := newTestNodes()

	_, err0 := nodes.FindAnyNodeByCountry([]string{"Portugal"})
	assert.Error(t, err0)

	_, err1 := nodes.FindAnyNodeByCountry([]string{"RANDOM_C_123"})
	assert.Error(t, err1)

	_, err2 := nodes.FindAnyNodeByCountryContinent([]string{"Portugal"})
	assert.Error(t, err2)

	_, err3 := nodes.FindAnyNodeByCountryContinent([]string{"RANDOM_C_123"})
	assert.Error(t, err3)

	_, err4 := nodes.getNodeByCountry("Portugal")
	assert.Error(t, err4)
}

func TestFindByContinent(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	continentNode, _ := nodes.FindAnyNodeByContinent([]string{"Europe"})
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestFindByContinentError(t *testing.T) {
	nodes := newTestNodes()

	_, err0 := nodes.FindAnyNodeByContinent([]string{"Europe"})
	assert.Error(t, err0)

	_, err1 := nodes.getNodeByContinent("Europe")
	assert.Error(t, err1)

	_, err2 := nodes.getNodeByContinent("RANDOM_C_123")
	assert.Error(t, err2)
}

func TestNodeHasSignificantChanges(t *testing.T) {
	oldNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")
	newNode := newTestNode("Node0", true, "Porto", "Portugal", "Europe")

	assert.Equal(t, false, nodeHasSignificantChanges(oldNode, oldNode))
	assert.Equal(t, true, nodeHasSignificantChanges(oldNode, newNode))
}

func TestAddHandler(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addHandler(node)
	assert.Equal(t, 1, nodes.CountNodes())
}

func TestEdgeToEdgeUpdateHandler(t *testing.T) {
	nodes := newTestNodes()
	oldNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")
	newNode := newTestNode("Node0", true, "Porto", "Portugal", "Europe")

	nodes.addHandler(oldNode)
	assert.Equal(t, 1, len(nodes.cities["PT-03"]))
	assert.Equal(t, 0, len(nodes.cities["PT-13"]))
	nodes.updateHandler(oldNode, newNode)
	assert.Equal(t, 0, len(nodes.cities["PT-03"]))
	assert.Equal(t, 1, len(nodes.cities["PT-13"]))
}

func TestNormalToEdgeUpdateHandler(t *testing.T) {
	nodes := newTestNodes()
	oldNode := newTestNode("Node0", false, "", "", "")
	newNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addHandler(oldNode)
	assert.Equal(t, 0, nodes.CountNodes())
	nodes.updateHandler(oldNode, newNode)
	assert.Equal(t, 1, nodes.CountNodes())
}

func TestEdgeToNormalUpdateHandler(t *testing.T) {
	nodes := newTestNodes()
	oldNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")
	newNode := newTestNode("Node0", false, "", "", "")

	nodes.addHandler(oldNode)
	assert.Equal(t, 1, nodes.CountNodes())
	nodes.updateHandler(oldNode, newNode)
	assert.Equal(t, 0, nodes.CountNodes())
}

func TestDeleteHandler(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addHandler(node)
	assert.Equal(t, 1, nodes.CountNodes())
	nodes.deleteHandler(node)
	assert.Equal(t, 0, nodes.CountNodes())
}
