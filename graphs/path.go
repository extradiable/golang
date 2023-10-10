package graph

import "fmt"

type out struct {
	Endpoint int
	Edge     int
}

type Path struct {
	graph *Graph
	start int
	out   []out
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

func (p Path) GetLastVertex() int {
	if len(p.out) == 0 {
		return p.start
	}
	return p.out[len(p.out)-1].Endpoint
}

func (p Path) Grow(vertexIndex, edgeIndex int) (Path, error) {
	if p.graph == nil {
		return p, ErrNoGraph
	}
	if err := p.graph.testVertex(vertexIndex); err != nil {
		return p, err
	}
	if err := p.graph.testEdge(edgeIndex); err != nil {
		return p, err
	}

	endpoints := p.graph.edges[edgeIndex].endpoints
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

func printPath(vi, vj *vertex, e *edge) {
	if e == nil {
		fmt.Printf("v%d\n", vi.id)
	} else {
		fmt.Printf("v%d ->e%d v%d\n", vi.id, e.id, vj.id)
	}
}

func (p Path) PrintPath() {
	p.TraversePath(printPath)
}

func (p Path) TraversePath(fn func(vi, vj *vertex, e *edge)) error {
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
