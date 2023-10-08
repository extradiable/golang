package graph

import "fmt"

type path struct {
	Endpoint int
	Edge     int
}

type Path struct {
	start int
	out   []path
}

func (p Path) GetLastVertex() int {
	if len(p.out) == 0 {
		return p.start
	}
	return p.out[len(p.out)-1].Endpoint
}

func (p Path) Grow(g Graph, endpoint, edge int) (Path, error) {
	if endpoint < 0 || endpoint > len(g.vertexes) {
		return p, fmt.Errorf("vertex index: %d is out of bounds", endpoint)
	}
	if edge < 0 || edge > len(g.edges) {
		return p, fmt.Errorf("edge index: %d is out of bounds", endpoint)
	}
	endpoints := g.edges[edge].Endpoints
	if endpoints[0] != endpoint && endpoints[1] != endpoint {
		return p, fmt.Errorf("edge: %d is not incident to vertex: %d", edge, endpoint)
	}
	last := p.GetLastVertex()
	if endpoints[0] != last && endpoints[1] != last {
		return p, fmt.Errorf("edge: %d is not incident to vertex: %d", edge, last)
	}
	p.out = append(p.out, path{
		Edge:     edge,
		Endpoint: endpoint,
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

func (p Path) PrintPath(g Graph) {
	p.TraversePath(g, printPath)
}

func (p Path) TraversePath(g Graph, fn func(vi, vj *Vertex, e *Edge)) error {
	vi, err := g.GetVertex(p.start)
	if err != nil {
		return err
	}
	if len(p.out) == 0 {
		fn(vi, nil, nil)
		return nil
	}
	for _, v := range p.out {
		edge, err := g.GetEdge(v.Edge)
		if err != nil {
			return err
		}
		vj, err := g.GetVertex(v.Endpoint)
		if err != nil {
			return err
		}
		fn(vi, vj, edge)
		vi, err = g.GetVertex(v.Endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

type Vertex struct {
	id   int
	out  []path
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
	var path Path
	path.start = -1
	if len(g.vertexes) == 0 {
		return path, fmt.Errorf("cannot create path out of empty graph")
	}
	if start < 0 || start >= len(g.vertexes) {
		return path, fmt.Errorf("vertex index: %d is out of bounds", start)
	}
	path.start = start
	return path, nil
}

func (g *Graph) AddVertex(data interface{}) int {
	id := int(len(g.vertexes))
	v := Vertex{
		id:   id,
		out:  make([]path, 0),
		Data: data,
	}
	g.vertexes = append(g.vertexes, v)
	return id
}

func (g *Graph) AddEdge(vi, vj int, data interface{}) (int, error) {
	var e int = -1
	if vi < 0 || vj < 0 {
		return e, fmt.Errorf("only positive numbers allowed (%d, %d)", vi, vj)
	}
	if vi >= len(g.vertexes) || vj >= len(g.vertexes) {
		return e, fmt.Errorf("vertex index is out of bounds (%d, %d)", vi, vj)
	}
	e = len(g.edges)
	edge := Edge{
		id:        e,
		Endpoints: []int{vi, vj},
		Data:      data,
	}
	g.edges = append(g.edges, edge)
	g.vertexes[vi].out = append(g.vertexes[vi].out, path{vj, e})
	if vi != vj {
		g.vertexes[vj].out = append(g.vertexes[vj].out, path{vi, e})
	}
	return e, nil
}

func (g Graph) GetAdjacencies(vertex int) ([]path, error) {
	if vertex < 0 || vertex >= len(g.vertexes) {
		return nil, fmt.Errorf("vertex index: %d is out of bounds", vertex)
	}
	return g.vertexes[vertex].out, nil
}

func (g Graph) GetVertex(index int) (*Vertex, error) {
	if index < 0 || index >= len(g.vertexes) {
		return nil, fmt.Errorf("vertex index: %d is out of bounds", index)
	}
	return &g.vertexes[index], nil
}

func (g Graph) GetEdge(index int) (*Edge, error) {
	if index < 0 || index >= len(g.vertexes) {
		return nil, fmt.Errorf("edge index: %d is out of bounds", index)
	}
	return &g.edges[index], nil
}
