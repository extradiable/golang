// Copyright 2023 The extradiable Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package datastructures

var threshold float32 = 0.3

// DStack: Dynamic Stack Structure
type DStack struct {
	//stack data
	data []interface{}
	//index to the top of the stack
	top int
}

// Creates a new dynamic stack
func CreateDStack() *DStack {
	stack := DStack{
		top:  0,
		data: make([]interface{}, 1),
	}
	return &stack
}

// Push value v into the top of the stack.
// Because the stack grows dinamically, the push operation can take O(n) in the worst case.
// Where n is the current size of the stack
func (s *DStack) Push(v interface{}) {
	if s.top < len(s.data) {
		s.data[s.top] = v
	} else {
		s.data = append(s.data, v)
	}
	s.top++
}

// Pop value v from the top of the stack and returns v.
// An error can be returned if Pop() is called on an empty queue.
// The stack is shrinked dinamically and so, the pop operation can take O(n) in the worst case.
func (s *DStack) Pop() (interface{}, error) {
	if s.top < 1 {
		return nil, ErrStackUnderflow
	}
	s.top--
	v := s.data[s.top]
	s.data[s.top] = nil
	if float32(s.top+1)/float32(cap(s.data)) <= threshold {
		tmp := make([]interface{}, s.top)
		copy(tmp, s.data)
		s.data = tmp
	}
	return v, nil
}

// returns true if the stack is currently empty
func (s *DStack) Empty() bool {
	return s.top == 0
}

// Returns the current size of the stack.
// The size of the stack reflects the number of elements currently stored in the stack
func (s *DStack) Size() int {
	return s.top
}
