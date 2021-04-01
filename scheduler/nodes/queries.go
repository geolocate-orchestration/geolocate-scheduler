package nodes

import (
	"errors"
)

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
