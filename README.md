Heisenlisp
==========

Heisenlisp is a functional Lisp-style interpreted language implemented in Go,
with special features for evaluating expressions involving uncertainty.

It can be interacted with in a REPL:

    ..? (= 2 2)
    ==> true

It is also meant to be able to be embedded in Go programs, although doing
so will take a bunch of familiarity with the code.

Project status
==============

Heisenlisp is meant to be an interesting experiment in language design.
While I have some potential applications for playing around with it in mind,
it is at this stage unlikely be of practical use to others.

Note that this implementation is very slow, even for an interpreted language.

More about the language
=======================

Heisenlisp is a Lisp-1, meaning functions and other values live in the same
namespace. In other words, if you make a variable `list`, it will shadow
the function `list`.

All objects in Heisenlisp are immutable.

All named impure forms (functions with side-effects) are required to have
names ending end with exclamation mark. Only impure functions can call other
impure functions. (The only current exception is the debugging feature
`trace`.)

The built-in impure forms are `set!`, `defun!`, and `defmacro!`.

There is currently no specification of the language besides the
implementation itself.

Heisenlisp is Turing-complete.

No importance has been placed on Lisp "purity". At the time of this writing
most of the language features are implemented in Go.

Uncertainty features
====================

Heisenlisp supports certain kinds of computations on uncertain values.
Essentially, it operates on "ternary" boolean logic: true, false,
or "maybe".

Here are some demonstrations:

    ..? (= 2 (any-of 2 3))
    ==> maybe
    ..? (* 2 (any-of 2 3))
    ==> #any-of(4 6)
    ..? (odd? (* 2 (any-of 2 3)))
    ==> false
    ..? (length (any-of nil (list 1) (list 1 2)))
    ==> #any-of(0 1 2)
    ..? (* 10 (+ 5 (number-in-range 'from 1 'to 10)))
    ==> #number-in-range([60,150])
    ..? (= 12 (fold-left * 1 (filter (lambda (x) maybe) (range 5))))
    ==> maybe

The special forms `if`, `and`, and `or` have been designed to accommodate
uncertainty, as have reduction functions such as `fold-left` and `fold-right`.

The language does _not_ operate on probability distributions. You can
use the Heisenlisp uncertainty features to determine the possible
outcomes of a dice roll, but not their probabilities:

    ..? (+ (any-of 1 2 3 4 5 6) (any-of 1 2 3 4 5 6))
    ==> #any-of(2 3 4 5 6 7 8 9 10 11 12)

The language also reserves the right for an implementation to return
an answer that is correct but not as precise as possible. This can
be used to avoid excessively large combinatorial explosions:

    ..? (* (apply any-of (range 5)) (+ 1 (apply any-of (range 5))))
    ==> #any-of(0 1 2 3 4 5 6 8 10 9 12 15 16 20)
    ..? (* (apply any-of (range 50)) (+ 1 (apply any-of (range 50))))
    ==> #unknown

(It is true that in principle this means that a conforming implementation
could have every single expression evaluate to `#unknown`. In practice,
that's not what this implementation does; it attempts to give useful
answers.)

When uncertainty is not desired, `may?` and `must?` can be used to convert
Heisenlisp-style "ternary" boolean logic to the normal kind:

    ..? (all? odd? (list 1 3 (any-of 4 5)))
    ==> maybe
    ..? (all? odd? (list 1 3 (any-of 5 7)))
    ==> true
    ..? (may? (all? odd? (list 1 3 (any-of 4 5))))
    ==> true
    ..? (must? (all? odd? (list 1 3 (any-of 4 5))))
    ==> false

Functions prefixed with an underscore operate directly on a value without
considering it as an uncertain value that represents other values:

    ..? (any-of 1 2 3)
    ==> #any-of(1 2 3)
    ..? (type (any-of 1 2 3))
    ==> integer
    ..? (_type (any-of 1 2 3))
    ==> any-of

Legal stuff
===========

I (@steinarvk) hold the copyright on this code. It is not associated with
any employer of mine, past or present.

The code is made available for use under the MIT license; see the LICENSE
file for details.
