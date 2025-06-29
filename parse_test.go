package main

import (
	"reflect"
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
					Value: "",
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
					Value: "",
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
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("For input %q:\nExpected: %+v\nGot: %+v", code, expected, result)
		}
	}
}
