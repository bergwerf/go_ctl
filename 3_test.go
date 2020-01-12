package ctl

import (
	"fmt"
	"testing"
)

// TestNetworkDeadlock tests a large network simulation to find deadlocks.
func TestNetworkDeadlock(t *testing.T) {
	// All channels (start to end node, first node and first channel are 1)
	connections := [][]int{
		{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 6}, {6, 7}, {7, 8}, {8, 9}, {9, 10},
		{10, 11}, {11, 12}, {12, 13}, {13, 14}, {14, 15}, {15, 16}, {16, 1},
		{3, 17}, {17, 3}, {7, 17}, {17, 7}, {11, 17}, {17, 11}, {15, 17}, {17, 15},
		{16, 2}, {4, 6}, {8, 10},
	}

	// route[receiving node][destination node] := forward channel
	routes := [][]int{
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{2, 0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		{17, 17, 0, 3, 3, 3, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17},
		{26, 26, 26, 0, 4, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26},
		{5, 5, 5, 5, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
		{6, 6, 6, 6, 6, 0, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
		{19, 19, 19, 19, 19, 19, 0, 7, 7, 7, 19, 19, 19, 19, 19, 19, 19},
		{27, 27, 27, 27, 27, 27, 27, 0, 8, 27, 27, 27, 27, 27, 27, 27, 27},
		{9, 9, 9, 9, 9, 9, 9, 9, 0, 9, 9, 9, 9, 9, 9, 9, 9},
		{10, 10, 10, 10, 10, 10, 10, 10, 10, 0, 10, 10, 10, 10, 10, 10, 10},
		{21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 0, 11, 11, 11, 21, 21, 21},
		{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 0, 12, 12, 12, 12, 12},
		{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 0, 13, 13, 13, 13},
		{14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 0, 14, 14, 14},
		{15, 15, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 0, 15, 23},
		{16, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 0, 25},
		{24, 24, 18, 18, 18, 18, 20, 20, 20, 20, 22, 22, 22, 22, 24, 24, 0},
	}

	// A problem instance
	type instance struct {
		// Main nodes (nodes that can send and receive)
		main []int
		// Least number of steps to a deadlock (or -1 if no deadlock is reachable)
		deadlock int
	}

	// I solved this problem with NuSMV some time ago (each instance takes several
	// seconds).
	instances := []instance{
		{[]int{1, 5, 9, 13}, -1},
		{[]int{2, 4, 6}, 9},
		{[]int{1, 3, 5, 15}, -1},
		{[]int{11, 13, 15}, -1},
		{[]int{11, 12, 13, 15}, 8},
		{[]int{1, 8, 10}, 12},
		{[]int{5, 12, 14}, 14},
		{[]int{5, 11, 14}, -1},
		{[]int{1, 2, 3, 4, 5}, -1},
	}

	// For now even just setting up the BDD describing all transitions takes
	// very long (I did not let it finish yet).
	t.Skip("This test requires serious optimizations to run")

	for _, instance := range instances {
		result := findDeadlock(connections, routes, instance.main)
		if result != instance.deadlock {
			t.Errorf("Expected a deadlock in %v steps", instance.deadlock)
		}
	}
}

// Given connection channels, routes, and main nodes. Compute how many steps it
// takes to reach a deadlock starting at an empty network. If no deadlock is
// reachable the function returns -1.
func findDeadlock(connections [][]int, routes [][]int, main []int) int {
	// Convert connections and routes to 0-based indices.
	connections = intMatSub(connections, 1)
	routes = intMatSub(routes, 1)
	nodeCount := len(routes)
	mainNodes := make([]uint, len(main))
	for i, m := range main {
		mainNodes[i] = uint(m) - 1
	}

	// Collect indices of ingoing channels for each node.
	nodeInputs := make([][]int, nodeCount)
	for i, conn := range connections {
		to := conn[1]
		nodeInputs[to] = append(nodeInputs[to], i)
	}

	// Create one integer value per channel.
	m := NewModel()
	channels := make([]*Integer, len(connections))
	for i := range channels {
		channels[i] = m.Int(fmt.Sprintf("c%v", i), uint(nodeCount-1))
	}

	// Identity transition for all but the selected channels.
	copyExcept := func(exclude []int) *BDD {
		bdd := True
		for i, channel := range channels {
			excluded := false
			for _, j := range exclude {
				if j == i {
					excluded = true
					break
				}
			}
			if !excluded {
				bdd = bdd.And(channel.Next().Eq(channel))
			}
		}
		return bdd
	}

	// Packet send transitions.
	for _, from := range mainNodes {
		for _, to := range mainNodes {
			if from == to {
				continue
			}
			i := routes[from][to]
			m.Add(channels[i].Eq(Int(0)),
				channels[i].Next().Eq(Int(to)).And(copyExcept([]int{i})))
		}
	}

	// Packet receive transitions.
	canReceive := False
	for _, to := range mainNodes {
		for _, i := range nodeInputs[to] {
			condition := channels[i].Eq(Int(to))
			canReceive = canReceive.Or(condition)
			m.Add(condition,
				channels[i].Next().Eq(Int(0)).And(copyExcept([]int{i})))
		}
	}

	// Packet forward transitions.
	canForward := False
	for i := 0; i < len(connections); i++ {
		current := connections[i][1]
		for _, to := range mainNodes {
			// Packets to the target node are not forwarded.
			if to == uint(current) {
				continue
			}
			// Forward from channel i to channel j.
			j := routes[current][to]
			ci, cj := channels[i], channels[j]
			condition := ci.Eq(Int(to)).And(cj.Eq(Int(0)))
			canForward = canForward.Or(condition)
			m.Add(condition,
				ci.Next().Eq(Int(0)).And(cj.Next().Eq(Int(to))).And(
					copyExcept([]int{i, j})))
		}
	}

	// Define a functioning network (no deadlock): the network is empty or packets
	// can be received or packets can be forwarded.
	emptyNetwork := True
	for _, ch := range channels {
		emptyNetwork = emptyNetwork.And(ch.Eq(Int(0)))
	}
	noDeadlock := emptyNetwork.Or(canReceive).Or(canForward)

	// Check if it is possible to end up in a deadlock from an empty network.
	deadlock := noDeadlock.Neg()
	sets := m.EF(deadlock)
	return LeastSteps(emptyNetwork, sets)
}

// Subtract n from integer matrix (to convert to 0-based values).
func intMatSub(mat [][]int, n int) [][]int {
	copy := make([][]int, len(mat))
	for i, row := range mat {
		copy[i] = make([]int, len(row))
		for j, k := range row {
			copy[i][j] = k - n
		}
	}
	return copy
}
