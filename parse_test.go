package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := map[string]*Node{
		// Function references (NAME nodes)
		"otherfunc": {
			Type:     ITEM,
			Value:    "otherfunc",
			Children: nil,
		},

		// Function calls (LIST nodes)
		"otherfunc()": {
			Type:     LIST,
			Value:    "otherfunc",
			Children: []*Node{},
		},
		"otherfunc arg": {
			Type:  LIST,
			Value: "otherfunc",
			Children: []*Node{
				{Type: ITEM, Value: "arg", Children: nil},
			},
		},

		// Existing test cases that should still work
		"funcname(arg1)": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
			},
		},
		"funcname(arg1, arg2)": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{Type: ITEM, Value: "arg2", Children: nil},
			},
		},
		"funcname arg1 arg2 arg3": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{Type: ITEM, Value: "arg2", Children: nil},
				{Type: ITEM, Value: "arg3", Children: nil},
			},
		},
		"funcname arg1\n\targ2 arg3": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{Type: LIST, Value: "arg2", Children: []*Node{
					{Type: ITEM, Value: "arg3", Children: nil},
				}},
			},
		},
		"funcname otherfunc(arg)": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "otherfunc",
					Children: []*Node{
						{Type: ITEM, Value: "arg", Children: nil},
					},
				},
			},
		},
		"funcname (arg1, arg2)": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "list",
					Children: []*Node{
						{Type: ITEM, Value: "arg1", Children: nil},
						{Type: ITEM, Value: "arg2", Children: nil},
					},
				},
			},
		},
		"funcname((arg1, arg2))": {
			Type:  LIST,
			Value: "funcname",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "list",
					Children: []*Node{
						{Type: ITEM, Value: "arg1", Children: nil},
						{Type: ITEM, Value: "arg2", Children: nil},
					},
				},
			},
		},
	}

	for code, expected := range testCases {
		tokens := tokenize(code)
		result := parse(tokens)
		compareNodes(t, code, expected, result)
	}
}

func TestParseDoEnd(t *testing.T) {
	testCases := map[string]*Node{
		// functioncall do\narg1\narg2\nend → functioncall(arg1, arg2)
		"functioncall do\narg1\narg2\nend": {
			Type:  LIST,
			Value: "functioncall",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{Type: ITEM, Value: "arg2", Children: nil},
			},
		},

		// functioncall do\n  arg1\n  arg2\nend → functioncall((arg1, arg2))
		"functioncall do\n\targ1\n\targ2\nend": {
			Type:  LIST,
			Value: "functioncall",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "list",
					Children: []*Node{
						{Type: ITEM, Value: "arg1", Children: nil},
						{Type: ITEM, Value: "arg2", Children: nil},
					},
				},
			},
		},

		// functioncall do\narg1\n  arga\n  argb\nend → functioncall(arg1, (arga, argb))
		"functioncall do\narg1\n\targa\n\targb\nend": {
			Type:  LIST,
			Value: "functioncall",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{
					Type:  LIST,
					Value: "list",
					Children: []*Node{
						{Type: ITEM, Value: "arga", Children: nil},
						{Type: ITEM, Value: "argb", Children: nil},
					},
				},
			},
		},

		"functioncall do\narg1 do\n\targa\n\targb\nend\nend": {
			Type:  LIST,
			Value: "functioncall",
			Children: []*Node{
				{
					Type:  LIST,
					Value: "arg1",
					Children: []*Node{
						{
							Type:  LIST,
							Value: "list",
							Children: []*Node{
								{Type: ITEM, Value: "arga", Children: nil},
								{Type: ITEM, Value: "argb", Children: nil},
							},
						},
					},
				},
			},
		},

		// functioncall arg1 do\narg2\nend → functioncall(arg1, arg2)
		"functioncall arg1 do\narg2\nend": {
			Type:  LIST,
			Value: "functioncall",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{Type: ITEM, Value: "arg2", Children: nil},
			},
		},

		// Simple do block
		"print do\nhello\nend": {
			Type:  LIST,
			Value: "print",
			Children: []*Node{
				{Type: ITEM, Value: "hello", Children: nil},
			},
		},

		// Multiple same-line args before do
		"func arg1 arg2 do\narg3\narg4\nend": {
			Type:  LIST,
			Value: "func",
			Children: []*Node{
				{Type: ITEM, Value: "arg1", Children: nil},
				{Type: ITEM, Value: "arg2", Children: nil},
				{Type: ITEM, Value: "arg3", Children: nil},
				{Type: ITEM, Value: "arg4", Children: nil},
			},
		},

		// Empty do block
		"func do\nend": {
			Type:     LIST,
			Value:    "func",
			Children: []*Node{},
		},
	}

	for code, expected := range testCases {
		tokens := tokenize(code)
		result := parse(tokens)
		compareNodes(t, code, expected, result)
	}
}

// node helper funcs

func nodeToString(node *Node, indent int) string {
	if node == nil {
		return "nil"
	}

	spaces := strings.Repeat("  ", indent)
	result := fmt.Sprintf("%sNode{Type:%v, Value:%q", spaces, node.Type, node.Value)

	if len(node.Children) == 0 {
		result += ", Children:nil}"
	} else {
		result += ", Children:[\n"
		for i, child := range node.Children {
			result += nodeToString(child, indent+1)
			if i < len(node.Children)-1 {
				result += ","
			}
			result += "\n"
		}
		result += spaces + "]}"
	}

	return result
}

func compareNodes(t *testing.T, input string, expected, got *Node) {
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("For input %q:\nExpected:\n%s\nGot:\n%s",
			input,
			nodeToString(expected, 0),
			nodeToString(got, 0))
	}
}
