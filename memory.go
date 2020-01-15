package ctl

import (
	"sync"
)

// Total number of cached elements.
var _cacheCounter = 0

// Global BDD lookup table. Since Go does not have weak pointers, there is no
// way to determine if a pointer is still in used outside this map except for
// a wrapping struct with lots of overhead. I decided to do no housekeeping.
var _lookup = new(sync.Map)

// Register new BDD node reference (and get unique pointer).
func registerNodeRef(node BDD) *BDD {
	ref, new := _lookup.LoadOrStore(node, &node)
	if new {
		_cacheCounter++
	}
	return ref.(*BDD)
}

// Cache of applying an operator to two BDDs.
var _applyCache = new(sync.Map)

type applyKey struct {
	op   uint
	p, q *BDD
}

// Return cashed operator application.
func applyCached(op uint, p *BDD, q *BDD) *BDD {
	key := applyKey{op, p, q}
	if result, in := _applyCache.Load(key); in {
		return result.(*BDD)
	}

	// Compute result.
	result := p.Apply(op, q)
	_applyCache.Store(key, result)
	_cacheCounter++
	return result
}
