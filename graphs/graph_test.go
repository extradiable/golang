package graph

import (
	"fmt"
	"testing"
)

func printPathLabels(vi, vj *Vertex, e *Edge) {
	if e == nil {
		fmt.Printf("%v\n", vi.Data)
	} else {
		fmt.Printf("%v ->%v %v\n", vi.Data, e.Data, vj.Data)
	}
}

// TestEmptyGraph:
// Verify that new stacks are empty
func TestEmptyStack(t *testing.T) {
	g := Graph{}
	for i := 0; i < 5; i++ {
		g.AddVertex(fmt.Sprintf("V%d", i+1))
	}
	g.AddEdge(0, 0, "E1")
	g.AddEdge(0, 1, "E2")
	g.AddEdge(0, 3, "E3")
	g.AddEdge(1, 4, "E4")
	g.AddEdge(4, 1, "E5")
	path, err := Trace(g, 0)
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = path.TraversePath(g, printPathLabels)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
