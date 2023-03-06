package datastructures

import (
	"strconv"
	"testing"
)

// TestEmptyStack:
// Verify that new stacks are empty
func TestEmptyStack(t *testing.T) {
	stack := CreateDStack()
	var boolExpected bool = true
	var boolResult bool = stack.Empty()
	if boolResult != boolExpected {
		t.Fatalf("stack.Empty(): expected value %s, got: %s", strconv.FormatBool(boolExpected), strconv.FormatBool(boolResult))
	}
}

// TestPushAndPop
// test a push followed by a pop
func TestPushAndPop(t *testing.T) {
	stack := CreateDStack()
	expectedValue := "hello"
	stack.Push(expectedValue)
	result, _ := stack.Pop()
	if expectedValue != result.(string) {
		t.Fatalf("stack.Pop(): expected value %s, got %v.", expectedValue, result)
	}
}

// TestUnderflow:
// Verifies stack underflow is returned apropiately
func TestUnderflow(t *testing.T) {
	stack := CreateDStack()
	iterations := 1024
	for i := 0; i < iterations; i++ {
		stack.Push(i)
	}
	for i := 0; i < iterations; i++ {
		_, err := stack.Pop()
		if err != nil {
			t.Fatalf("stack.Pop(): expected nil error got '%v'.", err)
		}
	}
	_, err := stack.Pop()
	if err != ErrStackUnderflow {
		t.Fatalf("stack.Pop(): expected '%v' error got '%v'.", ErrStackUnderflow, err)
	}
}
