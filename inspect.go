package ctl

// LeastSteps determines the least number of steps starting at an accepted
// initial state that satisfy some specification that was computed by model.
func LeastSteps(init *BDD, sets []*BDD) int {
	for i := 1; i < len(sets); i++ {
		if init.Intersects(sets[i]) {
			return i
		}
	}
	return -1
}

// GenerateExample generates a smallest path from an accepted initial state that
// satisfies the computed specification.
func GenerateExample(m *Model, init *BDD, sets []*BDD) []*State {
	// Find starting point (first set that intersects init).
	path := make([]*State, 0)
	beam := False
	i := 1
	for ; i < len(sets); i++ {
		beam = init.And(sets[i])
		if beam != False {
			// We found a starting point.
			break
		}
	}
	if beam == False {
		return []*State{}
	}

	// Go back to the goal.
	for ; i >= 0; i-- {
		// Unpack beam and pick one state (this could be done much quicker).
		states := expandStates(m.vars, false, unpackBDD(beam))
		if len(states) == 0 {
			panic("beam is empty")
		}

		// Add state to path and create BDD that only accepts this state.
		state := states[0]
		path = append(path, processState(m, state, false))
		s := True
		for id, b := range state {
			if !b {
				s = s.And(Node(id, True, False))
			} else {
				s = s.And(Node(id, False, True))
			}
		}

		// Create BDD that contains all sets that are reachable from this state
		// using only one transition.
		if i > 0 {
			beam = m.EXInv(s, sets[i-1])
		}
	}

	return path
}
