package main

import (
	"testing"
)

func TestTranspile(t *testing.T) {
	testCases := map[string]string{
		// Basic function calls
		"plus(1, 2)":    "(plus 1 2)",
		"plus(1, 2, 3)": "(plus 1 2 3)",
		"print(hello)":  "(print hello)",

		// Nested function calls
		"plus(plus(1, 2), 3)": "(plus (plus 1 2) 3)",

		// Variable assignment
		"set(x, 42)":          "(set x 42)",
		"set(y, plus(5, 10))": "(set y (plus 5 10))",

		// Function definitions
		"define(myfunc(a, b), plus(a, b))": "(define (myfunc a b) (plus a b))",

		// Function definitions with multiple body statements
		"define(complex(x, y), set(temp, plus(x, y)), plus(temp, 1))": "(define (complex x y) (set temp (plus x y)) (plus temp 1))",

		// Simple identifiers and numbers
		"42":    "42",
		"hello": "hello",
		"x":     "x",

		// Lists (data, not function calls)
		"list(1, 2, 3)": "(list 1 2 3)",

		// Implicit function calls (from whitespace syntax)
		"plus 1 2":    "(plus 1 2)",
		"print hello": "(print hello)",

		// Empty function call
		"print()": "(print)",
	}

	for input, expected := range testCases {
		tokens := tokenize(input)
		tree := parse(tokens)
		result := transpile(tree)

		if result != expected {
			t.Errorf("For input %q:\nExpected: %s\nGot: %s", input, expected, result)
		}
	}
}

func TestTranspileMultipleExpressions(t *testing.T) {
	testCases := map[string]string{
		// Multiple expressions - just output them sequentially
		"set(x, 10)\nset(y, 20)\nplus(x, y)": "(set x 10) (set y 20) (plus x y)",

		"print(hello)\nprint(world)": "(print hello) (print world)",
	}

	for input, expected := range testCases {
		tokens := tokenize(input)
		tree := parse(tokens)
		result := transpile(tree)

		if result != expected {
			t.Errorf("For input %q:\nExpected: %s\nGot: %s", input, expected, result)
		}
	}
}
