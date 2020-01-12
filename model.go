package ctl

import (
	"fmt"
)

// Model describes a set of variables and transitions.
type Model struct {
	vars  map[uint]string
	ints  []*Integer
	trans *BDD
}

// For variable a with index i we define ID(a) := i<<1, ID(next(a)) := i<<1 | 1.
func varID(i uint) uint {
	return i << 1
}

// Since we usually store ID(a), the argument is i<<1.
func varNextID(id uint) uint {
	return id | 1
}

// NewModel creates a new model.
func NewModel() *Model {
	return &Model{
		make(map[uint]string),
		make([]*Integer, 0),
		nil}
}

// Var creates a new variable identifier.
func (m *Model) Var(name string) uint {
	i := uint(len(m.vars) + 1)
	m.vars[i] = name
	return i
}

// Bool creates a new boolean variable.
func (m *Model) Bool(name string) *BDD {
	return Node(varID(m.Var(name)), True, False)
}

// Int creates a new integer variable that contains the given upperbound.
func (m *Model) Int(name string, upb uint) *Integer {
	return m.Bin(name, bitcount(upb))
}

// Bin creates a new integer variable with the given number of bits.
func (m *Model) Bin(name string, n uint) *Integer {
	bits := make([]uint, n)
	for i := range bits {
		bits[i] = varID(m.Var(fmt.Sprintf("%v@%v", name, i)))
	}
	integer := &Integer{bits, true, True}
	m.ints = append(m.ints, integer)
	return integer
}

// Name gets a variable name.
func (m *Model) Name(id uint) string {
	return m.vars[id>>1]
}

// IntName gets an integer variable name.
func (m *Model) IntName(i *Integer) string {
	if i.variable {
		bitID := i.bits[0]
		lbl := m.Name(bitID)
		return lbl[:len(lbl)-2]
	}
	// Compute value and return as string.
	return fmt.Sprintf("%v", i.ConstValue())
}

// Add adds a new transition.
func (m *Model) Add(condition *BDD, constraint *BDD) {
	if m.trans == nil {
		m.trans = condition.And(constraint)
	} else {
		m.trans = m.trans.Or(condition.And(constraint))
	}
}

// EX collects the set of states in start that transition to next in one step.
func (m *Model) EX(start *BDD, goal *BDD) *BDD {
	// A state is included if there exists next(a1)...next(an) such that:
	states := start.And(m.trans).And(goal.Next())
	// Eliminate exists.
	for i := range m.vars {
		states = states.Exists(varNextID(varID(i)))
	}
	return states
}

// EG collects all state sets for which there exists a path of n steps such that
// for each step a condition holds. The states for which there exists a path of
// n steps that satisfy this condition is returned in the n-th index. If the
// final set is empty there is no path for which the condition globally holds.
func (m *Model) EG(condition *BDD) []*BDD {
	result := make([]*BDD, 0)
	last := condition
	var next *BDD
	for {
		result = append(result, last)
		next = last.And(m.EX(condition, last))
		if next.Equals(last) {
			return result
		}
		last = next
	}
}

// EU collects all state sets that can transition to goal such that a given
// condition holds for all steps. The states for which this is possible in n
// steps is returned in the n-th index.
func (m *Model) EU(step *BDD, goal *BDD) []*BDD {
	result := make([]*BDD, 0)
	last := goal
	var next *BDD
	for {
		result = append(result, last)
		next = last.Or(m.EX(step, last))
		if next.Equals(last) {
			return result
		}
		last = next
	}
}

// EF collects all state sets that can transition to goal in n steps.
func (m *Model) EF(goal *BDD) []*BDD {
	return m.EU(True, goal)
}
