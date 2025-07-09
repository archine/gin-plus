package topo

import "fmt"

// Edge represents a dependency relationship between beans
type Edge struct {
	From string `json:"from"` // dependent bean
	To   string `json:"to"`   // dependency bean
}

func BuildEdge(from string, to []string) []*Edge {
	edges := make([]*Edge, 0, len(to))
	for _, t := range to {
		edges = append(edges, &Edge{From: from, To: t})
	}
	return edges
}

// Sort performs topological sorting to determine bean creation order
func Sort(edges []*Edge) ([]string, error) {
	edgeMap, inDegree := buildDependencyGraph(edges)

	// check for circular dependencies
	if cycle := detectCircularDependency(edges); cycle != nil {
		return nil, fmt.Errorf("circular dependency detected: %v", cycle)
	}

	queue := getZeroInDegreeNodes(inDegree)
	result := make([]string, 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// reduce in-degree of adjacent nodes
		for _, edge := range edgeMap[current] {
			inDegree[edge.To]--
			if inDegree[edge.To] == 0 {
				queue = append(queue, edge.To)
			}
		}
	}

	return result, nil
}

// detectCircularDependency detects circular dependencies in bean graph
func detectCircularDependency(edges []*Edge) []string {
	edgeMap, _ := buildDependencyGraph(edges)

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for beanName := range edgeMap {
		if !visited[beanName] {
			if foundCycle, cycle := dfsDetectCycle(beanName, edgeMap, visited, recStack, nil); foundCycle {
				return cycle
			}
		}
	}

	return nil
}

// buildDependencyGraph builds adjacency list and in-degree map from edges
func buildDependencyGraph(edges []*Edge) (map[string][]*Edge, map[string]int) {
	edgeMap := make(map[string][]*Edge)
	inDegree := make(map[string]int)

	// initialize all nodes
	for _, edge := range edges {
		if _, exists := inDegree[edge.From]; !exists {
			inDegree[edge.From] = 0
		}
		if _, exists := inDegree[edge.To]; !exists {
			inDegree[edge.To] = 0
		}
	}

	// build adjacency list and in-degree count
	for _, edge := range edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge)
		inDegree[edge.To]++
	}

	return edgeMap, inDegree
}

// getZeroInDegreeNodes returns nodes with zero in-degree (no dependencies)
func getZeroInDegreeNodes(inDegree map[string]int) []string {
	var nodes []string
	for node, degree := range inDegree {
		if degree == 0 {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// dfsDetectCycle uses DFS to detect cycles in dependency graph
func dfsDetectCycle(node string, edgeMap map[string][]*Edge, visited map[string]bool, recStack map[string]bool, currentPath []string) (bool, []string) {
	if recStack[node] {
		// cycle found, construct cycle path
		for i, n := range currentPath {
			if n == node {
				cycle := make([]string, len(currentPath[i:]))
				copy(cycle, currentPath[i:])
				cycle = append(cycle, node) // close the cycle
				return true, cycle
			}
		}
	}

	if visited[node] {
		return false, nil
	}

	visited[node] = true
	recStack[node] = true
	currentPath = append(currentPath, node)

	// traverse all dependencies
	for _, edge := range edgeMap[node] {
		if foundCycle, cycle := dfsDetectCycle(edge.To, edgeMap, visited, recStack, currentPath); foundCycle {
			return true, cycle
		}
	}

	recStack[node] = false
	return false, nil
}
