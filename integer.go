package ctl

import "fmt"

// Integer represents a bounded integer in the range 0..(2^(len(bits))-1).
type Integer struct {
	bits       []uint // Variable identifiers or bit values for each bit
	variable   bool   // Is this a variable?
	constraint *BDD   // Constraint on the bits (if any)
}

// Int creates a new integer constant.
func Int(value uint) *Integer {
	n := bitcount(value)
	bits := make([]uint, n)
	for i := range bits {
		bits[i] = (value >> i) & 1
	}
	return &Integer{bits, false, True}
}

// Next returns the integer that identifies i in the next step.
func (i *Integer) Next() *Integer {
	if i.variable {
		bits := make([]uint, len(i.bits))
		for j := range bits {
			bits[j] = varNextID(i.bits[j])
		}
		return &Integer{bits, true, i.constraint.Next()}
	}
	return i
}

// ConstValue returns the integer value of non-variables.
func (i *Integer) ConstValue() uint {
	sum := uint(0)
	for n, v := range i.bits {
		if v != 0 {
			sum += 1 << n
		}
	}
	return sum
}

// Bit returns a BDD representing the n-th bit.
func (i *Integer) Bit(n int) *BDD {
	if n >= len(i.bits) {
		return False
	}
	if i.variable {
		return Node(i.bits[n], True, False)
	}
	if i.bits[n] == 0 {
		return False
	}
	return True
}

// Add returns the integer that is the result of adding i and j.
func (i *Integer) Add(j *Integer, m *Model) *Integer {
	if !i.variable && !j.variable {
		return Int(i.ConstValue() + j.ConstValue())
	}

	// i + j = k
	name := fmt.Sprintf("add(%v,%v)", m.IntName(i), m.IntName(j))
	size := max(len(i.bits), len(j.bits)) + 1
	k := m.Bin(name, uint(size))

	// Implicitly compute binary addition.
	add := k.Bit(0).Eq(i.Bit(0).Xor(j.Bit(0)))
	for n := 1; n < size; n++ {
		i0, j0, k0 := i.Bit(n-1), j.Bit(n-1), k.Bit(n-1)
		i1, j1, k1 := i.Bit(n), j.Bit(n), k.Bit(n)
		carry := i0.And(j0).Or(i0.Xor(j0).And(k0.Neg()))
		add = add.And(k1.Eq(i1.Xor(j1).Xor(carry)))
	}

	// Constrain k by the addition of i and j and the constraints on i and j.
	k.constraint = i.constraint.And(j.constraint).And(add)
	return k
}

// Eq returns a BDD that is true when i == j.
func (i *Integer) Eq(j *Integer) *BDD {
	size := max(len(i.bits), len(j.bits))
	eq := True
	for n := 0; n < size; n++ {
		eq = eq.And(i.Bit(n).Eq(j.Bit(n)))
	}
	return i.constraint.And(j.constraint).And(eq)
}

// Lt returns a BDD that is true when i < j.
func (i *Integer) Lt(j *Integer) *BDD {
	size := max(len(i.bits), len(j.bits))
	return i.constraint.And(j.constraint).And(i.leq(j, size-1, true))
}

// Leq returns a BDD that is true when i <= j.
func (i *Integer) Leq(j *Integer) *BDD {
	size := max(len(i.bits), len(j.bits))
	return i.constraint.And(j.constraint).And(i.leq(j, size-1, false))
}

func (i *Integer) leq(j *Integer, n int, neq bool) *BDD {
	a, b := i.Bit(n), j.Bit(n)
	lt := a.Neg().And(b)
	// At the 0-th bit there is a difference between <= and <.
	if n == 0 {
		if neq {
			return lt
		}
		// Note that <= === ->
		return a.Imply(b)
	}
	// Either the n-th bit is <, or the inequality is confirmed later.
	return lt.Or(a.Eq(b).And(i.leq(j, n-1, neq)))
}
