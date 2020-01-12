package ctl

import (
	"testing"
)

// TestSimpleModel tests a very basic boolean model.
func TestSimpleModel(t *testing.T) {
	m := NewModel()
	a := m.Bool("a")
	b := m.Bool("b")

	// a = false & b = false -> a := !a & b := b
	m.Add(a.Eq(False).And(b.Eq(False)), a.Next().Eq(a.Neg()).And(b.Next().Eq(b)))
	// a = !b -> b := a & a := a
	m.Add(a.Eq(b.Neg()), b.Next().Eq(a).And(a.Next().Eq(a)))

	// Check that a:0;b:0 -> a:1;b:1 takes 2 steps.
	init := a.Eq(False).And(b.Eq(False))
	goal := a.Eq(True).And(b.Eq(True))
	sets := m.EF(goal)

	// Note a:0;b:1 -> a:0;b:0 so a:1;b:1 is reached from all states in <=4 steps.
	if len(sets) != 4 {
		t.Error("expected four sets")
	}
	if leastSteps(init, sets) != 2 {
		t.Error("expected two steps")
	}
}
