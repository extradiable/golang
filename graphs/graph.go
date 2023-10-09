package graph

import "fmt"

var NoGraphErr = fmt.Errorf("no graph is associated with this path")

type OutOfBoundsErr struct {
	Type  string
	Index int
}

func (err OutOfBoundsErr) Error() string {
	return fmt.Sprintf("%s index: %d is out of bounds", err.Type, err.Index)
}

type out struct {
	Endpoint int
	Edge     int
}

type Path struct {
	graph *Graph
	start int
	out   []out
}

func (p Path) GetLastVertex() int {
	if len(p.out) == 0 {
		return p.start
	}
	return p.out[len(p.out)-1].Endpoint
}

func (p Path) Grow(vertexIndex, edgeIndex int) (Path, error) {
	if p.graph == nil {
		return p, NoGraphErr
	}
	if err := p.graph.testVertex(vertexIndex); err != nil {
		return p, err
	}
	if err := p.graph.testEdge(edgeIndex); err != nil {
		return p, err
	}

	endpoints := p.graph.edges[edgeIndex].Endpoints
	if endpoints[0] != vertexIndex && endpoints[1] != vertexIndex {
		return p, fmt.Errorf("edge: %d is not incident to vertex: %d", edgeIndex, vertexIndex)
	}
	last := p.GetLastVertex()
	if endpoints[0] != last && endpoints[1] != last {
		return p, fmt.Errorf("edge: %d is not incident to vertex: %d", edgeIndex, last)
	}

	p.out = append(p.out, out{
		Edge:     edgeIndex,
		Endpoint: vertexIndex,
	})
	return p, nil
}

func printPath(vi, vj *Vertex, e *Edge) {
	if e == nil {
		fmt.Printf("v%d\n", vi.id)
	} else {
		fmt.Printf("v%d ->e%d v%d\n", vi.id, e.id, vj.id)
	}
}

func (p Path) PrintPath() {
	p.TraversePath(printPath)
}

func (p Path) TraversePath(fn func(vi, vj *Vertex, e *Edge)) error {
	if p.graph == nil {
		return fmt.Errorf("no graph associated with this path")
	}
	vi, err := p.graph.GetVertex(p.start)
	if err != nil {
		return err
	}
	if len(p.out) == 0 {
		fn(vi, nil, nil)
		return nil
	}
	for _, v := range p.out {
		edge, err := p.graph.GetEdge(v.Edge)
		if err != nil {
			return err
		}
		vj, err := p.graph.GetVertex(v.Endpoint)
		if err != nil {
			return err
		}
		fn(vi, vj, edge)
		vi = vj
	}
	return nil
}

type Vertex struct {
	id   int
	out  []out
	Data interface{}
}

func (v Vertex) GetData() interface{} {
	return v.Data
}

type Edge struct {
	id        int
	Endpoints []int
	Data      interface{}
}

func (e Edge) GetData() interface{} {
	return e.Data
}

type Graph struct {
	vertexes []Vertex
	edges    []Edge
}

func (g Graph) Order() int {
	return len(g.vertexes)
}

func (g Graph) Size() int {
	return len(g.edges)
}

func (g Graph) NewPath(start int) (Path, error) {
	var p Path
	if len(g.vertexes) == 0 {
		return p, fmt.Errorf("cannot create path out of empty graph")
	}
	if err := g.testVertex(start); err != nil {
		return p, err
	}
	p.graph = &g
	p.start = start
	return p, nil
}

func (g *Graph) AddVertex(data interface{}) int {
	id := int(len(g.vertexes))
	v := Vertex{
		id:   id,
		out:  make([]out, 0),
		Data: data,
	}
	g.vertexes = append(g.vertexes, v)
	return id
}

func (g *Graph) AddEdge(vi, vj int, data interface{}) (int, error) {
	var edgeId int = -1
	err := g.testVertex(vi)
	if err != nil {
		return edgeId, err
	}
	err = g.testVertex(vj)
	if err != nil {
		return edgeId, err
	}
	edgeId = len(g.edges)
	edge := Edge{
		id:        edgeId,
		Endpoints: []int{vi, vj},
		Data:      data,
	}
	g.edges = append(g.edges, edge)
	g.vertexes[vi].out = append(g.vertexes[vi].out, out{vj, edgeId})
	if vi != vj {
		g.vertexes[vj].out = append(g.vertexes[vj].out, out{vi, edgeId})
	}
	return edgeId, nil
}

func (g Graph) GetAdjacencies(vertexIndex int) ([]out, error) {
	if err := g.testVertex(vertexIndex); err != nil {
		return nil, err
	}
	return g.vertexes[vertexIndex].out, nil
}

func (g Graph) GetVertex(vertexIndex int) (*Vertex, error) {
	if err := g.testVertex(vertexIndex); err != nil {
		return nil, err
	}
	return &g.vertexes[vertexIndex], nil
}

func (g Graph) GetEdge(edgeIndex int) (*Edge, error) {
	if err := g.testEdge(edgeIndex); err != nil {
		return nil, err
	}
	return &g.edges[edgeIndex], nil
}

func (g Graph) vertexExists(index int) bool {
	if index < 0 || index >= len(g.vertexes) {
		return false
	}
	return true
}

func (g Graph) testVertex(index int) error {
	if b := g.vertexExists(index); b {
		return nil
	} else {
		return OutOfBoundsErr{"vertex", index}
	}
}

func (g Graph) EdgeExists(index int) bool {
	if index < 0 || index >= len(g.edges) {
		return false
	}
	return true
}

func (g Graph) testEdge(index int) error {
	if b := g.EdgeExists(index); b {
		return nil
	} else {
		return OutOfBoundsErr{"edge", index}
	}
}
