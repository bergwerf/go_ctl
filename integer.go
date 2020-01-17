package ctl

import "fmt"

// Integer represents a bounded integer in the range 0..(2^(len(bits))-1).
type Integer struct {
	bits       []*Variable // Variables or bit values for each bit
	value      uint        // Constant integer value
	variable   bool        // Is this a variable?
	constraint *BDD        // Constraint on the bits (if any)
}

// Int creates a new integer constant.
func Int(value uint) *Integer {
	return &Integer{[]*Variable{}, value, false, True}
}

// Len returns the number of bits this integer uses.
func (i *Integer) Len() int {
	if i.variable {
		return len(i.bits)
	}
	return int(bitcount(i.value))
}

// Name returns the name of this integer.
func (i *Integer) Name() string {
	if i.variable {
		lbl := i.bits[0].Name
		return lbl[:len(lbl)-2]
	}
	return fmt.Sprintf("%v", i.value)
}

// Aux returns if this is an auxiliary integer.
func (i *Integer) Aux() bool {
	if i.variable {
		return i.bits[0].aux
	}
	return true
}

// Next returns the integer that identifies i in the next step.
func (i *Integer) Next() *Integer {
	if i.variable {
		nextBits := make([]*Variable, len(i.bits))
		for j, bit := range i.bits {
			nextBits[j] = bit.Next()
		}
		return &Integer{nextBits, 0, true, i.constraint.Next()}
	}
	return i
}

// Bit returns a BDD representing the n-th bit.
func (i *Integer) Bit(n int) *BDD {
	if i.variable {
		if n >= len(i.bits) {
			return False
		}
		return Node(i.bits[n], True, False)
	}
	if (i.value>>n)&1 == 0 {
		return False
	}
	return True
}

// Add returns the integer that is the result of adding i and j.
func (i *Integer) Add(j *Integer, m *Model) *Integer {
	return i.addNoCarry(j, m)
}

// Do not use carry bits.
func (i *Integer) addNoCarry(j *Integer, m *Model) *Integer {
	if !i.variable && !j.variable {
		return Int(i.value + j.value)
	}

	// i + j = k
	name := fmt.Sprintf("add(%v,%v)", i.Name(), j.Name())
	size := max(i.Len(), j.Len()) + 1
	k := m.bin(name, uint(size), true)

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

// Use carry bits.
func (i *Integer) addCarry(j *Integer, m *Model) *Integer {
	if !i.variable && !j.variable {
		return Int(i.value + j.value)
	}

	// i + j + c = k
	cName := fmt.Sprintf("carry(%v,%v)", i.Name(), j.Name())
	kName := fmt.Sprintf("add(%v,%v)", i.Name(), j.Name())
	size := max(i.Len(), j.Len()) + 1
	c := m.bin(cName, uint(size), true)
	k := m.bin(kName, uint(size), true)

	// i <-> j <-> k <-> c
	add := c.Bit(0).Eq(False) // First carry bit is 0.
	for n := 0; n < size; n++ {
		ib, jb, cb, kb := i.Bit(n), j.Bit(n), c.Bit(n), k.Bit(n)
		add = add.And(ib.Eq(jb).Eq(cb).Eq(kb)) // k = i + j + c (mod 2)
		// The next carry bit is 1 iff (ib /\ jb) \/ (ib /\ cb) \/ (jb /\ cb).
		add = add.And(c.Bit(n + 1).Eq(ib.And(jb).Or(ib.And(cb)).Or(jb.And(cb))))
	}

	// Constrain k by the addition of i and j and the constraints on i and j.
	k.constraint = i.constraint.And(j.constraint).And(add)
	return k
}

// Eq returns a BDD that is true when i == j.
func (i *Integer) Eq(j *Integer) *BDD {
	size := max(i.Len(), j.Len())
	eq := True
	for n := 0; n < size; n++ {
		eq = eq.And(i.Bit(n).Eq(j.Bit(n)))
	}
	return i.constraint.And(j.constraint).And(eq)
}

// Lt returns a BDD that is true when i < j.
func (i *Integer) Lt(j *Integer) *BDD {
	size := max(i.Len(), j.Len())
	return i.constraint.And(j.constraint).And(i.leq(j, size-1, true))
}

// Leq returns a BDD that is true when i <= j.
func (i *Integer) Leq(j *Integer) *BDD {
	size := max(i.Len(), j.Len())
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
