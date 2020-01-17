package ctl

import (
	"fmt"
	"sort"
	"strings"
)

// State is a high level description of a variable assignment
type State struct {
	bools map[string]bool
	ints  map[string]uint
}

// Equals check if state s has the same assingment as state t.
func (s *State) Equals(t *State) bool {
	if len(s.bools) != len(t.bools) || len(s.ints) != len(t.ints) {
		return false
	}
	for k, v := range s.bools {
		if w, in := t.bools[k]; !in || v != w {
			return false
		}
	}
	for k, v := range s.ints {
		if w, in := t.ints[k]; !in || v != w {
			return false
		}
	}
	return true
}

// Less computes if the s has lower values than t.
func (s *State) Less(t *State) bool {
	// First consider booleans, then integers.
	if len(s.bools) != len(t.bools) {
		return len(s.bools) < len(t.bools)
	}
	names := make([]string, 0, len(s.bools))
	for name := range s.bools {
		names = append(names, name)
	}
	sort.Sort(ByStringLt(names))
	for _, name := range names {
		if yv, in := t.bools[name]; !in || yv != s.bools[name] {
			return in && !s.bools[name] && yv
		}
	}
	// Consider integers.
	if len(s.ints) != len(t.ints) {
		return len(s.ints) < len(t.ints)
	}
	names = make([]string, 0, len(s.ints))
	for name := range s.ints {
		names = append(names, name)
	}
	sort.Sort(ByStringLt(names))
	for _, name := range names {
		if yv, in := t.ints[name]; !in || yv != s.ints[name] {
			return in && s.ints[name] < yv
		}
	}
	// All equal.
	return false
}

// States is a list of states.
type States []*State

func (a States) Len() int           { return len(a) }
func (a States) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a States) Less(i, j int) bool { return a[i].Less(a[j]) }

// Equals checks if two lists of states are exactly the same.
func (a States) Equals(b States) bool {
	if len(a) != len(b) {
		return false
	}
	for i, s := range a {
		if !s.Equals(b[i]) {
			return false
		}
	}
	return true
}

// Get all accepted assignments (free variables are left unspecified).
func unpackBDD(p *BDD) []map[*Variable]bool {
	if p.Node() {
		t := unpackBDD(p.True)
		f := unpackBDD(p.False)
		result := make([]map[*Variable]bool, len(t)+len(f))
		for i := 0; i < len(t); i++ {
			t[i][p.Var] = true
			result[i] = t[i]
		}
		for i := 0; i < len(f); i++ {
			f[i][p.Var] = false
			result[len(t)+i] = f[i]
		}
		return result
	} else if p.Value {
		return []map[*Variable]bool{map[*Variable]bool{}}
	} else {
		return []map[*Variable]bool{}
	}
}

// Expand free variables in the given states.
func expandStates(vars []*Variable, aux bool, states []map[*Variable]bool) []map[*Variable]bool {
	if len(vars) == 0 {
		return states
	}

	// Get the head variable and check if we want it to be assigned.
	v := vars[0]
	if !aux && v.aux {
		return expandStates(vars[1:], aux, states)
	}

	// Make sure all states assign v.
	result := make([]map[*Variable]bool, 0, len(states))
	for _, state := range states {
		if _, in := state[v]; !in {
			copy := make(map[*Variable]bool, len(state)+1)
			for k, v := range state {
				copy[k] = v
			}
			state[v] = true
			copy[v] = false
			result = append(result, state, copy)
		} else {
			result = append(result, state)
		}
	}
	return expandStates(vars[1:], aux, result)
}

// Process one state (expand names and compute integer values).
// Auxiliary values are discarded unless aux is set.
func processState(m *Model, state map[*Variable]bool, aux bool) *State {
	bools := make(map[string]bool)
	ints := make(map[string]uint, len(m.ints))

	// Extract integer values.
	for _, i := range m.ints {
		if !aux && i.Aux() {
			continue
		}

		// Compute value
		value := uint(0)
		for n, v := range i.bits {
			if state[v] {
				value += 1 << n
			}
			delete(state, v)
		}

		// Store value
		ints[i.Name()] = value
	}

	// Extract boolean values.
	for v, b := range state {
		if aux || !v.aux {
			bools[v.Name] = b
		}
	}

	return &State{bools, ints}
}

// Process a list of states.
//
// Note: For reasons I do not yet understand auxiliary variables are sometimes
// restricted in unexpected ways such as in the following example. To get a
// clean list of states I remove any duplicate results here.
//
//  a | add(a,5) | add(a,a)
// ---|----------|----------
// 44 | 0        | 88
// 44 | 1        | 88
// 44 | 3        | 88
// 44 | 5        | 88
// 44 | 9        | 88
// 44 | 17       | 88
// 44 | 49       | 0
// 44 | 113      | 88
// 44 | 177      | 88
func processStates(m *Model, states []map[*Variable]bool, aux bool) States {
	result := make(States, 0, len(states))
	for _, state := range states {
		newState := processState(m, state, aux)
		duplicate := false
		for _, otherState := range result {
			if newState.Equals(otherState) {
				duplicate = true
				break
			}
		}

		if !duplicate {
			result = append(result, newState)
		}
	}

	sort.Sort(result)
	return result
}

// Convert states to tabular data.
func convertStatesToTable(states States) [][]string {
	table := make([][]string, 1, len(states)+1)

	// Assume each state has the same variables. Note that we sort the boolean and
	// integer names separately to align with States.Less.
	st0 := states[0]
	names := make([]string, 0, len(st0.bools)+len(st0.ints))
	intNames := make([]string, 0, len(st0.ints))
	for name := range st0.bools {
		names = append(names, name)
	}
	for name := range st0.ints {
		intNames = append(intNames, name)
	}
	sort.Sort(ByStringLt(names))
	sort.Sort(ByStringLt(intNames))
	names = append(names, intNames...)

	// Extract values from each state.
	table[0] = names
	for _, state := range states {
		values := make([]string, len(names))
		for i, name := range names {
			if i < len(st0.bools) {
				values[i] = fmt.Sprintf("%v", state.bools[name])
			} else {
				values[i] = fmt.Sprintf("%v", state.ints[name])
			}
		}
		table = append(table, values)
	}

	return table
}

// Print all the given states as a TSV.
func printStates(states States) {
	table := convertStatesToTable(states)
	for _, row := range table {
		println(strings.Join(row, "\t"))
	}
}
