package graph

import "fmt"

var ErrNoGraph = fmt.Errorf("no graph is associated with this path")

type OutOfBoundsErr struct {
	Type  string
	Index int
}

func (err OutOfBoundsErr) Error() string {
	return fmt.Sprintf("%s index: %d is out of bounds", err.Type, err.Index)
}

type Graph struct {
	vertexes []vertex
	edges    []edge
}

// returns the number of vertices defined in this graph
func (g Graph) Order() int {
	return len(g.vertexes)
}

// returns the number of edges defined in this graph
func (g Graph) Size() int {
	return len(g.edges)
}

// appends a vertex to this graph
// returns the id associated with this vertex. vertexes are 0-based indexed.
func (g *Graph) AddVertex(data interface{}) int {
	id := int(len(g.vertexes))
	v := vertex{
		id:   id,
		out:  make([]out, 0),
		Data: data,
	}
	g.vertexes = append(g.vertexes, v)
	return id
}

// appends an an undirected edge whose endpoints are V[vi] and V[vj] to this graph
// returns the id associated with the resulting edge. edges are 0-based indexed.
// an error is returned if 0 >= vi, vj < len(V)
func (g *Graph) AddEdge(vi, vj int, data interface{}) (int, error) {
	var edgeId int = -1
	if err := g.testVertex(vi, vj); err != nil {
		return edgeId, err
	}
	edgeId = len(g.edges)
	edge := edge{
		id:        edgeId,
		endpoints: []int{vi, vj},
		Data:      data,
	}
	g.edges = append(g.edges, edge)
	g.vertexes[vi].out = append(g.vertexes[vi].out, out{vj, edgeId})
	if vi != vj {
		g.vertexes[vj].out = append(g.vertexes[vj].out, out{vi, edgeId})
	}
	return edgeId, nil
}

// returns an array of vertexes and edges adjacent to this vertex
// or an error if the vertexIndex is out of bounds
func (g Graph) GetAdjacencies(vertexIndex int) ([]out, error) {
	if err := g.testVertex(vertexIndex); err != nil {
		return nil, err
	}
	return g.vertexes[vertexIndex].out, nil
}

// returns a pointer to the the vertex V[vertexIndex]
// or error if the vertexIndex is out of bounds
func (g Graph) GetVertex(vertexIndex int) (*vertex, error) {
	if err := g.testVertex(vertexIndex); err != nil {
		return nil, err
	}
	return &g.vertexes[vertexIndex], nil
}

// returns a pointer to the edge E[edgeIndex]
// or error if the edgeIndex is out of bounds
func (g Graph) GetEdge(edgeIndex int) (*edge, error) {
	if err := g.testEdge(edgeIndex); err != nil {
		return nil, err
	}
	return &g.edges[edgeIndex], nil
}

// returns true if 0 >= index < len(V)
// returns false otherwise
func (g Graph) vertexExists(index int) bool {
	if index < 0 || index >= len(g.vertexes) {
		return false
	}
	return true
}

// returns an error if any of the indexes is out of bounds
func (g Graph) testVertex(indexList ...int) error {
	for i := 0; i < len(indexList); i++ {
		if b := g.vertexExists(indexList[i]); !b {
			return OutOfBoundsErr{"vertex", indexList[i]}
		}
	}
	return nil
}

// returns true if 0 >= index < len(E)
// returns false otherwise
func (g Graph) EdgeExists(index int) bool {
	if index < 0 || index >= len(g.edges) {
		return false
	}
	return true
}

// returns an error if any of the indexes is out of bounds
func (g Graph) testEdge(indexList ...int) error {
	for i := 0; i < len(indexList); i++ {
		if b := g.EdgeExists(indexList[i]); !b {
			return OutOfBoundsErr{"edge", indexList[i]}
		}
	}
	return nil
}
