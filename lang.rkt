#lang racket/base

(require "main.rkt"
         (only-in "reader.rkt" read-syntax))

(provide (all-from-out "main.rkt")
         read-syntax)
