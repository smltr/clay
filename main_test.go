package main

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	testCases := map[string][]Token{
		"funcname(arg1)": {
			{WORD, "funcname", 0, 0},
			{LPAREN, "(", 0, 8},
			{WORD, "arg1", 0, 9},
			{RPAREN, ")", 0, 13},
			{EOF, "", 0, 14},
		},
		"funcname(arg1, arg2)": {
			{WORD, "funcname", 0, 0},
			{LPAREN, "(", 0, 8},
			{WORD, "arg1", 0, 9},
			{COMMA, ",", 0, 13},
			{SPACE, " ", 0, 14},
			{WORD, "arg2", 0, 15},
			{RPAREN, ")", 0, 19},
			{EOF, "", 0, 20},
		},
		"funcname arg1 arg2 arg3": {
			{WORD, "funcname", 0, 0},
			{SPACE, " ", 0, 8},
			{WORD, "arg1", 0, 9},
			{SPACE, " ", 0, 13},
			{WORD, "arg2", 0, 14},
			{SPACE, " ", 0, 18},
			{WORD, "arg3", 0, 19},
			{EOF, "", 0, 23},
		},
		"funcname (arg1, arg2)": {
			{WORD, "funcname", 0, 0},
			{SPACE, " ", 0, 8},
			{LPAREN, "(", 0, 9},
			{WORD, "arg1", 0, 10},
			{COMMA, ",", 0, 14},
			{SPACE, " ", 0, 15},
			{WORD, "arg2", 0, 16},
			{RPAREN, ")", 0, 20},
			{EOF, "", 0, 21},
		},
		"funcname arg1\n\targ2 arg3\notherfunc\n": {
			{WORD, "funcname", 0, 0},
			{SPACE, " ", 0, 8},
			{WORD, "arg1", 0, 9},
			{NEWLINE, "\n", 0, 13},
			{INDENT, "", 1, 0},
			{WORD, "arg2", 1, 1},
			{SPACE, " ", 1, 5},
			{WORD, "arg3", 1, 6},
			{NEWLINE, "\n", 1, 10},
			{DEDENT, "", 2, 0},
			{WORD, "otherfunc", 2, 0},
			{NEWLINE, "\n", 2, 9},
			{EOF, "", 3, 0},
		},
	}

	for code, expected := range testCases {
		result := tokenize(code)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("For input %q:\nExpected: %+v\nGot: %+v", code, expected, result)
		}
	}
}

func TestParse(t *testing.T) {
	testCases := map[string]*Node{
		// Function references (NAME nodes)
		"otherfunc": {
			Type:     NAME,
			Value:    "otherfunc",
			Children: nil,
		},

		// Function calls (CALL nodes)
		"otherfunc()": {
			Type:     CALL,
			Value:    "otherfunc",
			Children: []*Node{},
		},
		"otherfunc arg": {
			Type:  CALL,
			Value: "otherfunc",
			Children: []*Node{
				{Type: NAME, Value: "arg", Children: nil},
			},
		},

		// Existing test cases that should still work
		"funcname(arg1)": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{Type: NAME, Value: "arg1", Children: nil},
			},
		},
		"funcname(arg1, arg2)": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{Type: NAME, Value: "arg1", Children: nil},
				{Type: NAME, Value: "arg2", Children: nil},
			},
		},
		"funcname arg1 arg2 arg3": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{Type: NAME, Value: "arg1", Children: nil},
				{Type: NAME, Value: "arg2", Children: nil},
				{Type: NAME, Value: "arg3", Children: nil},
			},
		},
		"funcname arg1\n\targ2 arg3": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{Type: NAME, Value: "arg1", Children: nil},
				{Type: NAME, Value: "arg2", Children: nil},
				{Type: NAME, Value: "arg3", Children: nil},
			},
		},
		"funcname otherfunc(arg)": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{
					Type:  CALL,
					Value: "otherfunc",
					Children: []*Node{
						{Type: NAME, Value: "arg", Children: nil},
					},
				},
			},
		},
		"funcname (arg1, arg2)": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "",
					Children: []*Node{
						{Type: NAME, Value: "arg1", Children: nil},
						{Type: NAME, Value: "arg2", Children: nil},
					},
				},
			},
		},
		"funcname((arg1, arg2))": {
			Type:  CALL,
			Value: "funcname",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "",
					Children: []*Node{
						{Type: NAME, Value: "arg1", Children: nil},
						{Type: NAME, Value: "arg2", Children: nil},
					},
				},
			},
		},
	}

	for code, expected := range testCases {
		tokens := tokenize(code)
		result := parse(tokens)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("For input %q:\nExpected: %+v\nGot: %+v", code, expected, result)
		}
	}
}

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
