package ctl

// Trans describes one transition.
type Trans struct {
	condition  *BDD
	constraint *BDD
}

// Model describes a set of variables and transitions.
type Model struct {
	variables   map[uint]string
	transitions []Trans
	step        *BDD
}

// NewModel creates a new model.
func NewModel() Model {
	return Model{make(map[uint]string), make([]Trans, 0), True}
}

// Bool creates a new boolean variable.
func (m *Model) Bool(name string) *BDD {
	id := uint(len(m.variables) + 1)
	m.variables[id] = name
	return Node(id<<1, True, False)
}

// Add adds a new transition.
func (m *Model) Add(condition *BDD, constraint *BDD) {
	m.transitions = append(m.transitions, Trans{condition, constraint})
	m.step = m.step.Or(condition.And(constraint))
}

// EX collects the set of states in start that transition to next in one step.
func (m *Model) EX(start *BDD, goal *BDD) *BDD {
	// A state is included if there exists next(a1)...next(an) such that:
	states := start.And(m.step).And(goal.Next())
	// Eliminate exists.
	for key := range m.variables {
		states = states.Exists((key << 1) | 1)
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
	var prev *BDD
	for prev == nil || !prev.Equals(last) {
		result = append(result, last)
		last := prev
		prev = last.And(m.EX(condition, last))
	}
	return result
}

// EU collects all state sets that can transition to goal such that a given
// condition holds for all steps. The states for which this is possible in n
// steps is returned in the n-th index.
func (m *Model) EU(step *BDD, goal *BDD) []*BDD {
	result := make([]*BDD, 0)
	last := goal
	var prev *BDD
	for prev == nil || !prev.Equals(last) {
		result = append(result, last)
		last := prev
		prev = last.Or(m.EX(step, last))
	}
	return result
}

// EF collects all state sets that can transition to goal in n steps.
func (m *Model) EF(goal *BDD) []*BDD {
	return m.EU(True, goal)
}
