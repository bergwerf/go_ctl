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
	if leastSteps(a.Eq(Int(1)), m.EF(a.Eq(Int(98)))) != 6 {
		t.Error("expected six steps")
	}
}
