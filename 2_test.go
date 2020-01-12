package ctl

import (
	"testing"
)

// TestMarbleGame tests a simple model with numbers.
func TestMarbleGame(t *testing.T) {
	m := NewModel()
	a := m.Int("a", 100)

	// Five marbles are added.
	m.Add(a.Leq(Int(95)), a.Next().Eq(a.Add(Int(5), m)))
	// The number of marbles is doubled.
	m.Add(a.Leq(Int(50)), a.Next().Eq(a.Add(a, m)))

	// Check that 98 marbles are reachable in 6 steps.
	if LeastSteps(a.Eq(Int(1)), m.EF(a.Eq(Int(98)))) != 6 {
		t.Error("expected six steps")
	}
}

// TestCrash determines for which values of n in 1..10 the following program
// crashes and generates examples in the case of a crash.
//
//     a := 1
//     b := 1
//     for i := 1..10 {
//       if <unknown>
//       then { a = a + 2b; b = b + i }
//       else { b = a + b; a = a + i }
//     }
//     if b == <600 + n>
//     then <crash>
//
func TestCrash(t *testing.T) {
	m := NewModel()
	a := m.Int("a", 610)
	b := m.Int("b", 610)
	i := m.Int("i", 11)
	x := m.Bool("x")

	t.Skip("Currently 11 bit addition is already too slow")

	a1 := a.Add(b, m).Add(b, m)
	b1 := b.Add(i, m)
	b2 := a.Add(b, m)
	a2 := a.Add(i, m)
	upb := Int(610)
	inc := i.Next().Eq(i.Add(Int(1), m))

	// <unknown> = true
	m.Add(i.Leq(Int(10)).And(a1.Leq(upb)).And(b1.Leq(upb)),
		x.Next().Eq(True).And(a.Next().Eq(a1)).And(b.Next().Eq(b1)).And(inc))
	// <unknown> = false
	m.Add(i.Leq(Int(10)).And(a2.Leq(upb)).And(b2.Leq(upb)),
		x.Next().Eq(False).And(a.Next().Eq(a2)).And(b.Next().Eq(b2)).And(inc))

	// <crash>
	shouldCrash := []uint{2, 3, 8, 10} // Determined earlier with NuSMV.
	for n := uint(1); n <= 10; n++ {
		start := a.Eq(Int(1)).And(b.Eq(Int(1))).And(i.Eq(Int(1)))
		crash := i.Eq(Int(11)).And(b.Eq(Int(600 + n)))
		sets := m.EF(crash)
		path := GenerateExample(m, start, sets)
		if len(path) == 0 {
			// Check if indeed no crash was expected.
			for _, m := range shouldCrash {
				if n == m {
					t.Errorf("a crash is expected")
				}
			}
		} else {
			// Check if counter example is correct.
			a, b := uint(1), uint(1)
			for i, state := range path {
				// Compute change for i > 0.
				if i > 0 {
					if state.bools["x"] {
						a = a + 2*b
						b = b + uint(i)
					} else {
						b = a + b
						a = a + uint(i)
					}
				}
				// Compare simulated a and b with example value.
				if state.ints["a"] != a || state.ints["b"] != b {
					t.Errorf("a and b should start at 1")
				}
			}
			if len(path) != 11 || path[10].ints["b"] != 600+n {
				t.Errorf("generated example does not crash")
			}
		}
	}
}
