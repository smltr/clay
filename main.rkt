;; rubber/lang/main.rkt
#lang racket/base

;; Re-export ALL of Racket's functions EXCEPT the ones we provide ourselves
(provide (rename-out [module-begin #%module-begin])      ; Our custom module-begin
         (except-out (all-from-out racket/base)          ; Everything from Racket...
                     #%module-begin                       ; ...except their module-begin
                     read-syntax))                        ; ...and their read-syntax

;; This is required for any #lang - it sets up the module
(define-syntax-rule (module-begin body ...)
  (#%module-begin
   body ...))
