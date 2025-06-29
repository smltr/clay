package main

import (
	"strings"
)

func transpile(node *Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case ITEM:
		return node.Value

	case LIST:
		if node.Value == "" {
			// Multiple expressions at top level
			parts := make([]string, len(node.Children))
			for i, child := range node.Children {
				parts[i] = transpile(child)
			}
			return strings.Join(parts, " ")
		} else {
			// Function call
			parts := []string{node.Value}
			for _, child := range node.Children {
				parts = append(parts, transpile(child))
			}
			return "(" + strings.Join(parts, " ") + ")"
		}
	}

	return ""
}
