package nodes

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

// FindNodesByCity returns nodes in given city if exists
func (nodes *Nodes) FindNodesByCity(cities []string, filter *NodeFilter) map[string][]*Node {
	nodesMap := make(map[string][]*Node)

	for _, city := range cities {
		if options, err := nodes.getNodesByCity(city, filter); err == nil {
			nodesMap[city] = options
		}
	}

	return nodesMap
}

// FindNodesByCityCountry returns nodes in given city country if exists
func (nodes *Nodes) FindNodesByCityCountry(cities []string, filter *NodeFilter) map[string][]*Node {
	nodesMap := make(map[string][]*Node)

	for _, city := range cities {
		country, err := nodes.Query.FindSubdivisionCountryByName(city)
		if err != nil {
			// If subdivision name does not exists skip
			continue
		}
		if _, ok := nodesMap[country.Alpha2]; ok {
			// If country already processed skip
			continue
		}

		if options, err := nodes.getNodesByCountry(country.Alpha2, filter); err == nil {
			nodesMap[country.Alpha2] = options
		}
	}

	return nodesMap
}

// FindNodesByCityContinent returns nodes in given city continent if exists
func (nodes *Nodes) FindNodesByCityContinent(cities []string, filter *NodeFilter) map[string][]*Node {
	nodesMap := make(map[string][]*Node)

	for _, city := range cities {
		country, err := nodes.Query.FindSubdivisionCountryByName(city)
		if err != nil {
			// If subdivision name does not exists return error
			continue
		}
		if _, ok := nodesMap[country.Continent]; ok {
			// If country already processed skip
			continue
		}

		if options, err := nodes.getNodesByContinent(country.Continent, filter); err == nil {
			nodesMap[country.Continent] = options
		}
	}

	return nodesMap
}

// FindNodesByCountry returns nodes in given country if exists
func (nodes *Nodes) FindNodesByCountry(countries []string, filter *NodeFilter) map[string][]*Node {
	nodesMap := make(map[string][]*Node)

	for _, countryID := range countries {
		country, err := nodes.findCountry(countryID)
		if err != nil {
			// If country identifier string does not match any country skip
			continue
		}

		if options, err := nodes.getNodesByCountry(country.Alpha2, filter); err == nil {
			nodesMap[country.Alpha2] = options
		}
	}

	return nodesMap
}

// FindNodesByCountryContinent returns nodes in given country continent if exists
func (nodes *Nodes) FindNodesByCountryContinent(countries []string, filter *NodeFilter) map[string][]*Node  {
	nodesMap := make(map[string][]*Node)

	for _, countryID := range countries {
		country, err := nodes.findCountry(countryID)
		if err != nil {
			// If country identifier string does not match any country skip
			continue
		}
		if _, ok := nodesMap[country.Continent]; ok {
			// If country already processed skip
			continue
		}

		if options, err := nodes.getNodesByContinent(country.Continent, filter) ;err == nil {
			nodesMap[country.Continent] = options
		}
	}

	return nodesMap
}

// FindNodesByContinent returns nodes in given continent if exists
func (nodes *Nodes) FindNodesByContinent(continents []string, filter *NodeFilter) map[string][]*Node  {
	nodesMap := make(map[string][]*Node)

	for _, continentsID := range continents {
		if options, err := nodes.getNodesByContinent(continentsID, filter); err == nil {
			nodesMap[continentsID] = options
		}
	}

	return nodesMap
}
