;; clay/reader.rkt
#lang racket/base

(require racket/port racket/list racket/string)
(provide read-syntax)

#|




CLAY LANGUAGE PARSER FLOW DIAGRAM
=================================

Input Text:
displayln(\"hello\")
define func()
	something

    ↓ tokenize

Tokens:
["displayln" "(" "\"hello\"" ")" "NEWLINE" "define" "func" "(" ")" "NEWLINE" "TAB" "something"]

    ↓ parse

AST (Abstract Syntax Tree):
[(displayln "hello") (define (func) (begin something))]

    ↓ wrap in module

Final Racket Module:
(module clay-program "main.rkt"
  (displayln "hello")
  (define (func) (begin something)))

Key Language Features:
- func(arg1, arg2)     → (func arg1 arg2)           [explicit function call]
- func arg1 arg2       → (func arg1 arg2)           [whitespace function call]
- func arg1            → (func arg1                 [line continues...]
  	nested                (begin nested))           [indented block]
- ; comment            → [ignored]                  [comments]
- "string"             → "string"                   [string literals]
|#

;; Main entry point - called by Racket's reader system
;; Takes source info and input port, returns syntax object
(define (read-syntax src in)
	;; Read all text from input port
	;; Example: "displayln(\"hello world\")" → "displayln(\"hello world\")"
	(define text (port->string in))

	;; Handle empty files
	(if (string=? text "")
		eof
		(let* ([tokens (tokenize text)]                    ; Break text into tokens
			[clean-tokens (filter (lambda (t)           ; Remove empty tokens
			(not (string=? t ""))) tokens)]
			[ast (parse clean-tokens)])                 ; Convert tokens to expressions
			;; Wrap everything in a Racket module structure
			;; This tells Racket "this is a module that uses our main.rkt"
			(datum->syntax #f `(module clay-program "main.rkt" ,@ast)))))

#|
TOKENIZER SECTION
================
The tokenizer breaks text into meaningful chunks (tokens).

Example transformation:
"func(\"hello\"); comment"
→ ["func" "(" "\"hello\"" ")" "; comment gets skipped"]
|#

;; Main tokenizer - converts string to list of tokens
;; Handles: strings, comments, parentheses, whitespace, indentation
(define (tokenize text)
	(define tokens '())          ; Accumulator for tokens (built in reverse)
	(define current-word "")     ; Current word being built character by character
	(define chars (string->list text))  ; Convert string to list of characters for easy processing

	;; Helper function to add completed word to tokens
	;; Example: current-word="hello" → tokens=("hello" ...previous-tokens...)
	(define (add-word!)
		(when (not (string=? current-word ""))
			(set! tokens (cons current-word tokens))
			(set! current-word "")))

	;; Main character-by-character loop
	(let loop ([chars chars])
		(cond
			;; End of input - add any remaining word
			[(null? chars) (add-word!)]

			;; COMMENT HANDLING: ';' starts a comment, skip everything until newline
			;; Example: "; this is a comment\nnext line" → skip to "next line"
			[(char=? (car chars) #\;)
				(add-word!)  ; Save any word we were building
				(loop (skip-to-newline chars))]  ; Jump to after the newline

			;; STRING HANDLING: '"' starts a string literal
			;; Example: "\"hello world\"" → "\"hello world\"" (kept as single token)
			[(char=? (car chars) #\")
				(add-word!)  ; Save any word we were building
				(define-values (str remaining) (read-string chars))
				(set! tokens (cons str tokens))  ; Add complete string as one token
				(loop remaining)]

			;; SPECIAL CHARACTERS: Each becomes its own token
			;; Example: "func()" → ["func" "(" ")"]
			[(char=? (car chars) #\()
				(add-word!)
				(set! tokens (cons "(" tokens))
				(loop (cdr chars))]

			[(char=? (car chars) #\))
				(add-word!)
				(set! tokens (cons ")" tokens))
				(loop (cdr chars))]

			[(char=? (car chars) #\,)
				(add-word!)
				(set! tokens (cons "," tokens))
				(loop (cdr chars))]

			;; WHITESPACE HANDLING: Convert to special tokens for parsing
			;; Example: "line1\n\tindented" → ["line1" "NEWLINE" "TAB" "indented"]
			[(char=? (car chars) #\newline)
				(add-word!)
				(set! tokens (cons "NEWLINE" tokens))
				(loop (cdr chars))]

			[(char=? (car chars) #\tab)
				(add-word!)
				(set! tokens (cons "TAB" tokens))
				(loop (cdr chars))]

			;; SPACE: Just separates words, doesn't become a token
			;; Example: "hello world" → ["hello" "world"] (space disappears)
			[(char=? (car chars) #\space)
				(add-word!)  ; Finish current word
				(loop (cdr chars))]  ; Continue without adding space token

			;; REGULAR CHARACTERS: Build up current word
			;; Example: chars=['h','e','l','l','o',' '] → current-word="hello"
			;
			[else
				(set! current-word (string-append current-word (string (car chars))))
				(loop (cdr chars))]))

	;; Return tokens in correct order (we built them backwards)
	(reverse tokens))

;; Helper: Skip everything until we hit a newline (for comments)
;; Example: [';' 'c' 'o' 'm' 'm' 'e' 'n' 't' '\n' 'n' 'e' 'x' 't']
;;       → ['\n' 'n' 'e' 'x' 't']
(define (skip-to-newline chars)
	(cond
		[(null? chars) '()]
		[(char=? (car chars) #\newline) chars]  ; Found newline, return from here
		[else (skip-to-newline (cdr chars))]))  ; Keep skipping

;; Helper: Read a complete string including the quotes
;; Example: ['"' 'h' 'i' '"' 'r' 'e' 's' 't'] → ("\"hi\"", ['r' 'e' 's' 't'])
(define (read-string chars)
	(let loop ([chars (cdr chars)] [result "\""])  ; Start with opening quote
		(cond
			[(null? chars) (values result '())]  ; Unexpected end
			[(char=? (car chars) #\")
				;; Found closing quote - add it and return
				(values (string-append result "\"") (cdr chars))]
			[else
				;; Regular character in string - add to result
				(loop (cdr chars) (string-append result (string (car chars))))])))

#|
PARSER SECTION
=============
The parser converts tokens into Racket expressions (AST nodes).

Key transformations:
- ["func" "(" "arg" ")"] → (func arg)
- ["func" "arg1" "arg2"] → (func arg1 arg2)
- ["func" "NEWLINE" "TAB" "nested"] → (func (begin nested))
|#

;; Main parser - converts tokens to Abstract Syntax Tree
;; Returns list of expressions ready for Racket
(define (parse tokens)
	;; Helper: Skip over NEWLINE tokens (they're just noise between expressions)
	;; Example: ["NEWLINE" "NEWLINE" "func" "arg"] → ["func" "arg"]
	(define (skip-newlines tokens)
		(cond
			[(null? tokens) '()]
			[(string=? (first tokens) "NEWLINE") (skip-newlines (cdr tokens))]
			[else tokens]))

	;; Main parsing loop - process each top-level expression
	(let loop ([tokens (skip-newlines tokens)] [result '()])
		(cond
			[(null? tokens) (reverse result)]  ; Done - return expressions in order

			;; Skip newlines between expressions
			;; Example: ["expr1" "NEWLINE" "NEWLINE" "expr2"] → handle both expressions
			[(string=? (first tokens) "NEWLINE")
				(loop (skip-newlines (cdr tokens)) result)]

			;; Parse one complete expression (might include indented block)
			[else
				(define-values (expr remaining) (parse-one-expression tokens))
				(loop (skip-newlines remaining) (cons expr result))])))

;; Parse a single expression, including any indented block that follows
;; This handles both function calls and indented blocks
(define (parse-one-expression tokens)
	(cond
		[(null? tokens) (values '() '())]

		;; EXPLICIT FUNCTION CALL: func(args)
		;; Example: ["displayln" "(" "hello" ")"] → (displayln hello)
		[(and (>= (length tokens) 2) (string=? (second tokens) "("))
			(define func (parse-atom (first tokens)))  ; Get function name
			;; Parse everything between the parentheses
			(define-values (args remaining-after-parens) (parse-parenthesized (cddr tokens)))

			;; Check if there's an indented block after the parentheses
			;; Example: func()\n\tblock → (func (begin block))
			(define-values (indented-block final-remaining) (parse-indented-block remaining-after-parens))

			(define all-args (append args indented-block))
				(values `(,func ,@all-args) final-remaining)]

		;; WHITESPACE FUNCTION CALL: func arg1 arg2
		;; Example: ["displayln" "hello" "world"] → (displayln hello world)
		[else
		;; Get everything on this line
			(define line-tokens (take-until-newline tokens))
			(define after-line (drop-until-after-newline tokens))

			(if (null? line-tokens)
				(values '() after-line)
				(let* ([func (parse-atom (first line-tokens))]      ; First token is function
					[line-args (parse-args (cdr line-tokens))])  ; Rest are arguments

					;; Check for indented block after the line
					;; Example: func arg1\n\tblock → (func arg1 (begin block))
					(define-values (indented-block remaining) (parse-indented-block after-line))

					(define all-args (append line-args indented-block))
					(values `(,func ,@all-args) remaining)))]))

;; Parse indented block (starts at indentation level 1)
;; Example: ["TAB" "expr1" "NEWLINE" "TAB" "expr2"] → [(begin expr1 expr2)]
; (define (parse-indented-block tokens)
; 	(parse-indented-block-with-level tokens 1))

;; Parse indented block at specific indentation level
;; This handles nested indentation correctly
; (define (parse-indented-block-with-level tokens expected-level)
; 	(let loop ([tokens tokens] [block-content '()])
; 		(cond
; 			[(null? tokens)
; 			;; End of input - wrap any content in begin block
; 				(if (null? block-content)
; 					(values '() '())
; 					(values (list `(begin ,@(reverse block-content))) '()))]

; 			;; Check if this line is indented (starts with TAB)
; 			[(string=? (first tokens) "TAB")
; 			;; Count indentation level and get line content
; 				(define-values (indent-level line-tokens remaining-tokens)
; 					(parse-line-with-indentation tokens))

; 				(cond
; 				;; Line at our expected level - parse it
; 				[(= indent-level expected-level)
; 					(define-values (parsed-line final-remaining)
; 					(parse-line-with-possible-nesting line-tokens remaining-tokens expected-level))
; 					(loop final-remaining (cons parsed-line block-content))]

; 				;; Line deeper than expected - belongs to previous expression
; 				[(> indent-level expected-level)
; 					(values (reverse block-content) tokens)]

; 				;; Line less indented - we're done with this block
; 				[else
; 					(values (reverse block-content) tokens)])]

; 			;; No indentation - end of block
; 			[else
; 				(if (null? block-content)
; 				(values '() tokens)
; 				(values (list `(begin ,@(reverse block-content))) tokens))])))

;; Parse a line and count its indentation level
;; Example: ["TAB" "TAB" "expr"] → (2, ["expr"], remaining-tokens)
; (define (parse-line-with-indentation tokens)
; 	(let loop ([tokens tokens] [level 0])
; 		(cond
; 			[(null? tokens) (values level '() '())]
; 			[(string=? (first tokens) "TAB")
; 				(loop (cdr tokens) (+ level 1))]  ; Count tabs
; 			[else
; 			;; Found non-TAB, get the rest of the line
; 				(define line-tokens (take-until-newline tokens))
; 				(define remaining (drop-until-after-newline tokens))
; 				(values level line-tokens remaining)])))

;; Parse a line that might have nested content after it
;; Example: line "func arg" followed by more indented lines
; (define (parse-line-with-possible-nesting line-tokens remaining-tokens current-level)
; 	(cond
; 		[(null? line-tokens) (values '() remaining-tokens)]

; 		[else
; 		;; Parse the current line as an expression
; 			(define-values (line-expr _unused) (parse-one-expression line-tokens))

; 			;; Check if the next lines are more deeply indented (nested content)
; 			(define-values (nested-content final-remaining)
; 				(if (and (not (null? remaining-tokens))
; 					(string=? (first remaining-tokens) "TAB"))
; 					(let-values ([(next-level _line-tokens _remaining)
; 						(parse-line-with-indentation remaining-tokens)])
; 						(if (> next-level current-level)
; 							;; There is nested content - parse it
; 							(parse-indented-block-with-level remaining-tokens (+ current-level 1))
; 							;; No nested content
; 							(values '() remaining-tokens)))
; 					;; No more tokens or no TAB, so no nested content
; 					(values '() remaining-tokens)))

; 			;; Combine the line expression with any nested content
; 			(define final-expr
; 				(if (null? nested-content)
; 					line-expr
; 					(append (if (list? line-expr) line-expr (list line-expr)) nested-content)))

; 			(values final-expr final-remaining)]))

;; Skip all leading TAB tokens (helper function)
; (define (skip-leading-tabs tokens)
; 	(cond
; 		[(null? tokens) '()]
; 		[(string=? (first tokens) "TAB") (skip-leading-tabs (cdr tokens))]
; 		[else tokens]))

;; Parse arguments inside parentheses: (arg1, arg2, arg3)
;; Example: ["arg1" "," "arg2" ")"] → ([arg1 arg2], [remaining-tokens])
(define (parse-parenthesized tokens)
	(let loop ([tokens tokens] [args '()] [depth 1])
		(cond
			[(null? tokens) (values (reverse args) '())]

			;; Found closing paren at our depth level
			[(string=? (first tokens) ")")
			   (if (= depth 1)
					(values (reverse args) (cdr tokens))  ; Done with this level
					(loop (cdr tokens) (cons (parse-atom (first tokens)) args) (- depth 1)))]

			;; Found opening paren - increase depth
			[(string=? (first tokens) "(")
				(loop (cdr tokens) (cons (parse-atom (first tokens)) args) (+ depth 1))]

			;; Skip commas (they just separate arguments)
			[(string=? (first tokens) ",")
				(loop (cdr tokens) args depth)]

			;; Handle nested function calls inside parentheses
			;; Example: outer(inner(arg))
			[(and (>= (length tokens) 2) (string=? (second tokens) "("))
				(define-values (nested-expr remaining-tokens) (parse-one-expression tokens))
				(loop remaining-tokens (cons nested-expr args) depth)]

			;; Regular argument
			[else
				(define arg (parse-atom (first tokens)))
				(loop (cdr tokens) (cons arg args) depth)])))

;; Parse whitespace-separated arguments: func arg1 arg2 arg3
;; Example: ["arg1" "arg2" "(" "list-arg" ")"] → [arg1 arg2 (list list-arg)]
(define (parse-args tokens)
	(let loop ([tokens tokens] [args '()])
		(cond
			[(null? tokens) (reverse args)]

			;; Function call with parens inside arguments
			;; Example: func arg1 inner(arg2) arg3
			[(and (>= (length tokens) 2) (string=? (second tokens) "("))
				(define-values (expr remaining) (parse-one-expression tokens))
				(loop remaining (cons expr args))]

			;; Parenthesized list: (arg1, arg2) becomes (list arg1 arg2)
			[(string=? (first tokens) "(")
				(define-values (list-args remaining) (parse-parenthesized (cdr tokens)))
				(loop remaining (cons `(list ,@list-args) args))]

			;; Regular argument
			[else
				(define arg (parse-atom (first tokens)))
				(loop (cdr tokens) (cons arg args))])))

;; Parse a single atom (string, number, or symbol)
;; Example: "\"hello\"" → "hello", "42" → 42, "func" → 'func
(define (parse-atom token)
	(cond
	;; String literal: "hello" → "hello" (remove quotes)
		[(and (string? token) (> (string-length token) 1)
			(string-prefix? token "\"")
			(string-suffix? token "\""))
			(substring token 1 (- (string-length token) 1))]
		;; Number: "42" → 42
		[(string->number token) (string->number token)]
		;; Symbol: "func" → 'func
		[else (string->symbol token)]))

#|
HELPER FUNCTIONS
===============
Utilities for working with token lists
|#

;; Get all tokens until we hit a NEWLINE
;; Example: ["func" "arg" "NEWLINE" "other"] → ["func" "arg"]
(define (take-until-newline tokens)
	(cond
		[(null? tokens) '()]
		[(string=? (first tokens) "NEWLINE") '()]
		[else (cons (first tokens) (take-until-newline (cdr tokens)))]))

;; Skip tokens until after the NEWLINE
;; Example: ["func" "arg" "NEWLINE" "other"] → ["other"]
(define (drop-until-after-newline tokens)
	(cond
		[(null? tokens) '()]
		[(string=? (first tokens) "NEWLINE") (cdr tokens)]
		[else (drop-until-after-newline (cdr tokens))]))
