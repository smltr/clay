package main

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

	switch tokens[*i].Type {
	case WORD:
		word := tokens[*i].Value
		(*i)++

		// Check what immediately follows the word
		if *i >= len(tokens) || tokens[*i].Type == EOF || tokens[*i].Type == NEWLINE {
			// Bare word - this is a function reference, not a call
			return &Node{Type: ITEM, Value: word, Children: nil}
		}

		switch tokens[*i].Type {
		case LPAREN:
			// func(args) - explicit function call
			return parseExplicitCall(tokens, word, i)
		case SPACE:
			// func args or func (list) - implicit function call
			return parseImplicitCall(tokens, word, i)
		default:
			// Bare word - function reference
			return &Node{Type: ITEM, Value: word, Children: nil}
		}

	case LPAREN:
		// (args) - standalone list
		return parseList(tokens, i)
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

func parseImplicitCall(tokens []Token, funcName string, i *int) *Node {
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
		for *i < len(tokens) && tokens[*i].Type != NEWLINE && tokens[*i].Type != EOF {
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

	// Handle indented arguments
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

				// REMOVED: The "skip to end of line" logic that was interfering
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
