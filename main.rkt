;; clay/main.rkt
#lang racket/base

;; Import more Racket modules to get more built-ins
(require racket/list
         racket/string
         racket/math
         racket/function)

;; Re-export ALL of Racket's functions EXCEPT the ones we provide ourselves
(provide (rename-out [module-begin #%module-begin])      ; Our custom module-begin
         (except-out (all-from-out racket/base)          ; Everything from Racket base...
                     #%module-begin                       ; ...except their module-begin
                     read-syntax)                         ; ...and their read-syntax
         (all-from-out racket/list)                       ; List functions
         (all-from-out racket/string)                     ; String functions
         (all-from-out racket/math)                       ; Math functions
         (all-from-out racket/function)                   ; Function utilities
)

;; This is required for any #lang - it sets up the module
(define-syntax-rule (module-begin body ...)
  (#%module-begin
   body ...))
