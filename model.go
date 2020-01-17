package ctl

import (
	"fmt"
)

// Variable identifies a boolean variable.
type Variable struct {
	Name string     // Variable name (should be unique)
	seq  uint       // Variable sequence number (may change)
	aux  bool       // Flag for auxiliary variables (not in the visible state)
	next bool       // Flag for next twin variable
	twin *Variable  // Twin variable (next or normal)
	tail *Variable  // Variable that comes after this one in the ordering.
	root **Variable // Model reference for on the fly assignment
}

// Check if this variable is already assigned to a position in the ordering.
// Note that when constructing a BDD with just one variable, it does not matter
// that this variable is not in the ordering.
func (v *Variable) ordered() bool {
	// An ordered variable has a sequence number >0.
	return v.seq != 0
}

// Lt checks if v is before w in the variable ordering. If either variable is
// not yet assigned a location in the ordering, they will be put close together.
func (v *Variable) Lt(w *Variable) bool {
	// Check if the variables are equal or either of them is nil.
	if v == w || w == nil {
		return false
	} else if v == nil {
		return true
	}

	// Check if either variable is not in the ordering.
	if !v.ordered() || !w.ordered() {
		// If both are not ordered, add them to the front of the ordering.
		if !v.ordered() && !w.ordered() {
			// root (-> v -> w) -> root.tail
			v.Norm().tail = w.Norm()
			w.Norm().tail = *v.root
			*v.root = v.Norm()
		} else if v.ordered() {
			// v (-> w) -> v.tail
			w.Norm().tail = v.Norm().tail
			v.Norm().tail = w.Norm()
		} else if w.ordered() {
			// w (-> v) -> w.tail
			v.Norm().tail = w.Norm().tail
			w.Norm().tail = v.Norm()
		}

		// Update sequence numbers of all variables.
		for r, seq := *v.root, uint(1); r != nil; r, seq = r.tail, seq+2 {
			r.seq = seq
			r.twin.seq = seq + 1
		}
	}

	return v.seq < w.seq
}

// Next returns a reference to the next twin variable.
func (v *Variable) Next() *Variable {
	if !v.next {
		return v.twin
	}
	return v
}

// Norm returns a reference to the normal twin variable.
func (v *Variable) Norm() *Variable {
	if v.next {
		return v.twin
	}
	return v
}

// Model describes a set of variables and transitions.
type Model struct {
	vars  []*Variable // All variables in the model
	ints  []*Integer  // All integers in the model
	order *Variable   // Variable ordering
	trans *BDD
}

// NewModel creates a new model.
func NewModel() *Model {
	return &Model{
		make([]*Variable, 0),
		make([]*Integer, 0),
		nil,
		nil}
}

// Var creates a new variable reference.
func (m *Model) Var(name string, aux bool) *Variable {
	// Fallback to naive ordering.
	seq := uint(2*len(m.vars) + 1)
	i1, i2 := seq, seq+1

	nextName := fmt.Sprintf("next(%v)", name)
	v := &Variable{name, i1, aux, false, nil, nil, &m.order}
	v.twin = &Variable{nextName, i2, aux, true, v, nil, &m.order}
	m.vars = append(m.vars, v)
	return v
}

// Bool creates a new boolean variable.
func (m *Model) Bool(name string) *BDD {
	v := m.Var(name, false)
	return Node(v, True, False)
}

// Int creates a new integer variable that contains the given upperbound.
func (m *Model) Int(name string, upb uint) *Integer {
	return m.bin(name, bitcount(upb), false)
}

// bin creates a new integer variable with the given number of bits. Auxiliary
// binary numbers are created for arithmetic operations, but these do not belong
// to the state (when generating counter examples etc.).
func (m *Model) bin(name string, n uint, aux bool) *Integer {
	bits := make([]*Variable, n)
	for i := range bits {
		bits[i] = m.Var(fmt.Sprintf("%v@%v", name, i), aux)
	}
	integer := &Integer{bits, 0, true, True}
	m.ints = append(m.ints, integer)
	return integer
}

// Add adds a new transition.
func (m *Model) Add(condition *BDD, constraint *BDD) {
	if m.trans == nil {
		m.trans = condition.And(constraint)
	} else {
		m.trans = m.trans.Or(condition.And(constraint))
	}
}

// EX returns the states in start that transition to next in one step.
func (m *Model) EX(start *BDD, goal *BDD) *BDD {
	// A state is included if there exists next(a1)...next(an) such that:
	states := start.And(m.trans).And(goal.Next())
	// Eliminate exists.
	for _, v := range m.vars {
		states = states.Exists(v.Next())
	}
	return states
}

// EXInv returns the states in goal that transition from start in one step.
func (m *Model) EXInv(start *BDD, goal *BDD) *BDD {
	// A state is included if there exists a1...an such that:
	states := goal.Next().And(m.trans).And(start)
	// Eliminate exists.
	for _, v := range m.vars {
		states = states.Exists(v)
	}
	// The states BDD contains all next variables, convert this back to normal.
	return states.Norm()
}

// EG returns states for which there exists a path of n steps such that for each
// step a condition holds. The states for which there exists a path of n steps
// that satisfy this condition is returned in the n-th index. If the final set
// is empty there is no path for which the condition globally holds.
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

// EU returns all states that can transition to goal such that a given condition
// holds for all steps. The states for which this is possible in n steps is
// returned in the n-th index.
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

// PrintStates is a utility to print all states in the given BDD as TSV data.
func (m *Model) PrintStates(p *BDD, aux bool) {
	states := expandStates(m.vars, aux, unpackBDD(p))
	printStates(processStates(m, states, aux))
}
