package ctl

import (
	"sync"
)

// Total number of cached elements (used for performance analysis).
var _cacheCounter = 0

// Global BDD lookup table. Since Go does not have weak pointers, there is no
// way to determine if a pointer is still in use outside of this map except
// using explicit reference counting. This is not yet implemented.
var _lookup = new(sync.Map)

// Register a new BDD node reference (and get unique pointer).
func registerNodeRef(node BDD) *BDD {
	ref, new := _lookup.LoadOrStore(node, &node)
	if new {
		_cacheCounter++
	}
	return ref.(*BDD)
}

// BDD application cache.
var _applyCache = new(sync.Map)

type applyKey struct {
	op   uint
	p, q *BDD
}

// Apply operator to BDDs, or return a cached result.
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
