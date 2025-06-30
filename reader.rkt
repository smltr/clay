;; clay/reader.rkt
#lang racket/base

(require racket/port racket/list racket/string)
(provide read-syntax)

(define (read-syntax src in)
  (define text (port->string in))
  (if (string=? text "")
      eof
      (let* ([tokens (tokenize text)]
             [clean-tokens (filter (lambda (t) (not (string=? t ""))) tokens)]
             [ast (parse clean-tokens)])
        (datum->syntax #f `(module clay-program "main.rkt" ,@ast)))))

;; Tokenizer with comment support
(define (tokenize text)
  (define tokens '())
  (define current-word "")
  (define chars (string->list text))

  (define (add-word!)
    (when (not (string=? current-word ""))
      (set! tokens (cons current-word tokens))
      (set! current-word "")))

  (let loop ([chars chars])
    (cond
      [(null? chars) (add-word!)]

      ;; Handle comments - skip everything until newline
      [(char=? (car chars) #\;)
       (add-word!)
       (loop (skip-to-newline chars))]

      ;; Handle strings
      [(char=? (car chars) #\")
       (add-word!)
       (define-values (str remaining) (read-string chars))
       (set! tokens (cons str tokens))
       (loop remaining)]

      ;; Handle special characters
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
      [(char=? (car chars) #\newline)
       (add-word!)
       (set! tokens (cons "NEWLINE" tokens))
       (loop (cdr chars))]
      [(char=? (car chars) #\tab)
       (add-word!)
       (set! tokens (cons "TAB" tokens))
       (loop (cdr chars))]

      ;; Skip spaces
      [(char=? (car chars) #\space)
       (add-word!)
       (loop (cdr chars))]

      ;; Regular characters
      [else
       (set! current-word (string-append current-word (string (car chars))))
       (loop (cdr chars))]))

  (reverse tokens))

;; Skip to newline for comments
(define (skip-to-newline chars)
  (cond
    [(null? chars) '()]
    [(char=? (car chars) #\newline) chars]
    [else (skip-to-newline (cdr chars))]))

;; Read string including quotes
(define (read-string chars)
  (let loop ([chars (cdr chars)] [result "\""])
    (cond
      [(null? chars) (values result '())]
      [(char=? (car chars) #\")
       (values (string-append result "\"") (cdr chars))]
      [else
       (loop (cdr chars) (string-append result (string (car chars))))])))

;; Main parser
(define (parse tokens)
  (define (skip-newlines tokens)
    (cond
      [(null? tokens) '()]
      [(string=? (first tokens) "NEWLINE") (skip-newlines (cdr tokens))]
      [else tokens]))

  (let loop ([tokens (skip-newlines tokens)] [result '()])
    (cond
      [(null? tokens) (reverse result)]

      ;; Skip newlines between expressions
      [(string=? (first tokens) "NEWLINE")
       (loop (skip-newlines (cdr tokens)) result)]

      ;; Parse one expression
      [else
       (define-values (expr remaining) (parse-one-expression tokens))
       (loop (skip-newlines remaining) (cons expr result))])))

;; Parse a single expression, including any indented block that follows
;; Parse a single expression, including any indented block that follows
(define (parse-one-expression tokens)
  (cond
    [(null? tokens) (values '() '())]

    ;; Function call with parentheses: func(args)
    [(and (>= (length tokens) 2)
          (string=? (second tokens) "("))
     (define func (parse-atom (first tokens)))
     (define-values (args remaining-after-parens) (parse-parenthesized (cddr tokens)))

     ;; Check for indented block after the parentheses
     (define-values (indented-block final-remaining) (parse-indented-block remaining-after-parens))

     (define all-args (append args indented-block))

     (values `(,func ,@all-args) final-remaining)]

    ;; Everything else: func arg1 arg2 ...
    [else
     (define line-tokens (take-until-newline tokens))
     (define after-line (drop-until-after-newline tokens))

     (if (null? line-tokens)
         (values '() after-line)
         (let* ([func (parse-atom (first line-tokens))]
                [line-args (parse-args (cdr line-tokens))])

           ;; Check for indented block after the line
           (define-values (indented-block remaining) (parse-indented-block after-line))

           (define all-args (append line-args indented-block))

           (values `(,func ,@all-args) remaining)))]))

(define (parse-indented-block tokens)
  (parse-indented-block-with-level tokens 1))

;; Parse indented block at a specific indentation level
(define (parse-indented-block-with-level tokens expected-level)
	(let loop ([tokens tokens] [block-content '()])
    (cond
      [(null? tokens)
       (if (null? block-content)
           (values '() '())
           ;; Always wrap indented blocks in begin
           (values (list `(begin ,@(reverse block-content))) '()))]

      ;; Check if this line starts with TABs
      [(string=? (first tokens) "TAB")
       (define-values (indent-level line-tokens remaining-tokens)
         (parse-line-with-indentation tokens))

       (cond
         ;; Line at our expected level - parse it and check for nested content
         [(= indent-level expected-level)
          (define-values (parsed-line final-remaining)
            (parse-line-with-possible-nesting line-tokens remaining-tokens expected-level))
          (loop final-remaining (cons parsed-line block-content))]

         ;; Line at deeper level - shouldn't happen at this point, but handle gracefully
         [(> indent-level expected-level)
          ;; This line belongs to the previous expression, so we're done with this block
          (values (reverse block-content) tokens)]

         ;; Line at shallower level - we're done with this block
         [else
          (values (reverse block-content) tokens)])]

      [else
             (if (null? block-content)
                 (values '() tokens)
                 ;; Always wrap indented blocks in begin
                 (values (list `(begin ,@(reverse block-content))) tokens))])))

(define (parse-line-with-indentation tokens)
  (let loop ([tokens tokens] [level 0])
    (cond
      [(null? tokens) (values level '() '())]
      [(string=? (first tokens) "TAB")
       (loop (cdr tokens) (+ level 1))]
      [else
       ;; Found non-TAB, get the rest of the line
       (define line-tokens (take-until-newline tokens))
       (define remaining (drop-until-after-newline tokens))
       (values level line-tokens remaining)])))

(define (parse-line-with-possible-nesting line-tokens remaining-tokens current-level)
  (cond
    [(null? line-tokens) (values '() remaining-tokens)]

    [else
     ;; Parse the current line as an expression
     (define-values (line-expr _unused) (parse-one-expression line-tokens))

     ;; Check if the next lines are more deeply indented (nested content)
     (define-values (nested-content final-remaining)
       (if (and (not (null? remaining-tokens))
                (string=? (first remaining-tokens) "TAB"))
           (let-values ([(next-level _line-tokens _remaining) (parse-line-with-indentation remaining-tokens)])
             (if (> next-level current-level)
                 ;; There is nested content
                 (parse-indented-block-with-level remaining-tokens (+ current-level 1))
                 ;; No nested content
                 (values '() remaining-tokens)))
           ;; No more tokens or no TAB, so no nested content
           (values '() remaining-tokens)))

     ;; Combine the line expression with any nested content
     (define final-expr
       (if (null? nested-content)
           line-expr
           (append (if (list? line-expr) line-expr (list line-expr)) nested-content)))

     (values final-expr final-remaining)]))

;; Skip all leading TAB tokens
(define (skip-leading-tabs tokens)
  (cond
    [(null? tokens) '()]
    [(string=? (first tokens) "TAB") (skip-leading-tabs (cdr tokens))]
    [else tokens]))
;; Parse arguments in parentheses
(define (parse-parenthesized tokens)
  (let loop ([tokens tokens] [args '()] [depth 1])
    (cond
      [(null? tokens) (values (reverse args) '())]

      [(string=? (first tokens) ")")
       (if (= depth 1)
           (values (reverse args) (cdr tokens))
           (loop (cdr tokens) (cons (parse-atom (first tokens)) args) (- depth 1)))]

      [(string=? (first tokens) "(")
       (loop (cdr tokens) (cons (parse-atom (first tokens)) args) (+ depth 1))]

      [(string=? (first tokens) ",")
       (loop (cdr tokens) args depth)]

      ;; Handle nested function calls - FIXED THE CONDITION
      [(and (>= (length tokens) 2)
            (string=? (second tokens) "("))
       (define-values (nested-expr remaining-tokens) (parse-one-expression tokens))
       (loop remaining-tokens (cons nested-expr args) depth)]

      [else
       (define arg (parse-atom (first tokens)))
       (loop (cdr tokens) (cons arg args) depth)])))

;; Parse whitespace-separated arguments
(define (parse-args tokens)
  (let loop ([tokens tokens] [args '()])
    (cond
      [(null? tokens) (reverse args)]

      ;; Function call with parens
      [(and (>= (length tokens) 2)
            (string=? (second tokens) "("))
       (define-values (expr remaining) (parse-one-expression tokens))
       (loop remaining (cons expr args))]

      ;; Parenthesized list: (arg1, arg2)
      [(string=? (first tokens) "(")
       (define-values (list-args remaining) (parse-parenthesized (cdr tokens)))
       (loop remaining (cons `(list ,@list-args) args))]

      ;; Regular argument
      [else
       (define arg (parse-atom (first tokens)))
       (loop (cdr tokens) (cons arg args))])))

;; Parse a single atom (string, number, or symbol)
(define (parse-atom token)
  (cond
    [(and (string? token)
          (> (string-length token) 1)
          (string-prefix? token "\"")
          (string-suffix? token "\""))
     (substring token 1 (- (string-length token) 1))]
    [(string->number token) (string->number token)]
    [else (string->symbol token)]))

;; Helper functions
(define (take-until-newline tokens)
  (cond
    [(null? tokens) '()]
    [(string=? (first tokens) "NEWLINE") '()]
    [else (cons (first tokens) (take-until-newline (cdr tokens)))]))

(define (drop-until-after-newline tokens)
  (cond
    [(null? tokens) '()]
    [(string=? (first tokens) "NEWLINE") (cdr tokens)]
    [else (drop-until-after-newline (cdr tokens))]))
