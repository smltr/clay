;; rubber/lang/reader.rkt
#lang racket/base

(require racket/port racket/list)
(provide read-syntax)

(define (read-syntax src in)
  (displayln "READ-SYNTAX CALLED!")
  (define text (port->string in))
  (displayln (format "Read text: ~v" text))

  (if (string=? text "")
      eof
      (let* ([tokens (tokenize text)]
             [ast (parse tokens)])
        (displayln (format "Tokens: ~v" tokens))
        (displayln (format "Parsed AST: ~v" ast))
        (datum->syntax #f `(module rubber-program "main.rkt" ,@ast)))))

;; Enhanced tokenizer that handles newlines as separate tokens
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

      ;; Handle strings - read until closing quote
      [(char=? (car chars) #\")
       (add-word!)
       (define str (read-string-token chars))
       (set! tokens (cons str tokens))
       (loop (skip-string-token chars))]

      ;; Handle newlines specially
      [(char=? (car chars) #\newline)
       (add-word!)
       (set! tokens (cons "NEWLINE" tokens))
       (loop (cdr chars))]

      ;; Handle other whitespace (spaces, tabs)
      [(char-whitespace? (car chars))
       (add-word!)
       (loop (cdr chars))]

      ;; Handle parentheses
      [(char=? (car chars) #\()
       (add-word!)
       (set! tokens (cons "(" tokens))
       (loop (cdr chars))]
      [(char=? (car chars) #\))
       (add-word!)
       (set! tokens (cons ")" tokens))
       (loop (cdr chars))]

      ;; Handle commas
      [(char=? (car chars) #\,)
       (add-word!)
       (set! tokens (cons "," tokens))
       (loop (cdr chars))]

      ;; Regular character - add to current word
      [else
       (set! current-word (string-append current-word (string (car chars))))
       (loop (cdr chars))]))

  (filter (lambda (s) (not (string=? s ""))) (reverse tokens)))

;; [keep the string token functions the same]
(define (read-string-token chars)
  (let loop ([chars chars] [result ""])
    (cond
      [(null? chars) result]
      [(char=? (car chars) #\")
       (if (string=? result "")
           (loop (cdr chars) "\"")
           (string-append result "\""))]
      [else
       (loop (cdr chars) (string-append result (string (car chars))))])))

(define (skip-string-token chars)
  (let loop ([chars chars] [in-string? #f])
    (cond
      [(null? chars) '()]
      [(char=? (car chars) #\")
       (if in-string?
           (cdr chars)
           (loop (cdr chars) #t))]
      [in-string?
       (loop (cdr chars) #t)]
      [else chars])))

;; Enhanced parser that stops at newlines
(define (parse tokens)
  ;; First, remove empty newlines at start
  (define clean-tokens (skip-newlines tokens))

  (let loop ([tokens clean-tokens] [result '()])
    (cond
      [(null? tokens) (reverse result)]

      ;; Skip newlines between expressions
      [(string=? (first tokens) "NEWLINE")
       (loop (skip-newlines (cdr tokens)) result)]

      ;; Handle mixed syntax: func arg1(args) arg2
      [(and (>= (length tokens) 3)
            (not (string=? (second tokens) "("))
            (member "(" (take-until-newline (cdr tokens))))
       (define func (string->symbol (first tokens)))
       (define line-tokens (take-until-newline (cdr tokens)))
       (define-values (args remaining-in-line) (parse-whitespace-args line-tokens))
       (define remaining-after-line (drop-until-after-newline (cdr tokens)))
       (loop remaining-after-line (cons `(,func ,@args) result))]

      ;; Handle explicit parentheses: func(args)
      [(and (>= (length tokens) 2)
            (string=? (second tokens) "("))
       (define func (string->symbol (first tokens)))
       (define-values (args remaining) (parse-paren-args (cddr tokens)))
       (loop remaining (cons `(,func ,@args) result))]

      ;; Handle simple whitespace calls: func arg1 arg2
      [(>= (length tokens) 2)
       (define func (string->symbol (first tokens)))
       (define line-tokens (take-until-newline (cdr tokens)))
       (define args (map parse-single-token line-tokens))
       (define remaining (drop-until-after-newline (cdr tokens)))
       (loop remaining (cons `(,func ,@args) result))]

      ;; Single token
      [else
       (define token (parse-single-token (first tokens)))
       (define remaining (drop-until-after-newline (cdr tokens)))
       (loop remaining (cons token result))])))

;; Helper functions for handling newlines
(define (skip-newlines tokens)
  (cond
    [(null? tokens) '()]
    [(string=? (first tokens) "NEWLINE") (skip-newlines (cdr tokens))]
    [else tokens]))

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

;; [keep the other parsing functions the same]
(define (parse-whitespace-args tokens)
  (let loop ([tokens tokens] [args '()])
    (cond
      [(null? tokens) (values (reverse args) '())]

      ;; Handle function call: func(args)
      [(and (>= (length tokens) 2)
            (string=? (second tokens) "("))
       (define func (string->symbol (first tokens)))
       (define-values (nested-args remaining) (parse-paren-args (cddr tokens)))
       (loop remaining (cons `(,func ,@nested-args) args))]

      ;; Handle single token
      [else
       (define arg (parse-single-token (first tokens)))
       (loop (cdr tokens) (cons arg args))])))

(define (parse-paren-args tokens)
  (let loop ([tokens tokens] [args '()])
    (cond
      [(null? tokens) (values (reverse args) '())]
      [(string=? (first tokens) ")")
       (values (reverse args) (cdr tokens))]

      ;; Skip commas and newlines - they're just separators/formatting
      [(or (string=? (first tokens) ",")
           (string=? (first tokens) "NEWLINE"))
       (loop (cdr tokens) args)]

      ;; Handle nested function call: func(...)
      [(and (>= (length tokens) 2)
            (string=? (second tokens) "("))
       (define func (string->symbol (first tokens)))
       (define-values (nested-args remaining) (parse-paren-args (cddr tokens)))
       (loop remaining (cons `(,func ,@nested-args) args))]

      ;; Handle single token
      [else
       (define arg (parse-single-token (first tokens)))
       (loop (cdr tokens) (cons arg args))])))


(define (parse-single-token token)
  (cond
    [(and (string? token)
          (> (string-length token) 1)
          (char=? (string-ref token 0) #\")
          (char=? (string-ref token (- (string-length token) 1)) #\"))
     (substring token 1 (- (string-length token) 1))]
    [(string->number token) (string->number token)]
    [else (string->symbol token)]))
