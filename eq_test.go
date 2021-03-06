// This file provides unit tests for the Eq data type.

package intern_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"testing"

	"github.com/spakin/intern"
)

// TestNewEq tests if we can create a large number of symbols for which
// duplicates are certain to occur.
func TestNewEq(*testing.T) {
	const sLen = 3                       // Symbol length in characters
	const nSymbols = 1000000             // Must be greater than len(charSet) choose sLen
	prng := rand.New(rand.NewSource(12)) // Constant for reproducibility
	for i := 0; i < nSymbols; i++ {
		_ = intern.NewEq(randomString(prng, sLen))
	}
}

// TestEqString tests if we can convert strings to symbols and back to
// strings.
func TestEqString(t *testing.T) {
	// Prepare the test.
	const ns = 10000                     // Number of strings to generate
	strs := make([]string, ns)           // Original strings
	syms := make([]intern.Eq, ns)        // Interned strings
	prng := rand.New(rand.NewSource(34)) // Constant for reproducibility

	// Generate a bunch of strings.
	for i := range strs {
		nc := prng.Intn(20) + 1 // Number of characters
		strs[i] = randomString(prng, nc)
	}

	// Intern each string to an Eq.
	for i, s := range strs {
		syms[i] = intern.NewEq(s)
	}

	// Ensure that converting an Eq back to a string is a lossless
	// operation.  We use fmt.Sprintf as this represents a typical way an
	// Eq might be converted to a string.
	for i, str := range strs {
		sym := syms[i]
		sStr := fmt.Sprintf("%s", sym)
		if str != sStr {
			t.Fatalf("Expected %q but saw %q", str, sStr)
		}
	}
}

// TestEqStringMulti tests if we can convert strings to symbols and back to
// strings.  Unlike TestEqString, TestEqStringMulti uses NewEqMulti.
func TestEqStringMulti(t *testing.T) {
	// Prepare the test.
	const ns = 10000                     // Number of strings to generate
	strs := make([]string, ns)           // Original strings
	prng := rand.New(rand.NewSource(34)) // Constant for reproducibility

	// Generate a bunch of strings.
	for i := range strs {
		nc := prng.Intn(20) + 1 // Number of characters
		strs[i] = randomString(prng, nc)
	}

	// Intern each string to an Eq.
	syms := intern.NewEqMulti(strs)

	// Ensure that converting an Eq back to a string is a lossless
	// operation.  We use fmt.Sprintf as this represents a typical way an
	// Eq might be converted to a string.
	for i, str := range strs {
		sym := syms[i]
		sStr := fmt.Sprintf("%s", sym)
		if str != sStr {
			t.Fatalf("Expected %q but saw %q", str, sStr)
		}
	}
}

// TestBadEq ensures we panic when converting an invalid Eq to a
// string.
func TestBadEq(t *testing.T) {
	defer func() { _ = recover() }()
	var bad intern.Eq
	_ = bad.String() // Should panic
	t.Fatalf("Failed to catch invalid intern.Eq %d", bad)
}

// TestForgetAllEqs ensures we panic when converting a forgotten Eq to a
// string.
func TestForgetAllEqs(t *testing.T) {
	defer func() { _ = recover() }()
	sym := intern.NewEq("old string")
	str := sym.String()
	intern.ForgetAllEqs()
	str = sym.String() // Should panic
	t.Fatalf("Failed to catch invalid intern.Eq %d (%q)", sym, str)
}

// TestEqCase ensures that symbol comparisons are case-sensitive.
func TestEqCase(t *testing.T) {
	// Convert a set of strings to Eqs.
	strs := []string{
		"roadrunner",
		"Roadrunner",
		"roadRunner",
		"ROADRUNNER",
		"rOaDrUnNeR",
		"ROADrunner",
		"roadRUNNER",
	}
	syms := make([]intern.Eq, len(strs))
	for i, s := range strs {
		syms[i] = intern.NewEq(s)
	}

	// Ensure that each symbol is equal only to itself.
	numEq := 0
	for _, s1 := range syms {
		for _, s2 := range syms {
			if s1 == s2 {
				numEq++
			}
		}
	}
	if numEq != len(syms) {
		t.Fatalf("Expected %d case-sensitive comparisons but saw %d",
			len(syms), numEq)
	}
}

// TestEqConcurrent performs a bunch of accesses in parallel in an attempt to
// expose race conditions.
func TestEqConcurrent(*testing.T) {
	const symsPerThread = 100000
	nThreads := runtime.NumCPU() * 2 // Oversubscribe CPUs by a factor of 2.

	// Spawn a number of goroutines.
	begin := make(chan bool, nThreads)
	done := make(chan bool, nThreads)
	for j := 0; j < nThreads; j++ {
		go func() {
			prng := rand.New(rand.NewSource(2021)) // Constant for reproducibility and to invite conflicts
			_ = <-begin
			for i := 0; i < symsPerThread; i++ {
				nc := prng.Intn(20) + 1 // Number of characters
				_ = intern.NewEq(randomString(prng, nc))
			}
			done <- true
		}()
	}

	// Tell all goroutines to begin then wait for them all to finish.
	for j := 0; j < nThreads; j++ {
		begin <- true
	}
	for j := 0; j < nThreads; j++ {
		_ = <-done
	}
}

// TestEqMarshalJSON marshals Eqs to JSON and back and checks that the outputs
// match the input.
func TestEqMarshalJSON(t *testing.T) {
	for r, rStr := range []string{
		"NoForget",
		"Forget",
	} {
		t.Run(rStr, func(t *testing.T) {
			// Create a long slice of Eqs.
			intern.ForgetAllEqs()
			iSyms := intern.NewEqMulti(ozChars)

			// Encode the Eqs as JSON.
			b, err := json.MarshalIndent(iSyms, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			// On our second iteration, forget our entire mapping.
			if r == 1 {
				intern.ForgetAllEqs()
			}

			// Convert the JSON back to a slice of Eqs.
			var oSyms []intern.Eq
			err = json.Unmarshal(b, &oSyms)
			if err != nil {
				t.Fatal(err)
			}

			// Ensure that the outputs match the original strings.
			for i, s := range ozChars {
				if s != oSyms[i].String() {
					t.Fatalf("Expected %q but saw input symbol %q", s, oSyms[i])
				}
			}
		})
	}
}

// TestEqMarshalGob marshals Eqs to a gob and back and checks that the outputs
// match the input.
func TestEqMarshalGob(t *testing.T) {
	for r, rStr := range []string{
		"NoForget",
		"Forget",
	} {
		t.Run(rStr, func(t *testing.T) {
			// Create a long slice of Eqs.
			intern.ForgetAllEqs()
			iSyms := intern.NewEqMulti(ozChars)

			// Encode the Eqs as a gob.
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(&iSyms)
			if err != nil {
				t.Fatal(err)
			}

			// On our second iteration, forget our entire mapping.
			if r == 1 {
				intern.ForgetAllEqs()
			}

			// Convert the gob back to a slice of Eqs.
			var oSyms []intern.Eq
			dec := gob.NewDecoder(&buf)
			err = dec.Decode(&oSyms)
			if err != nil {
				t.Fatal(err)
			}

			// Ensure that the outputs match the original strings.
			for i, s := range ozChars {
				if s != oSyms[i].String() {
					t.Fatalf("Expected %q but saw input symbol %q", s, oSyms[i])
				}
			}
		})
	}
}
