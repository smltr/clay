package main

import (
	"fmt"
	"strconv"
)

type Environment struct {
	bindings map[string]any
	parent   *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		bindings: make(map[string]any),
		parent:   parent,
	}
}

func (env *Environment) Get(name string) (any, bool) {
	if val, ok := env.bindings[name]; ok {
		return val, true
	}
	if env.parent != nil {
		return env.parent.Get(name)
	}
	return nil, false
}

func (env *Environment) Set(name string, value any) {
	env.bindings[name] = value
}

func setupBuiltins(env *Environment) {
	// naming 'add' now because '+' doesn't parse as a word
	env.Set("plus", func(args ...any) any {
		result := 0
		for _, arg := range args {
			if num, ok := arg.(int); ok {
				result += num
			}
		}
		return result
	})

	env.Set("print", func(args ...any) any {
		for i, arg := range args {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(arg)
		}
		fmt.Println()
		return nil
	})
}

func eval(node *Node, env *Environment) any {
	if node == nil {
		return nil
	}

	switch node.Type {
	case ITEM:
		// Try to parse as number first
		if num, err := strconv.Atoi(node.Value); err == nil {
			return num
		}

		// Look up in environment
		if val, ok := env.Get(node.Value); ok {
			return val
		}

		// Return as string literal if not found
		return node.Value

	case LIST:
		return evalList(node, env)
	}

	return nil
}

func evalList(node *Node, env *Environment) any {
	if len(node.Children) == 0 && node.Value == "" {
		// Empty list
		return nil
	}

	// Handle special forms first
	switch node.Value {
	case "define":
		return handleDefine(node, env)
	case "set":
		return handleSet(node, env)
	case "list":
		return handleList(node, env)
	default:
		// Regular function call
		return handleFunctionCall(node, env)
	}
}

func handleList(node *Node, env *Environment) interface{} {
	// Evaluate all children and return as a slice
	result := make([]interface{}, len(node.Children))
	for i, child := range node.Children {
		result[i] = eval(child, env)
	}
	return result
}

func handleDefine(node *Node, env *Environment) any {
	// define funcname(args) body...
	if len(node.Children) >= 2 {
		funcDef := node.Children[0] // This should be funcname(args)
		body := node.Children[1:]   // Rest is the body

		// Store function definition
		env.Set(funcDef.Value, map[string]any{
			"args": funcDef.Children,
			"body": body,
		})
	}
	return nil
}

func handleSet(node *Node, env *Environment) any {
	// set varname value
	if len(node.Children) >= 2 {
		varName := node.Children[0].Value
		value := eval(node.Children[1], env)
		env.Set(varName, value)
		return value
	}
	return nil
}

func handleFunctionCall(node *Node, env *Environment) any {
	funcName := node.Value
	fmt.Printf("dump: %+v", node)
	// Look up the function
	funcVal, ok := env.Get(funcName)
	if !ok {
		return fmt.Errorf("undefined function: %s", funcName)
	}

	// Evaluate all arguments
	args := make([]any, len(node.Children))
	for i, child := range node.Children {
		args[i] = eval(child, env)
	}

	// Check if it's a built-in function (Go function)
	if goFunc, ok := funcVal.(func(...any) any); ok {
		return goFunc(args...)
	}

	// Check if it's a user-defined function
	if funcDef, ok := funcVal.(map[string]any); ok {
		return callUserFunction(funcDef, args, env)
	}

	return fmt.Errorf("not a function: %s", funcName)
}

func callUserFunction(funcDef map[string]any, args []any, env *Environment) any {
	params := funcDef["args"].([]*Node)
	body := funcDef["body"].([]*Node)

	// Create new environment for function scope
	funcEnv := NewEnvironment(env)

	// Bind parameters to arguments
	for i, param := range params {
		if i < len(args) {
			funcEnv.Set(param.Value, args[i])
		}
	}

	// Evaluate function body
	var result any
	for _, stmt := range body {
		result = eval(stmt, funcEnv)
	}

	return result
}
