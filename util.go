package ctl

import (
	"math"
)

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func bitcount(i uint) uint {
	if i < 1 {
		return 1
	}
	return uint(math.Floor(math.Log2(float64(i)) + 1))
}

// ByStringLt defines a sort interface for ordinary string sorting.
type ByStringLt []string

func (a ByStringLt) Len() int           { return len(a) }
func (a ByStringLt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStringLt) Less(i, j int) bool { return a[i] < a[j] }
