package ctl

// BDD represents a Binary pecision piagram.
type BDD struct {
	Value bool
	Var   *Variable
	True  *BDD
	False *BDD
}

// True is a BDD true leaf.
var True = &BDD{true, nil, nil, nil}

// False is a BDD false leaf.
var False = &BDD{false, nil, nil, nil}

// Node returns a BDD node.
func Node(v *Variable, t *BDD, f *BDD) *BDD {
	if t.Equals(f) {
		return t
	}
	return registerNodeRef(BDD{false, v, t, f})
}

// Node checks if the given BDD is a node.
func (p *BDD) Node() bool {
	return p.Var != nil
}

// Equals compares this BDD with the given BDD.
func (p *BDD) Equals(q *BDD) bool {
	// Pointing to the same instance?
	if p == q {
		return true
	}
	// Referencing the same variable? (or both a leaf)
	if p.Var != q.Var {
		return false
	}
	// Same branches or leaf values?
	if p.Node() {
		return p.True.Equals(q.True) && p.False.Equals(q.False)
	}
	return p.Value == q.Value
}

// Next returns a BDD with all next variable identifiers. By convention all
// variable ID's are left-shifted 1 place. The same variable in the next state
// is encoded by setting the first bit to 1.
func (p *BDD) Next() *BDD {
	if p.Node() {
		return Node(p.Var.Next(), p.True.Next(), p.False.Next())
	}
	return p
}

// Norm returns a BDD where all next variables are reverted to normal (e.g.
// p.Next().Norm() = p).
func (p *BDD) Norm() *BDD {
	if p.Node() {
		return Node(p.Var.Norm(), p.True.Norm(), p.False.Norm())
	}
	return p
}

// Set returns a BDD where the variable v is set to true/false.
func (p *BDD) Set(v *Variable, value bool) *BDD {
	if p.Var == v {
		if value {
			return p.True
		}
		return p.False
	}
	if p.Node() {
		return Node(p.Var, p.True.Set(v, value), p.False.Set(v, value))
	}
	return p
}

// Apply applies the given binary operator to the BDDs p and q. The binary
// operator is represented as a truth table for [00, 01, 10, 11] in bit flags.
func (p *BDD) Apply(op uint, q *BDD) *BDD {
	// Push operator downward.
	if p.Node() || q.Node() {
		if p.Var == q.Var {
			return Node(p.Var,
				applyCached(op, p.True, q.True),
				applyCached(op, p.False, q.False))
		} else if p.Node() && p.Var.Lt(q.Var) || !q.Node() {
			return Node(p.Var,
				applyCached(op, p.True, q),
				applyCached(op, p.False, q))
		} else { // if q.Node() && q.Var.Lt(p.Var) || !p.Node() {
			return Node(q.Var,
				applyCached(op, p, q.True),
				applyCached(op, p, q.False))
		}
	}
	// Or evaluate operator.
	i := 0
	if p.Value {
		i = 2
	}
	if q.Value {
		i++
	}
	if (op>>(3-i))&1 == 0 {
		return False
	}
	return True
}

// Neg this
func (p *BDD) Neg() *BDD {
	return p.Imply(False)
}

// Imply q
func (p *BDD) Imply(q *BDD) *BDD {
	return p.Apply(0b1101, q)
}

// And q
func (p *BDD) And(q *BDD) *BDD {
	return p.Apply(0b0001, q)
}

// Or q
func (p *BDD) Or(q *BDD) *BDD {
	return p.Apply(0b0111, q)
}

// Eq q
func (p *BDD) Eq(q *BDD) *BDD {
	return p.Apply(0b1001, q)
}

// Xor q
func (p *BDD) Xor(q *BDD) *BDD {
	return p.Apply(0b0110, q)
}

// Contains determines if all true assignments in q are also true in this BDD.
func (p *BDD) Contains(q *BDD) bool {
	return q.Imply(p) == True
}

// Intersects determines if there exists a truth assignment (state) that
// satisfies both p and q (this is more liberal than Contains).
func (p *BDD) Intersects(q *BDD) bool {
	return q.And(p) != False
}

// Exists determines if there exists a satisfying assignment for variable v.
func (p *BDD) Exists(v *Variable) *BDD {
	return p.Set(v, true).Or(p.Set(v, false))
}
