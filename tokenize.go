package main

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
