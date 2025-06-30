package main

import "fmt"

type Node struct {
	Type     NodeType
	Value    string
	Children []*Node
}

type NodeType int

const (
	ITEM NodeType = iota
	LIST
)

func parse(tokens []Token) *Node {
	if len(tokens) == 0 {
		return nil
	}

	i := 0
	expressions := []*Node{}

	// Parse all top-level expressions
	for i < len(tokens) && tokens[i].Type != EOF {
		// Skip leading newlines
		for i < len(tokens) && tokens[i].Type == NEWLINE {
			i++
		}

		if i >= len(tokens) || tokens[i].Type == EOF {
			break
		}

		expr := parseExpression(tokens, &i)
		if expr != nil {
			expressions = append(expressions, expr)
		}

		// Skip trailing newlines between expressions
		for i < len(tokens) && tokens[i].Type == NEWLINE {
			i++
		}
	}

	// If only one expression, return it directly
	if len(expressions) == 1 {
		return expressions[0]
	}

	// If multiple expressions, wrap in a list
	if len(expressions) > 1 {
		return &Node{
			Type:     LIST,
			Value:    "",
			Children: expressions,
		}
	}

	return nil
}

func parseExpression(tokens []Token, i *int) *Node {
	// Skip newlines
	for *i < len(tokens) && tokens[*i].Type == NEWLINE {
		(*i)++
	}

	if *i >= len(tokens) || tokens[*i].Type == EOF {
		return nil
	}

	if tokens[*i].Type == WORD {
		word := tokens[*i].Value
		(*i)++

		// Look ahead to decide parsing strategy
		if *i >= len(tokens) || tokens[*i].Type == EOF || tokens[*i].Type == NEWLINE {
			// Bare word
			return &Node{Type: ITEM, Value: word, Children: nil}
		}

		switch tokens[*i].Type {
		case LPAREN:
			// func(args) - existing logic
			return parseExplicitCall(tokens, word, i)
		case DO:
			// func do ... end
			return parseDoEndCall(tokens, word, i)
		case SPACE:
			// Check if it's "func args do" or just "func args"
			return parseSpaceCall(tokens, word, i)
		default:
			// Bare word
			return &Node{Type: ITEM, Value: word, Children: nil}
		}
	}

	return nil
}

func parseExplicitCall(tokens []Token, funcName string, i *int) *Node {
	(*i)++ // skip LPAREN

	children := []*Node{}

	for *i < len(tokens) && tokens[*i].Type != RPAREN {
		if tokens[*i].Type == COMMA {
			(*i)++ // skip comma
			continue
		}
		if tokens[*i].Type == SPACE {
			(*i)++ // skip space
			continue
		}

		// Parse each argument as an expression
		arg := parseArgument(tokens, i)
		if arg != nil {
			children = append(children, arg)
		}
	}

	if *i < len(tokens) && tokens[*i].Type == RPAREN {
		(*i)++ // skip RPAREN
	}

	return &Node{
		Type:     LIST,
		Value:    funcName,
		Children: children,
	}
}

func parseArgument(tokens []Token, i *int) *Node {
	if *i >= len(tokens) || tokens[*i].Type == EOF {
		return nil
	}

	if tokens[*i].Type == WORD {
		word := tokens[*i].Value
		(*i)++

		// Check if this word is followed by explicit parens
		if *i < len(tokens) && tokens[*i].Type == LPAREN {
			// This is a function call as an argument: otherfunc(arg)
			return parseExplicitCall(tokens, word, i)
		} else {
			// This is just a name
			return &Node{
				Type:     ITEM,
				Value:    word,
				Children: nil,
			}
		}
	} else if tokens[*i].Type == LPAREN {
		// This is a list argument
		return parseList(tokens, i)
	}

	return nil
}

func parseList(tokens []Token, i *int) *Node {
	(*i)++ // skip LPAREN

	children := []*Node{}

	for *i < len(tokens) && tokens[*i].Type != RPAREN {
		if tokens[*i].Type == COMMA {
			(*i)++ // skip comma
			continue
		}
		if tokens[*i].Type == SPACE {
			(*i)++ // skip space
			continue
		}

		arg := parseArgument(tokens, i)
		if arg != nil {
			children = append(children, arg)
		}
	}

	if *i < len(tokens) && tokens[*i].Type == RPAREN {
		(*i)++ // skip RPAREN
	}

	return &Node{
		Type:     LIST,
		Value:    "list",
		Children: children,
	}
}

func parseDoEndCall(tokens []Token, funcName string, i *int) *Node {
	fmt.Printf("parseDoEndCall: at token %d: %v %q\n", *i, tokens[*i].Type, tokens[*i].Value)
	(*i)++ // skip DO
	fmt.Printf("parseDoEndCall: after skipping DO, at token %d: %v %q\n", *i, tokens[*i].Type, tokens[*i].Value)

	blockArgs := parseDoBlock(tokens, i)

	return &Node{
		Type:     LIST,
		Value:    funcName,
		Children: blockArgs,
	}
}

func parseDoBlock(tokens []Token, i *int) []*Node {
	args := []*Node{}

	// Skip newline after DO
	if *i < len(tokens) && tokens[*i].Type == NEWLINE {
		(*i)++
	}

	// Parse until END
	for *i < len(tokens) && tokens[*i].Type != END && tokens[*i].Type != EOF {
		if tokens[*i].Type == NEWLINE {
			(*i)++
			continue
		}

		// Handle INDENT tokens FIRST, before calling parseExpression
		if tokens[*i].Type == INDENT {
			(*i)++ // skip INDENT

			groupedArgs := []*Node{}
			for *i < len(tokens) && tokens[*i].Type != DEDENT && tokens[*i].Type != END {
				if tokens[*i].Type == NEWLINE {
					(*i)++
					continue
				}
				expr := parseExpression(tokens, i)
				if expr != nil {
					groupedArgs = append(groupedArgs, expr)
				}
			}

			if *i < len(tokens) && tokens[*i].Type == DEDENT {
				(*i)++ // skip DEDENT
			}

			// Create a list node for the grouped args
			if len(groupedArgs) > 0 {
				args = append(args, &Node{
					Type:     LIST,
					Value:    "list",
					Children: groupedArgs,
				})
			}
		} else {
			// Only call parseExpression for non-INDENT tokens
			expr := parseExpression(tokens, i)
			if expr != nil {
				args = append(args, expr)
			}
		}
	}

	// Skip END token
	if *i < len(tokens) && tokens[*i].Type == END {
		(*i)++
	}

	return args
}

func parseSpaceCall(tokens []Token, funcName string, i *int) *Node {
	(*i)++ // skip the SPACE after function name

	children := []*Node{}

	// Check if next token is LPAREN (for func (list) case)
	if *i < len(tokens) && tokens[*i].Type == LPAREN {
		// func (arg1, arg2) - single list argument
		list := parseList(tokens, i)
		if list != nil {
			children = append(children, list)
		}
	} else {
		// func arg1 arg2 - multiple arguments on same line
		// Stop at NEWLINE, DO, or EOF
		for *i < len(tokens) && tokens[*i].Type != NEWLINE && tokens[*i].Type != DO && tokens[*i].Type != EOF {
			if tokens[*i].Type == SPACE {
				(*i)++ // skip spaces between arguments
				continue
			}

			arg := parseArgument(tokens, i)
			if arg != nil {
				children = append(children, arg)
			} else {
				break
			}
		}
	}

	// Check if followed by DO
	if *i < len(tokens) && tokens[*i].Type == DO {
		// func args do ... end
		(*i)++
		blockArgs := parseDoBlock(tokens, i)
		children = append(children, blockArgs...)
		return &Node{
			Type:     LIST,
			Value:    funcName,
			Children: children,
		}
	}

	// Handle indented arguments (existing logic)
	if *i < len(tokens) && tokens[*i].Type == NEWLINE {
		(*i)++ // skip newline

		if *i < len(tokens) && tokens[*i].Type == INDENT {
			(*i)++ // skip indent

			// Parse each indented line as a complete expression
			for *i < len(tokens) && tokens[*i].Type != DEDENT && tokens[*i].Type != EOF {
				if tokens[*i].Type == NEWLINE {
					(*i)++
					continue
				}

				// Parse the entire line as one expression
				expr := parseExpression(tokens, i)
				if expr != nil {
					children = append(children, expr)
				}
			}

			if *i < len(tokens) && tokens[*i].Type == DEDENT {
				(*i)++ // skip dedent
			}
		}
	}

	return &Node{
		Type:     LIST,
		Value:    funcName,
		Children: children,
	}
}
