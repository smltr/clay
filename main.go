package main

import (
	"fmt"
	"strings"
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

type TokenType int

const (
	WORD TokenType = iota
	INDENT
	DEDENT
	NEWLINE
	LPAREN
	RPAREN
	COMMA
	SPACE
	EOF
)

type Node struct {
	Type     NodeType
	Value    string
	Children []*Node
}

type NodeType int

const (
	NAME NodeType = iota
	CALL          // TODO merge this into LIST
	LIST
)

func tokenize(s string) []Token {
	var tokens []Token
	line := 0
	col := 0
	i := 0
	indentStack := []int{0}
	atLineStart := true

	for i < len(s) {
		char := s[i]

		// Check indentation at start of any line (tabs only)
		if atLineStart {
			indent := countTabIndentation(s, i)
			currentIndent := indentStack[len(indentStack)-1]

			if indent > currentIndent {
				// INDENT
				indentStack = append(indentStack, indent)
				tokens = append(tokens, Token{INDENT, "", line, 0})
			} else if indent < currentIndent {
				// DEDENT
				for len(indentStack) > 1 && indentStack[len(indentStack)-1] > indent {
					indentStack = indentStack[:len(indentStack)-1]
					tokens = append(tokens, Token{DEDENT, "", line, 0})
				}
			}

			// Skip the tab characters (but not spaces)
			for i < len(s) && s[i] == '\t' {
				col++
				i++
			}
			atLineStart = false
			continue
		}

		// Handle special characters
		switch char {
		case ' ':
			// Emit space as token instead of skipping
			tokens = append(tokens, Token{SPACE, " ", line, col})
			col++
			i++
		case '(':
			tokens = append(tokens, Token{LPAREN, "(", line, col})
			col++
			i++
		case ')':
			tokens = append(tokens, Token{RPAREN, ")", line, col})
			col++
			i++
		case ',':
			tokens = append(tokens, Token{COMMA, ",", line, col})
			col++
			i++
		case '\n':
			tokens = append(tokens, Token{NEWLINE, "\n", line, col})
			line++
			col = 0
			i++
			atLineStart = true
		case '\t':
			// Tabs outside of line start are treated as spaces for now
			tokens = append(tokens, Token{SPACE, "\t", line, col})
			col++
			i++
		default:
			// Handle words
			if isWordChar(char) {
				startCol := col
				word := ""
				for i < len(s) && isWordChar(s[i]) {
					word += string(s[i])
					col++
					i++
				}
				tokens = append(tokens, Token{WORD, word, line, startCol})
			} else {
				col++
				i++
			}
		}
	}

	// Add remaining DEDENTs at EOF
	for len(indentStack) > 1 {
		indentStack = indentStack[:len(indentStack)-1]
		tokens = append(tokens, Token{DEDENT, "", line, col})
	}

	tokens = append(tokens, Token{EOF, "", line, col})
	return tokens
}

func countTabIndentation(s string, start int) int {
	count := 0
	for i := start; i < len(s) && s[i] == '\t'; i++ {
		count++
	}
	return count
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

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

		// Check what immediately follows the word
		if *i >= len(tokens) || tokens[*i].Type == EOF || tokens[*i].Type == NEWLINE {
			// Bare word - this is a function reference, not a call
			return &Node{Type: NAME, Value: word, Children: nil}
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
			return &Node{Type: NAME, Value: word, Children: nil}
		}

	} else if tokens[*i].Type == LPAREN {
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
		Type:     CALL,
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
		Type:     CALL,
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
				Type:     NAME,
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
		Value:    "",
		Children: children,
	}
}

func printExplicit(node *Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case NAME:
		// Function references and simple values - just print the name
		return node.Value

	case CALL:
		// Function calls - print as funcname(args...)
		if len(node.Children) == 0 {
			return node.Value + "()"
		}

		args := make([]string, len(node.Children))
		for i, child := range node.Children {
			args[i] = printExplicit(child)
		}
		return node.Value + "(" + strings.Join(args, ", ") + ")"

	case LIST:
		// Lists - print as (item1, item2, ...)
		if len(node.Children) == 0 {
			return "()"
		}

		items := make([]string, len(node.Children))
		for i, child := range node.Children {
			items[i] = printExplicit(child)
		}
		return "(" + strings.Join(items, ", ") + ")"

	default:
		return ""
	}
}

var code = `
define myfunc(arg1, arg2)
	set sum add(arg1, arg2)
	print sum

print(hello)
`

func main() {
	pipe := func(s string) string {
		return printExplicit(parse(tokenize(s)))
	}

	fmt.Println(pipe(code))
}
