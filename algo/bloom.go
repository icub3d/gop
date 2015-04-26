// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package algo

import (
	"encoding/binary"
	"hash/fnv"
	"math"
)

// BloomFilter is a representation of the bloom filter data
// structure. You create one by calling NewBloomFilter or
// NewBloomFilterEstimate.
type BloomFilter struct {
	m  uint   // The size of the BitSet.
	k  uint   // The number of hashes.
	n  uint   // The numer of Add()'s (used for estimating false positives).
	bs BitSet // The BitSet
}

// NewBloomFilter creates a bloom filter of size m and with k
// hashes. For more details on what that means, see:
// http://en.wikipedia.org/wiki/Bloom_filter.
func NewBloomFilter(m uint, k uint) *BloomFilter {
	return &BloomFilter{
		m:  m,
		k:  k,
		bs: NewBitSet(m),
	}
}

// NewBloomFilterEstimate creates a bloom filter with a size and
// number of hashes based on the given estimated number of values
// being added and the desired false positive rate. You should choose
// your false positive rate carefully based on your data set and
// needs. The space efficiency grows quickly the smaller your failure
// rate.
func NewBloomFilterEstimate(n uint, p float64) *BloomFilter {
	// These are based on the equations found at:
	// http://en.wikipedia.org/wiki/Bloom_filter#Optimal_number_of_hash_functions.
	m := uint(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2))
	k := uint(math.Ceil(float64(m) / float64(n) * math.Log(2)))
	return NewBloomFilter(m, k)
}

// Add inserts the given value into the Bloom filter. Calls to
// Exists(data) will now always return true.
func (bf *BloomFilter) Add(data []byte) {
	h := fnv.New64()
	h.Write(data)
	s := h.Sum(nil)
	l := uint(binary.BigEndian.Uint32(s[0:4]))
	u := uint(binary.BigEndian.Uint32(s[4:8]))
	for x := uint(0); x < bf.k; x++ {
		bf.bs.Set((l + u*x) % bf.m)
	}
	bf.n++
}

// Exists determines if the given value is likely in the bloom
// filter. There is a possibility that, based on the number of values
// added and the size of the bloom filter, Add(data) was never called.
func (bf *BloomFilter) Exists(data []byte) bool {
	h := fnv.New64()
	h.Write(data)
	s := h.Sum(nil)
	l := uint(binary.BigEndian.Uint32(s[0:4]))
	u := uint(binary.BigEndian.Uint32(s[4:8]))
	for x := uint(0); x < bf.k; x++ {
		if !bf.bs.IsSet((l + u*x) % bf.m) {
			return false
		}
	}
	return true
}

// FalsePositives estimates the false positive rate of this bloom
// filter based on the formula found at
// http://en.wikipedia.org/wiki/Bloom_filter#Probability_of_false_positives.
func (bf *BloomFilter) FalsePositives() float64 {
	return math.Pow(float64(1-math.Pow(math.E, float64(-1*int(bf.k*bf.n/bf.m)))), float64(bf.k))
}
