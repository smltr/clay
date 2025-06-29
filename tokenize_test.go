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
