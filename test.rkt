#lang reader "lang.rkt"

displayln("Clay test suite")
displayln("=================")

define(assert(condition, message),
  unless(condition, error(message)))

define(test(name, body),
  displayln(string-append("Testing: ", name)),
  body,
  displayln("✓ PASSED"))

define(plus(a, b), +(a, b))

test("defining a function and using math",
    assert(equal?(plus(1, 2), 3), "plus(1, 2) should be 3"))


define minus(a, b) -(a, b)

displayln("testing...")

test("defining a function using whitepsace to separate arguments",
    assert(equal?(minus(2, 1), 1), "minus(2, 1) should be 1"))

define multiply(a, b) do
    * a b
end

test("defining a function using do end syntax",
    assert(equal?(multiply(2, 3), 6), "multiply(2, 3) should be 6"))
