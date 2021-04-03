package nodes

import (
	"errors"
)

// CountNodes returns the number of cluster edge nodes
func (nodes *Nodes) CountNodes() int {
	return len(nodes.Nodes)
}

// GetAllNodes list all cluster edge nodes
func (nodes *Nodes) GetAllNodes() []*Node {
	return nodes.Nodes
}

// GetNodes list all cluster edge nodes matching filter
func (nodes *Nodes) GetNodes(filter *NodeFilter) []*Node {
	nodesList := make([]*Node, 0)

	for _, node := range nodes.Nodes {
		if nodeMatchesFilters(node, filter) {
			nodesList = append(nodesList, node)
		}
	}

	return nodesList
}

// FindAnyNodeByCity returns one node in given city if exists
func (nodes *Nodes) FindAnyNodeByCity(cities []string, filter *NodeFilter) (*Node, error) {
	for _, city := range cities {
		node, err := nodes.getNodeByCity(city)
		if err == nil && nodeMatchesFilters(node, filter) {
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities")
}

// FindAnyNodeByCityCountry returns one node in given city country if exists
func (nodes *Nodes) FindAnyNodeByCityCountry(cities []string, filter *NodeFilter) (*Node, error) {
	for _, city := range cities {
		countryResult, err := nodes.Query.FindSubdivisionCountryByName(city)
		if err != nil {
			// If subdivision name does not exists return error
			return nil, err
		}

		node, err := nodes.getNodeByCountry(countryResult.Alpha2)
		if err == nil && nodeMatchesFilters(node, filter) {
			// If any node exists in the given city country return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

// FindAnyNodeByCityContinent returns one node in given city continent if exists
func (nodes *Nodes) FindAnyNodeByCityContinent(cities []string, filter *NodeFilter) (*Node, error) {
	for _, city := range cities {
		countryResult, err := nodes.Query.FindSubdivisionCountryByName(city)
		if err != nil {
			// If subdivision name does not exists return error
			return nil, err
		}

		node, err := nodes.getNodeByContinent(countryResult.Continent)
		if err == nil && nodeMatchesFilters(node, filter) {
			// If any node exists in the given city continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given cities countries nor continents")
}

// FindAnyNodeByCountry returns one node in given country if exists
func (nodes *Nodes) FindAnyNodeByCountry(countries []string, filter *NodeFilter) (*Node, error) {
	for _, countryID := range countries {
		country, err := nodes.findCountry(countryID)
		if err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		}

		node, err := nodes.getNodeByCountry(country.Alpha2)
		if err == nil && nodeMatchesFilters(node, filter) {
			// If any node exists in the given country return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries")
}

// FindAnyNodeByCountryContinent returns one node in given country continent if exists
func (nodes *Nodes) FindAnyNodeByCountryContinent(countries []string, filter *NodeFilter) (*Node, error) {
	for _, countryID := range countries {
		country, err := nodes.findCountry(countryID)
		if err != nil {
			// If country identifier string does not match any country return error
			return nil, err
		}

		node, err := nodes.getNodeByContinent(country.Continent)
		if err == nil && nodeMatchesFilters(node, filter) {
			// If any node exists in the given continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given countries continents")
}

// FindAnyNodeByContinent returns one node in given continent if exists
func (nodes *Nodes) FindAnyNodeByContinent(continents []string, filter *NodeFilter) (*Node, error) {
	for _, continentsID := range continents {
		node, err := nodes.getNodeByContinent(continentsID)
		if err == nil && nodeMatchesFilters(node, filter) {
			// If any node exists in the given continent return it
			return node, nil
		}
	}

	return nil, errors.New("no nodes match given continents")
}
