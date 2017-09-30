// Package cyclebreaker is a horrible hack to avoid import cycles.
// Import cyclebreaker/impl (e.g. in the main function) to get the implementation of these.
package cyclebreaker

/* For reference, here's the cycle that inspired this hack.

Equals(): fundamentally a symmetric function so we
          don't want to make it an interface function
          that needs to be re-implemented everywhere.
          Depends on "numerics" to do numeric equality
          comparison.
numerics: Depends on "anyof" for MaybeValue, because
          the result of some comparisons is Maybe.
anyof:    Depends on Equals() because lists of values
          need to be de-duplicated when a new AnyOf
          value is created.

*/

import "github.com/steinarvk/heisenlisp/types"

var Equals func(a, b types.Value) (types.TernaryTruthValue, error)

var AtomEquals func(a, b types.Value) bool
