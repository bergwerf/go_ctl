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

// Get all accepted assignments.
func unpackBDD(p *BDD) []map[uint]bool {
	if p.Node() {
		t := unpackBDD(p.True)
		f := unpackBDD(p.False)
		result := make([]map[uint]bool, len(t)+len(f))
		for i := 0; i < len(t); i++ {
			t[i][p.ID] = true
			result[i] = t[i]
		}
		for i := 0; i < len(f); i++ {
			f[i][p.ID] = false
			result[len(t)+i] = f[i]
		}
		return result
	} else if p.Value {
		return []map[uint]bool{map[uint]bool{}}
	} else {
		return []map[uint]bool{}
	}
}

// Expand free variables in the given states.
func expandStates(ids []uint, states []map[uint]bool) []map[uint]bool {
	if len(ids) == 0 {
		return states
	}
	// Make sure all states assign the first id.
	id := ids[0]
	result := make([]map[uint]bool, 0, len(states))
	for _, state := range states {
		if _, in := state[id]; !in {
			copy := make(map[uint]bool, len(state)+1)
			for k, v := range state {
				copy[k] = v
			}
			state[id] = true
			copy[id] = false
			result = append(result, state, copy)
		} else {
			result = append(result, state)
		}
	}
	// Recurse on the tail.
	return expandStates(ids[1:], result)
}

// Get all accepted states within the state space of the given booleans and
// integers. If all is set, then all variables that are fixed by p are included.
//
// If all == true, then some states may appear duplicated because auxiliary
// variables are free depending on which transition underlies the given state.
func unpackStates(m *Model,
	booleans []uint, integers []*Integer, all bool, p *BDD) States {
	ids := make([]uint, 0, len(booleans)) // Does not include cap for integers.
	for _, id := range booleans {
		ids = append(ids, id)
	}
	for _, i := range integers {
		ids = append(ids, i.bits...)
	}
	states := expandStates(ids, unpackBDD(p))

	// Split assignment into booleans and integers.
	result := make(States, 0, len(states))
	for _, assignment := range states {
		bools := make(map[string]bool)
		ints := make(map[string]uint, len(m.ints))

		// Extract integer values.
		if all {
			integers = m.ints
		}
		for _, i := range integers {
			// Compute value
			value := uint(0)
			for n, id := range i.bits {
				if assignment[id] {
					value += 1 << n
				}
				delete(assignment, id)
			}

			// Store value
			name := m.IntName(i)
			ints[name] = value
		}

		// Extract boolean values.
		if all {
			for id, v := range assignment {
				bools[m.Name(id)] = v
			}

		} else {
			for _, id := range booleans {
				bools[m.Name(id)] = assignment[id]
			}
		}

		// Check if a state with these values already exists.
		newState := &State{bools, ints}
		duplicate := false
		for _, state := range result {
			if newState.Equals(state) {
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
