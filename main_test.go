package main

import "testing"

func TestPrintExplicit(t *testing.T) {
	testCases := map[string]string{
		// Function references (stay as references)
		"otherfunc": "otherfunc",

		// Function calls (zero args need explicit parens)
		"otherfunc()": "otherfunc()",

		// Implicit to explicit function calls
		"otherfunc arg":              "otherfunc(arg)",
		"funcname arg1 arg2 arg3":    "funcname(arg1, arg2, arg3)",
		"funcname arg1\n\targ2 arg3": "funcname(arg1, arg2(arg3))",

		// Already explicit calls (should stay the same)
		"funcname(arg1)":       "funcname(arg1)",
		"funcname(arg1, arg2)": "funcname(arg1, arg2)",

		// List cases - note the double parens for list arguments
		"funcname (arg1, arg2)":  "funcname((arg1, arg2))",
		"funcname((arg1, arg2))": "funcname((arg1, arg2))",

		// Nested function calls
		"funcname otherfunc(arg)": "funcname(otherfunc(arg))",

		// Mixed cases
		"funcname otherfunc":                  "funcname(otherfunc)", // otherfunc as reference passed to funcname
		"define myfunc\n\tprint hello\nother": "(define(myfunc, print(hello)), other)",
	}

	for input, expected := range testCases {
		tokens := tokenize(input)
		tree := parse(tokens)
		result := printExplicit(tree)

		if result != expected {
			t.Errorf("For input %q:\nExpected: %s\nGot: %s", input, expected, result)
		}
	}
}
