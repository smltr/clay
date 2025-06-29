package main

import (
	"fmt"
	"strings"
)

// printExplicit not really a part of the main pipeline right now
// just want to be able to test out end to end visually
func printExplicit(node *Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case ITEM:
		// Function references and simple values - just print the name
		return node.Value

	case LIST:
		// Function calls - print as funcname(args...)
		if len(node.Children) == 0 {
			return node.Value + "()"
		}

		args := make([]string, len(node.Children))
		for i, child := range node.Children {
			args[i] = printExplicit(child)
		}
		return node.Value + "(" + strings.Join(args, ", ") + ")"

	default:
		return ""
	}
}

var code = `
define myfunc(a, b)
		plus(a, b)

myfunc 1 2
`

func pipe(s string) any {
	env := NewEnvironment(nil)
	setupBuiltins(env)

	return eval(parse(tokenize(s)), env)
}

func main() {
	res := pipe(code)
	fmt.Printf("%+v\n", res)
}
