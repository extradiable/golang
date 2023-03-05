package datastructures

import (
	"strconv"
	"testing"
)

// TestNewStack verifies that new Stacks are empty
func TestNewStack(t *testing.T) {
	stack := CreateDStack()
	var boolExpected bool = true
	var boolResult bool = stack.Empty()
	if boolResult != boolExpected {
		t.Fatalf("[step 1] stack.Empty(): expected value %s, got: %s", strconv.FormatBool(boolExpected), strconv.FormatBool(boolResult))
	}
	var intExpected int = 0
	var intResult int = stack.Size()
	if intResult != intExpected {
		t.Fatalf("[step 2] stack.Size(): expected value %d, got %d.", intExpected, intResult)
	}
	_, err := stack.Pop()
	if err == nil {
		t.Fatal("[step 3] stack.Pop(): expected error: stack is empty, got: nil")
	}
	stack.Push(1)
	boolExpected = false
	boolResult = stack.Empty()
	if boolResult != boolExpected {
		t.Fatalf("[step 4] stack.Empty(): expected value %s, got: %s", strconv.FormatBool(boolExpected), strconv.FormatBool(boolResult))
	}
	intExpected = 1
	intResult = stack.Size()
	if intResult != intExpected {
		t.Fatalf("[step 5] stack.Size(): expected value %d, got %d.", intExpected, intResult)
	}
	stack.Pop()
	boolExpected = true
	boolResult = stack.Empty()
	if boolResult != boolExpected {
		t.Fatalf("[step 6] stack.Empty(): expected value %s, got: %s", strconv.FormatBool(boolExpected), strconv.FormatBool(boolResult))
	}
	intExpected = 0
	intResult = stack.Size()
	if intResult != intExpected {
		t.Fatalf("[step 7] stack.Size(): expected value %d, got %d.", intExpected, intResult)
	}
}

func TestMultiplePush(t *testing.T) {
	stack := CreateDStack()
	var pushes int = 17
	var pops int = 12
	for i := 1; i <= pushes; i++ {
		stack.Push(i * 10)
	}
	for i := 1; i <= pops; i++ {
		stack.Pop()
	}
	intExpected := pushes - pops
	intResult := stack.Size()
	if intResult != intExpected {
		t.Fatalf("[step 1] stack.Size(): expected value %d, got %d.", intExpected, intResult)
	}
	pushes = 120
	for i := 1; i <= pushes; i++ {
		stack.Push(i * 10)
	}
	pops = 90
	for i := 1; i <= pops; i++ {
		stack.Pop()
	}
}
