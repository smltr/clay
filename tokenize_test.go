package main

import (
	"fmt"
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

func TestTokenizeDoEnd(t *testing.T) {
	testCases := map[string][]Token{
		"do": {
			{DO, "do", 0, 0},
			{EOF, "", 0, 2},
		},
		"end": {
			{END, "end", 0, 0},
			{EOF, "", 0, 3},
		},
		"func do\narg\nend": {
			{WORD, "func", 0, 0},
			{SPACE, " ", 0, 4},
			{DO, "do", 0, 5},
			{NEWLINE, "\n", 0, 7},
			{WORD, "arg", 1, 0},
			{NEWLINE, "\n", 1, 3},
			{END, "end", 2, 0},
			{EOF, "", 2, 3},
		},
		"func do\n\targ1\n\targ2\nend": {
			{WORD, "func", 0, 0},
			{SPACE, " ", 0, 4},
			{DO, "do", 0, 5},
			{NEWLINE, "\n", 0, 7},
			{INDENT, "", 1, 0},
			{WORD, "arg1", 1, 1},
			{NEWLINE, "\n", 1, 5},
			{WORD, "arg2", 2, 1},
			{NEWLINE, "\n", 2, 5},
			{DEDENT, "", 3, 0},
			{END, "end", 3, 0},
			{EOF, "", 3, 3},
		},
		"func arg1 do\narg2\nend": {
			{WORD, "func", 0, 0},
			{SPACE, " ", 0, 4},
			{WORD, "arg1", 0, 5},
			{SPACE, " ", 0, 9},
			{DO, "do", 0, 10},
			{NEWLINE, "\n", 0, 12},
			{WORD, "arg2", 1, 0},
			{NEWLINE, "\n", 1, 4},
			{END, "end", 2, 0},
			{EOF, "", 2, 3},
		},
	}

	for code, expected := range testCases {
		result := tokenize(code)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("For input %q:\nExpected: %+v\nGot: %+v", code, expected, result)
		}
	}
}

func TestTokenizeSimple(t *testing.T) {
	code := "func do\narg\nend"
	tokens := tokenize(code)
	for i, token := range tokens {
		fmt.Printf("%d: %v %q\n", i, token.Type, token.Value)
	}
}
