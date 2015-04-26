// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package algo

import (
	"fmt"
	"testing"
)

func TestBitSet(t *testing.T) {
	bs := NewBitSet(1)
	// Set several values, including values that will increase the size.
	for x := 0; x < 1024; x += 2 {
		bs.SetInt(x)
	}
	bs.SetInt(-1)

	// Unset some of the values we just set.
	for x := 0; x < 1024; x += 4 {
		bs.UnsetInt(x)
	}
	bs.UnsetInt(-1)

	// Check all the bits.
	for x := -1; x < 2048; x++ {
		r := bs.IsSetInt(x)
		e := true
		if x < 0 || x >= 1024 || (x%2) == 1 || (x%4) == 0 {
			e = false
		}

		if r != e {
			t.Errorf("bs.IsSetInt(%v) == %v but wanted %v", x, r, e)
		}
	}

}

func TestBitSetComplement(t *testing.T) {
	bs := NewBitSet(1)
	for x := 0; x < 1024; x += 2 {
		bs.SetInt(x)
	}
	c := bs.Complement()
	for x := 0; x < 1024; x += 2 {
		if bs.IsSetInt(x) == c.IsSetInt(x) {
			t.Errorf("bs.Complement(%v) == c.Complement(%v) == %v, expected them to differ",
				x, x, bs.IsSetInt(x))
		}
	}
}

func TestBitSetUnion(t *testing.T) {
	tests := []struct {
		sets     [][]int
		max      int
		expected []int
	}{
		// First is smaller.
		{
			sets: [][]int{
				[]int{96, 98},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{96, 98, 1022, 1024, 4096},
		},
		// Second is smaller.
		{
			sets: [][]int{
				[]int{96, 1022, 1024, 4096},
				[]int{96, 98},
			},
			max:      4096,
			expected: []int{96, 98, 1022, 1024, 4096},
		},
		// Both the same size.
		{
			sets: [][]int{
				[]int{48, 96, 1024, 33},
				[]int{96, 98, 102, 33},
			},
			max:      4096,
			expected: []int{33, 48, 96, 98, 102, 1024},
		},
	}

	for k, test := range tests {
		bss := make([]BitSet, len(test.sets))
		for x, ns := range test.sets {
			bss[x] = NewBitSet(1)
			for _, n := range ns {
				bss[x].SetInt(n)
			}
		}
		bss[0].Union(bss[1])
		for x := 0; x < test.max; x++ {
			if intInArray(x, test.expected) {
				if !bss[0].IsSetInt(x) {
					t.Errorf("Test %v: r.IsSetInt(%v) == false, expected true", k, x)
				}
			} else if bss[0].IsSetInt(x) {
				t.Errorf("Test %v: r.IsSetInt(%v) == true, expected false", k, x)
			}
		}
	}
}

func TestBitSetIntersect(t *testing.T) {
	tests := []struct {
		sets     [][]int
		max      int
		expected []int
	}{
		// First is smaller.
		{
			sets: [][]int{
				[]int{96, 98},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{96},
		},
		// Second is smaller.
		{
			sets: [][]int{
				[]int{96, 1022, 1024, 4096},
				[]int{96, 98},
			},
			max:      4096,
			expected: []int{96},
		},
		// Both the same size.
		{
			sets: [][]int{
				[]int{48, 96, 1024, 33},
				[]int{96, 98, 102, 33},
			},
			max:      4096,
			expected: []int{33, 96},
		},
	}

	for k, test := range tests {
		bss := make([]BitSet, len(test.sets))
		for x, ns := range test.sets {
			bss[x] = NewBitSet(1)
			for _, n := range ns {
				bss[x].SetInt(n)
			}
		}
		bss[0].Intersect(bss[1])
		for x := 0; x < test.max; x++ {
			if intInArray(x, test.expected) {
				if !bss[0].IsSetInt(x) {
					t.Errorf("Test %v: r.IsSetInt(%v) == false, expected true", k, x)
				}
			} else if bss[0].IsSetInt(x) {
				t.Errorf("Test %v: r.IsSetInt(%v) == true, expected false", k, x)
			}
		}
	}
}

func TestBitSetDifference(t *testing.T) {
	tests := []struct {
		sets     [][]int
		max      int
		expected []int
	}{
		// First is smaller.
		{
			sets: [][]int{
				[]int{96, 98},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{98},
		},
		// Second is smaller.
		{
			sets: [][]int{
				[]int{96, 98, 1022, 1024, 4096},
				[]int{96, 98},
			},
			max:      4096,
			expected: []int{1022, 1024, 4096},
		},
		// Both the same size.
		{
			sets: [][]int{
				[]int{48, 96, 1024, 33},
				[]int{96, 98, 102, 33},
			},
			max:      4096,
			expected: []int{48, 1024},
		},
	}

	for k, test := range tests {
		bss := make([]BitSet, len(test.sets))
		for x, ns := range test.sets {
			bss[x] = NewBitSet(1)
			for _, n := range ns {
				bss[x].SetInt(n)
			}
		}
		bss[0].Difference(bss[1])
		for x := 0; x < test.max; x++ {
			if intInArray(x, test.expected) {
				if !bss[0].IsSetInt(x) {
					t.Errorf("Test %v: r.IsSetInt(%v) == false, expected true", k, x)
				}
			} else if bss[0].IsSetInt(x) {
				t.Errorf("Test %v: r.IsSetInt(%v) == true, expected false", k, x)
			}
		}
	}
}

func TestUnionBitSets(t *testing.T) {
	tests := []struct {
		sets     [][]int
		max      int
		expected []int
	}{
		// Basic test.
		{
			sets: [][]int{
				[]int{48, 96, 1024},
				[]int{96, 98},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{48, 96, 98, 1022, 1024, 4096},
		},
		// All the same size.
		{
			sets: [][]int{
				[]int{48, 96, 1024, 33},
				[]int{96, 98, 102, 33},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{33, 48, 96, 98, 102, 1022, 1024, 4096},
		},
		// No sets.
		{
			sets:     [][]int{},
			max:      4096,
			expected: []int{},
		},
	}

	for k, test := range tests {
		bss := make([]BitSet, len(test.sets))
		for x, ns := range test.sets {
			bss[x] = NewBitSet(1)
			for _, n := range ns {
				bss[x].SetInt(n)
			}
		}
		r := UnionBitSets(bss...)
		for x := 0; x < test.max; x++ {
			if intInArray(x, test.expected) {
				if !r.IsSetInt(x) {
					t.Errorf("Test %v: r.IsSetInt(%v) == false, expected true", k, x)
				}
			} else if r.IsSetInt(x) {
				t.Errorf("Test %v: r.IsSetInt(%v) == true, expected false", k, x)
			}
		}
	}
}

func TestIntersectBitSets(t *testing.T) {
	tests := []struct {
		sets     [][]int
		max      int
		expected []int
	}{
		// Basic test.
		{
			sets: [][]int{
				[]int{48, 96, 1024},
				[]int{96, 98},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{96},
		},
		// All the same size.
		{
			sets: [][]int{
				[]int{48, 96, 1024, 33},
				[]int{96, 98, 102, 33},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{96},
		},
		// No sets.
		{
			sets:     [][]int{},
			max:      4096,
			expected: []int{},
		},
	}

	for k, test := range tests {
		bss := make([]BitSet, len(test.sets))
		for x, ns := range test.sets {
			bss[x] = NewBitSet(1)
			for _, n := range ns {
				bss[x].SetInt(n)
			}
		}
		r := IntersectBitSets(bss...)
		for x := 0; x < test.max; x++ {
			if intInArray(x, test.expected) {
				if !r.IsSetInt(x) {
					t.Errorf("Test %v: r.IsSetInt(%v) == false, expected true", k, x)
				}
			} else if r.IsSetInt(x) {
				t.Errorf("Test %v: r.IsSetInt(%v) == true, expected false", k, x)
			}
		}
	}
}

func TestDifferenceBitSets(t *testing.T) {
	tests := []struct {
		sets     [][]int
		max      int
		expected []int
	}{
		// Basic test.
		{
			sets: [][]int{
				[]int{48, 96, 1024},
				[]int{96, 98},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{48},
		},
		// All the same size.
		{
			sets: [][]int{
				[]int{48, 96, 1024, 33},
				[]int{96, 98, 102, 33},
				[]int{96, 1022, 1024, 4096},
			},
			max:      4096,
			expected: []int{48},
		},
		// No sets.
		{
			sets:     [][]int{},
			max:      4096,
			expected: []int{},
		},
		// Only one set.
		{
			sets: [][]int{
				[]int{48, 96},
			},
			max:      4096,
			expected: []int{48, 96},
		},
	}

	for k, test := range tests {
		bss := make([]BitSet, len(test.sets))
		for x, ns := range test.sets {
			bss[x] = NewBitSet(1)
			for _, n := range ns {
				bss[x].SetInt(n)
			}
		}
		r := DifferenceBitSets(bss...)
		for x := 0; x < test.max; x++ {
			if intInArray(x, test.expected) {
				if !r.IsSetInt(x) {
					t.Errorf("Test %v: r.IsSetInt(%v) == false, expected true", k, x)
				}
			} else if r.IsSetInt(x) {
				t.Errorf("Test %v: r.IsSetInt(%v) == true, expected false", k, x)
			}
		}
	}
}

func intInArray(i int, a []int) bool {
	for _, n := range a {
		if n == i {
			return true
		}
	}
	return false
}

func ExampleBitSet() {
	bs := NewBitSet(256)
	bs.Set(uint(124))
	bs.SetInt(145)
	bs.SetInt(512)
	for x := -1; x <= 1024; x++ {
		if bs.IsSetInt(x) {
			fmt.Println(x)
		}
	}
	// Output:
	// 124
	// 145
	// 512
}
