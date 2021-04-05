package nodes

import (
	"errors"
	"fmt"
	"github.com/aida-dos/gountries"
	"k8s.io/klog/v2"
)

func (nodes *Nodes) getNodesByCity(cityName string, filter *NodeFilter) ([]*Node, error) {
	city, err := nodes.Query.FindSubdivisionByName(cityName)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	cityCode := fmt.Sprintf("%s-%s", city.CountryAlpha2, city.Code)
	if options, ok := nodes.Cities[cityCode]; ok {
		return filterNodes(options, filter), nil
	}

	return nil, errors.New("no node matches city")
}

func (nodes *Nodes) getNodesByCountry(countryCode string, filter *NodeFilter) ([]*Node, error) {
	if options, ok := nodes.Countries[countryCode]; ok {
		return filterNodes(options, filter), nil
	}

	return nil, errors.New("no nodes match given country")
}

func (nodes *Nodes) getNodesByContinent(continentName string, filter *NodeFilter) ([]*Node, error) {
	continent, err := nodes.ContinentsList.FindContinent(continentName)

	if err != nil {
		klog.Errorln(err)
	}

	if options, ok := nodes.Continents[continent.Code]; ok {
		return filterNodes(options, filter), nil
	}

	return nil, errors.New("no node matches given continent")
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
