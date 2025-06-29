package main

import (
	"testing"
)

func setupTestEnv() *Environment {
	env := NewEnvironment(nil)
	setupBuiltins(env)
	return env
}

func TestBasicArithmetic(t *testing.T) {
	env := setupTestEnv()

	testCases := map[string]int{
		"plus(1, 2)":          3,
		"plus(1, 2, 3)":       6,
		"plus(5, 10)":         15,
		"plus(plus(1, 2), 3)": 6,
		"plus plus(1, 1) 2":   4,
		"plus 1 100":          101,
	}

	for code, expected := range testCases {
		tokens := tokenize(code)
		tree := parse(tokens)
		result := eval(tree, env)

		if result != expected {
			t.Errorf("For %s: expected %d, got %v", code, expected, result)
		}
	}
}

func TestVariableAssignment(t *testing.T) {
	env := setupTestEnv()

	// Test set
	code := "set(x, 42)"
	tokens := tokenize(code)
	tree := parse(tokens)
	result := eval(tree, env)

	if result != 42 {
		t.Errorf("Expected set to return 42, got %v", result)
	}

	// Test variable lookup
	code = "x"
	tokens = tokenize(code)
	tree = parse(tokens)
	result = eval(tree, env)

	if result != 42 {
		t.Errorf("Expected x to be 42, got %v", result)
	}

	// Test using variable in expression
	code = "plus(x, 8)"
	tokens = tokenize(code)
	tree = parse(tokens)
	result = eval(tree, env)

	if result != 50 {
		t.Errorf("Expected x plus 8 = 50, got %v", result)
	}
}

func TestFunctionDefinition(t *testing.T) {
	env := setupTestEnv()

	// Define a function: define myfunc(a, b) plus(a, b)
	code := `
	define myfunc(a, b)
		plus(a, b)`

	tokens := tokenize(code)
	tree := parse(tokens)
	eval(tree, env)

	// Call the function: myfunc(3, 4)
	code = "myfunc(3, 4)"
	tokens = tokenize(code)
	tree = parse(tokens)
	result := eval(tree, env)

	if result != 7 {
		t.Errorf("Expected myfunc(3, 4) = 7, got %v", result)
	}
}

func TestComplexFunction(t *testing.T) {
	env := setupTestEnv()

	// Define function with multiple statements
	code := `
	define complex(x, y)
		set temp plus(x, y)
		plus(temp, 1)`

	tokens := tokenize(code)
	tree := parse(tokens)
	eval(tree, env)

	// Call the function
	code = "complex(5, 10)"
	tokens = tokenize(code)
	tree = parse(tokens)
	result := eval(tree, env)

	if result != 16 {
		t.Errorf("Expected complex(5, 10) = 16, got %v", result)
	}
}

func TestMultipleExpressions(t *testing.T) {
	env := setupTestEnv()

	// Multiple expressions separated by newlines
	code := `set x 10
set y 20
plus(x, y)`

	tokens := tokenize(code)
	tree := parse(tokens)
	result := eval(tree, env)

	// Should return the result of the last expression
	if result != 30 {
		t.Errorf("Expected final result to be 30, got %v", result)
	}
}

func TestNumbers(t *testing.T) {
	env := setupTestEnv()

	testCases := map[string]int{
		"42":  42,
		"0":   0,
		"123": 123,
	}

	for code, expected := range testCases {
		tokens := tokenize(code)
		tree := parse(tokens)
		result := eval(tree, env)

		if result != expected {
			t.Errorf("For %s: expected %d, got %v", code, expected, result)
		}
	}
}

func TestUndefinedFunction(t *testing.T) {
	env := setupTestEnv()

	code := "unknown_func(1, 2)"
	tokens := tokenize(code)
	tree := parse(tokens)
	result := eval(tree, env)

	// Should return an error
	if _, ok := result.(error); !ok {
		t.Errorf("Expected error for undefined function, got %v", result)
	}
}
