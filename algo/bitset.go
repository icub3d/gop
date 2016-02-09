// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package algo

import (
	"sort"
	"strconv"
)

// BitSet is a set of bit that can be turned on/off. They are commonly
// used for space efficiency in data structures like bloom filters.
type BitSet []int

// NewBitSet creates a BitSize of size n bits.
func NewBitSet(n uint) BitSet {
	return make(BitSet, (n/strconv.IntSize)+1)
}

// SetInt is a convienance function for using ints instead of
// uints. It is equivalient to bs.Set(uint(n)). Set is a noop if n <
// 0.
func (bs *BitSet) SetInt(n int) {
	if n < 0 {
		return
	}

	bs.Set(uint(n))
}

// Set turns on the given bit in this BitSet. More space is allocated
// to the BitSet if n is larger than the current size of this
// BitSet.
func (bs *BitSet) Set(n uint) {
	// Resize if necessary.
	if len(*bs) < int(n/strconv.IntSize)+1 {
		nbs := make(BitSet, (n/strconv.IntSize)+1)
		copy(nbs, *bs)
		*bs = nbs
	}

	(*bs)[n/strconv.IntSize] |= 1 << (uint(n) % strconv.IntSize)
}

// UnsetInt is a convienance function for using ints instead of
// uints. It is equivalient to bs.Unset(uint(n)). Set is a noop if n <
// 0.
func (bs *BitSet) UnsetInt(n int) {
	if n < 0 {
		return
	}
	bs.Unset(uint(n))
}

// Unset clears the bit at n such that a call to IsSet(n) will return
// false.
func (bs *BitSet) Unset(n uint) {
	if len(*bs) >= int(n/strconv.IntSize)+1 {
		(*bs)[n/strconv.IntSize] &^= 1 << (uint(n) % strconv.IntSize)
	}
}

// IsSetInt is a convienance function for using ints instead of
// uints. It is equivalient to bs.IsSet(uint(n)).
func (bs *BitSet) IsSetInt(n int) bool {
	return bs.IsSet(uint(n))
}

// IsSet returns true if n has been Set(), false otherwise.
func (bs *BitSet) IsSet(n uint) bool {
	if n < 0 || len(*bs) < int(n/strconv.IntSize)+1 {
		return false
	}
	return (*bs)[n/strconv.IntSize]&(1<<(uint(n)%strconv.IntSize)) != 0
}

// Complement returns the complement of the BitSet.
func (bs BitSet) Complement() BitSet {
	nbs := make(BitSet, len(bs))
	for n, w := range bs {
		nbs[n] = ^w
	}
	return nbs
}

// Union updates this BitSet to include all of the set values in the
// given BitSet.
func (bs *BitSet) Union(obs BitSet) {
	// Resize if necessary.
	if len(*bs) < len(obs) {
		nbs := make(BitSet, len(obs))
		copy(nbs, *bs)
		*bs = nbs
	}
	for x := 0; x < len(obs); x++ {
		(*bs)[x] |= obs[x]
	}
}

// Intersect updates this BitSet to include only those values that are
// both in this BitSet and the given BitSet.
func (bs *BitSet) Intersect(obs BitSet) {
	min := MinInt(len(*bs), len(obs))
	for x := 0; x < min; x++ {
		(*bs)[x] &= obs[x]
	}
	for x := min; x < len(*bs); x++ {
		(*bs)[x] = 0
	}
}

// Difference removes all of the set values in this BitSet that are in
// the given BitSet.
func (bs *BitSet) Difference(obs BitSet) {
	min := MinInt(len(*bs), len(obs))
	for x := 0; x < min; x++ {
		(*bs)[x] &^= obs[x]
	}
}

// DifferenceBitSets creates a new BitSet whose bits are set only if
// they are in the first BitSet but in none of the rest of the
// BitSets. For example, if the given sets are A, B, and C, then this
// returns a new BitSet that is equivalent to A - B - C which in set
// theory translates to ((A - B) - C) or A - (B union C).
func DifferenceBitSets(bss ...BitSet) BitSet {
	if len(bss) < 1 {
		return NewBitSet(1)
	}
	nbs := make(BitSet, len(bss[0]))
	copy(nbs, bss[0])
	nbs.Difference(UnionBitSets(bss[1:]...))
	return nbs
}

// IntersectBitSets creates a new BitSet whose bits are set only if
// that bit is set in all of the given BitSets.
func IntersectBitSets(bss ...BitSet) BitSet {
	if len(bss) < 1 {
		return NewBitSet(1)
	}
	sort.Sort(bitSets(bss))
	nbs := make(BitSet, len(bss[0]))
	for x := 0; x < len(nbs); x++ {
		nbs[x] = bss[0][x]
		for y := 1; y < len(bss); y++ {
			nbs[x] &= bss[y][x]
		}
	}
	return nbs
}

// UnionBitSets creates a new BitSet with all the bits set from the given BitSets.
func UnionBitSets(bss ...BitSet) BitSet {
	if len(bss) < 1 {
		return NewBitSet(1)
	}
	sort.Sort(bitSets(bss))
	nbs := make(BitSet, len(bss[len(bss)-1]))
	n := 0 // The position in the list that we are still using.
	for x := 0; x < len(bss[len(bss)-1]); x++ {
		// We can skip the smaller BitSets.
		for len(bss[n]) <= x {
			n++
		}
		// Union the rest of the BitSet's for this word.
		for y := n; y < len(bss); y++ {
			nbs[x] |= bss[y][x]
		}
	}
	return nbs
}

// bitSets is used for sorting an array of BitSets
type bitSets []BitSet

func (a bitSets) Len() int           { return len(a) }
func (a bitSets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bitSets) Less(i, j int) bool { return len(a[i]) < len(a[j]) }
