package ctl

import (
	"fmt"
	"testing"
)

// SimpleModel tests a very basic model.
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
	if leastSteps(sets, init) != 2 {
		t.Error("expected two steps")
	}
}

func leastSteps(steps []*BDD, init *BDD) int {
	for i := 1; i < len(steps); i++ {
		if steps[i].Contains(init) {
			return i
		}
	}
	return -1
}

func findTruePaths(m *Model, p *BDD) []map[string]bool {
	if p.Node() {
		name := m.variables[p.ID>>1]
		t := findTruePaths(m, p.True)
		f := findTruePaths(m, p.False)
		result := make([]map[string]bool, len(t)+len(f))
		for i := 0; i < len(t); i++ {
			t[i][name] = true
			result[i] = t[i]
		}
		for i := 0; i < len(f); i++ {
			f[i][name] = false
			result[len(t)+i] = f[i]
		}
		return result
	} else if p.Value {
		return []map[string]bool{map[string]bool{}}
	} else {
		return []map[string]bool{}
	}
}

func printTruePaths(paths []map[string]bool) {
	for _, path := range paths {
		line := ""
		for name := range path {
			line += fmt.Sprintf("%v: %v, ", name, path[name])
		}
		println(line)
	}
}
