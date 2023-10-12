package graph

func Trace(g Graph, startVertex int) (Path, error) {
	path, err := g.NewPath(startVertex)
	if err != nil {
		return path, err
	}
	visitedEdges := make([]bool, g.Size())
	nextUnvisitedNeighbor := make([]int, g.Order())
	currentVertex := startVertex
	for {
		neighbors, err := g.GetAdjacencies(currentVertex)
		if err != nil {
			return path, err
		}
		if nextUnvisitedNeighbor[currentVertex] == len(neighbors) {
			break
		}
		neighbor := neighbors[nextUnvisitedNeighbor[currentVertex]]
		nextVertex, nextEdge := neighbor.Endpoint, neighbor.Edge
		nextUnvisitedNeighbor[currentVertex]++
		if visitedEdges[nextEdge] {
			continue
		}
		visitedEdges[nextEdge] = true
		path, err = path.Grow(nextVertex, nextEdge)
		if err != nil {
			return path, err
		}
		currentVertex = nextVertex
	}
	return path, err
}
