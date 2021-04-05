package nodes

import (
	"github.com/aida-dos/gountries"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		ObjectMeta: metaV1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func newTestNodes() *Nodes {
	return &Nodes{
		ClientSet: nil,

		Query:          gountries.New(),
		ContinentsList: gountries.NewContinents(),

		Nodes:      make([]*Node, 0),
		Cities:     make(map[string][]*Node),
		Countries:  make(map[string][]*Node),
		Continents: make(map[string][]*Node),
	}
}

func TestNew(t *testing.T) {
	nodes := New(nil)
	assert.Equal(t, 0, nodes.CountNodes())
}

func TestGetAndCountNodes(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")
	nodes.addNode(node)

	assert.Equal(t, 1, nodes.CountNodes())
	assert.Equal(t, 1, len(nodes.GetAllNodes()))
}

func TestGetNodesFilter(t *testing.T) {
	nodes := newTestNodes()
	filter := &NodeFilter{CPU: 5000}

	node0 := &Node{Name: "Node0", CPU: 2500}
	nodes.Nodes = append(nodes.Nodes, node0)

	assert.Equal(t, 0, len(nodes.GetNodes(filter)))

	node1 := &Node{Name: "Node1", CPU: 10000}
	nodes.Nodes = append(nodes.Nodes, node1)

	assert.Equal(t, 1, len(nodes.GetNodes(filter)))
	assert.Equal(t, "Node1", nodes.GetNodes(filter)[0].Name)
}

func TestAddEdgeNode(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	cityNode := nodes.FindNodesByCity([]string{"Braga"}, nil)["Braga"][0]
	assert.Equal(t, "Node0", cityNode.Name)

	countryNode := nodes.FindNodesByCountry([]string{"Portugal"}, nil)["PT"][0]
	assert.Equal(t, "Node0", countryNode.Name)

	continentNode := nodes.FindNodesByContinent([]string{"Europe"}, nil)["Europe"][0]
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestAddEdgeNodeError(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "RANDOM_C_123", "RANDOM_C_123", "RANDOM_C_123")

	nodes.addNode(node)
	assert.Equal(t, 1, nodes.CountNodes())
	assert.Equal(t, 0, len(nodes.Cities))
	assert.Equal(t, 0, len(nodes.Countries))
	assert.Equal(t, 0, len(nodes.Continents))
}

func TestAddNormalNode(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", false, "", "", "")

	nodes.addNode(node)
	assert.Equal(t, 0, nodes.CountNodes())
}

func TestUpdateNodeCoreData(t *testing.T) {
	nodes := newTestNodes()
	oldNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(oldNode)

	assert.Equal(t, 1, len(nodes.Cities["PT-03"]))
	assert.Equal(t, 0, len(nodes.Cities["PT-13"]))

	newNode := newTestNode("Node0", true, "Porto", "Portugal", "Europe")
	nodes.updateNode(oldNode, newNode)

	assert.Equal(t, 0, len(nodes.Cities["PT-03"]))
	assert.Equal(t, 1, len(nodes.Cities["PT-13"]))
}

func TestUpdateNodeResources(t *testing.T) {
	nodes := newTestNodes()
	oldNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(oldNode)

	newNode := newTestNode("Node0", true, "Braga", "Portugal", "Europe")
	newNode.Labels["test_label"] = "test"

	nodes.updateNode(oldNode, newNode)

	assert.Equal(t, "test", nodes.Nodes[0].Labels["test_label"])
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

	cityNode:= nodes.FindNodesByCity([]string{"Braga"}, nil)["Braga"][0]
	assert.Equal(t, "Node0", cityNode.Name)

	countryNode := nodes.FindNodesByCityCountry([]string{"Porto"}, nil)["PT"][0]
	assert.Equal(t, "Node0", countryNode.Name)

	continentNode := nodes.FindNodesByCityContinent([]string{"Madrid"}, nil)["Europe"][0]
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestFindByCityError(t *testing.T) {
	nodes := newTestNodes()

	values0 := nodes.FindNodesByCity([]string{"Braga"}, nil)
	assert.Equal(t, 0, len(values0))

	values1 := nodes.FindNodesByCityCountry([]string{"Braga"}, nil)
	assert.Equal(t, 0, len(values1))

	values2 := nodes.FindNodesByCityCountry([]string{"RANDOM_C_123"}, nil)
	assert.Equal(t, 0, len(values2))

	values3 := nodes.FindNodesByCityContinent([]string{"Braga"}, nil)
	assert.Equal(t, 0, len(values3))

	values4 := nodes.FindNodesByCityContinent([]string{"RANDOM_C_123"}, nil)
	assert.Equal(t, 0, len(values4))

	_, err5 := nodes.getNodesByCity("Braga", nil)
	assert.Error(t, err5)

	_, err6 := nodes.getNodesByCity("RANDOM_C_123", nil)
	assert.Error(t, err6)
}

func TestFindByCountry(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	countryNode := nodes.FindNodesByCountry([]string{"PT"}, nil)["PT"][0]
	assert.Equal(t, "Node0", countryNode.Name)

	continentNode := nodes.FindNodesByCountryContinent([]string{"Spain"}, nil)["Europe"][0]
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestFindByCountryError(t *testing.T) {
	nodes := newTestNodes()

	values0 := nodes.FindNodesByCountry([]string{"Portugal"}, nil)
	assert.Equal(t, 0, len(values0))

	values1 := nodes.FindNodesByCountry([]string{"RANDOM_C_123"}, nil)
	assert.Equal(t, 0, len(values1))

	values2 := nodes.FindNodesByCountryContinent([]string{"Portugal"}, nil)
	assert.Equal(t, 0, len(values2))

	values3 := nodes.FindNodesByCountryContinent([]string{"RANDOM_C_123"}, nil)
	assert.Equal(t, 0, len(values3))

	_, err4 := nodes.getNodesByCountry("Portugal", nil)
	assert.Error(t, err4)
}

func TestFindByContinent(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	continentNode := nodes.FindNodesByContinent([]string{"Europe"}, nil)["Europe"][0]
	assert.Equal(t, "Node0", continentNode.Name)
}

func TestFindByContinentError(t *testing.T) {
	nodes := newTestNodes()

	values0 := nodes.FindNodesByContinent([]string{"Europe"}, nil)
	assert.Equal(t, 0, len(values0))

	_, err1 := nodes.getNodesByContinent("Europe", nil)
	assert.Error(t, err1)

	_, err2 := nodes.getNodesByContinent("RANDOM_C_123", nil)
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
	assert.Equal(t, 1, len(nodes.Cities["PT-03"]))
	assert.Equal(t, 0, len(nodes.Cities["PT-13"]))
	nodes.updateHandler(oldNode, newNode)
	assert.Equal(t, 0, len(nodes.Cities["PT-03"]))
	assert.Equal(t, 1, len(nodes.Cities["PT-13"]))
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

func TestFindNodeByName(t *testing.T) {
	nodes := newTestNodes()
	node := newTestNode("Node0", true, "Braga", "Portugal", "Europe")

	nodes.addNode(node)

	foundNode, _ := nodes.findNodeByName("Node0")
	assert.Equal(t, "Node0", foundNode.Name)

	_, err := nodes.findNodeByName("Node1")
	assert.Error(t, err)
}

func newNodeFilter(
	nodeLabel string, nodeCPU int64, nodeMemory int64,
	filterLabel string, filterCPU int64, filterMemory int64,
) (*Node, *NodeFilter) {

	node := &Node{
		Labels: map[string]string{
			nodeLabel: "test",
		},
		CPU:    nodeCPU,
		Memory: nodeMemory,
	}

	filter := &NodeFilter{
		Labels: []string{filterLabel},
		CPU:    filterCPU,
		Memory: filterMemory,
	}

	return node, filter
}

func TestNodeFilter(t *testing.T) {
	node, filter := newNodeFilter("test", 10, 10, "test", 1, 1)
	assert.True(t, nodeMatchesFilters(node, filter))
}

func TestNodeNoFilter(t *testing.T) {
	node, _ := newNodeFilter("test", 10, 10, "test", 1, 1)
	assert.True(t, nodeMatchesFilters(node, &NodeFilter{}))
}

func TestNodeFilterFailLabel(t *testing.T) {
	node, filter := newNodeFilter("test", 10, 10, "test1", 1, 1)
	assert.False(t, nodeMatchesFilters(node, filter))
}

func TestNodeFilterFailCPU(t *testing.T) {
	node, filter := newNodeFilter("test", 0, 10, "test", 1, 1)
	assert.False(t, nodeMatchesFilters(node, filter))
}

func TestNodeFilterFailMemory(t *testing.T) {
	node, filter := newNodeFilter("test", 10, 0, "test", 1, 1)
	assert.False(t, nodeMatchesFilters(node, filter))
}

func TestGetRandomFromListHit(t *testing.T) {
	node := &Node{ Name: "Node0" }
	value, _ := GetRandomFromList([]*Node{node})
	assert.Equal(t, "Node0", value.Name)
}

func TestGetRandomFromListEmptyError(t *testing.T) {
	_, err := GetRandomFromList([]*Node{})
	assert.Error(t, err)
}

func TestGetRandomFromMapHit(t *testing.T) {
	node := &Node{ Name: "Node0" }
	value, _ := GetRandomFromMap(map[string][]*Node{"node": {node}})
	assert.Equal(t, "Node0", value.Name)
}

func TestGetRandomFromMapError(t *testing.T) {
	_, err := GetRandomFromMap(map[string][]*Node{})
	assert.Error(t, err)
}
