package graph

type vertex struct {
	id   int
	out  []out
	Data interface{}
}

func (v vertex) GetData() interface{} {
	return v.Data
}
