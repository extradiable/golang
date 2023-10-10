package graph

type edge struct {
	id        int
	endpoints []int
	Data      interface{}
}

func (e edge) GetData() interface{} {
	return e.Data
}
