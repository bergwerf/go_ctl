package ctl

import "testing"

// SimpleModel tests a very basic model.
func TestSimpleModel(t *testing.T) {
	m := NewModel()
	a := m.Bool("a")
	b := m.Bool("b")

	// a = false & b = false -> a := !a & b := b
	m.Add(a.Eq(False).And(b.Eq(False)), a.Next().Eq(a.Neg()).And(b.Next().Eq(b)))
	// a = !b -> b := a & a := a
	m.Add(a.Eq(b.Neg()), b.Next().Eq(a).And(a.Next().Eq(a)))

	// Check if a = true & b = true is reachable.
	init := a.Eq(False).And(b.Eq(False))
	goal := a.Eq(True).And(b.Eq(True))
	sets := m.EF(goal)

	if !sets[len(sets)-1].Contains(init) {
		t.Error("expected reachability")
	}
}
