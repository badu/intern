// This file provides the Eq type, which represents strings that can
// be compared only for equality.

package intern

import (
	"fmt"
)

// An Eq is a string that has been interned to an integer.  Eq supports only
// equality and inequality comparisons, not greater than/less than comparisons.
// (No checks are performed to enforce that usage model, unfortunately.)
type Eq uint64

// eq maintains all the state needed to manipulate Eqs.
var eq state

// init initializes our global state.
func init() {
	eq.forgetAll()
}

// NewEq maps a string to an Eq symbol.  It guarantees that two equal strings
// will always map to the same Eq.
func NewEq(s string) Eq {
	var err error
	st := &eq
	st.Lock()
	defer st.Unlock()
	sym, err := st.assignSymbol(s, false)
	if err != nil {
		panic(fmt.Sprintf("Internal error: Unexpected error (%s)", err))
	}
	return Eq(sym)
}

// String converts an Eq back to a string.  It panics if given an Eq that was
// not created using NewEq.
func (s Eq) String() string {
	return eq.toString(uint64(s), "Eq")
}

// ForgetAllEqs discards all existing mappings from strings to Eqs so the
// associated memory can be reclaimed.  Use this function only when you know
// for sure that no previously mapped Eqs will subsequently be used.
func ForgetAllEqs() {
	eq.Lock()
	eq.forgetAll()
	eq.Unlock()
}
