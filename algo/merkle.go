// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package algo

import (
	"bytes"
	"hash"
	"math"
	"strconv"
)

// If you are having trouble deciphering the math, the simplest parts
// are described on Wikipedia:
// http://en.wikipedia.org/wiki/Binary_tree#Arrays. The rest have to
// do with thereoms about a full and complete binary tree, by which a
// Merkle Tree nearly behaves. You can see
// http://courses.cs.vt.edu/~cs3114/Fall09/wmcquain/Notes/T03a.BinaryTreeTheorems.pdf
// for some explanations of the tree.
//
// We use an array because in the worst case, we have to go one level
// deep for just one additional leaf which nearly doubles the size of
// the array. From a memory perspective though, it only adds a small
// portion of pointers compared to storing pointers to the parent and
// children. For example, fiven a Tree with 9 leafs, the array size
// would be 2*2^4-1 = 31 while the pointer based method would requite
// 20*3 = 60 pointers. Additionally, leafs can be found in constant
// time instead of logrithmic time. This is just one of those cases
// where an array based implementation is better.

// MerkleTree is an implementation of a Merkle tree (see:
// http://en.wikipedia.org/wiki/Merkle_tree). You create it with the
// NewMerkleTree* functions.
type MerkleTree [][]byte

// NewMerkleTree creates a Merkle tree from the given data using the
// given hash.
func NewMerkleTree(data [][]byte, h hash.Hash) MerkleTree {
	hs := make([][]byte, len(data))
	for i, d := range data {
		h.Reset()
		h.Write(d)
		hs[i] = h.Sum(nil)
	}
	return NewMerkleTreeFromHashes(hs, h)
}

// NewMerkleTreeFromHashes creates a Merkle tree from the given hashes
// of data. This is a shortcut if the hashes of the data is already
// known so they won't need to be recreated. The same hash used to
// hash the data should be given.
func NewMerkleTreeFromHashes(data [][]byte, h hash.Hash) MerkleTree {
	if len(data) < 1 {
		return MerkleTree{}
	} else if len(data) < 2 {
		return MerkleTree{data[0]}
	}

	// Find the number of leaves and therewith derive the number of
	// nodes. The number of leaves is a value greater than or equal to
	// len(data) that is a power of 2.
	l := uint(len(data)) - 1
	for x := uint(1); x <= uint(strconv.IntSize/2); x *= 2 {
		l |= l >> x
	}
	l++
	mt := make(MerkleTree, 2*l-1)
	height := uint(math.Ceil(math.Log2(float64(2*l-1)))) - 1

	// Add the leaf then interior nodes.
	copy(mt[int(math.Pow(2.0, float64(height)))-1:], data)
	for x := int(height) - 1; x >= 0; x-- {
		nodes := int(math.Pow(2.0, float64(x)))
		pos := int(math.Pow(2.0, float64(x))) - 1
		for y := 0; y < nodes; y++ {
			i := pos + y
			if mt[2*i+1] == nil {
				break
			}
			h.Reset()
			h.Write(mt[2*i+1])
			if mt[2*i+2] != nil {
				h.Write(mt[2*i+2])
			}
			mt[i] = h.Sum(nil)
		}
	}

	return mt
}

// Verify verifies the given hash value against the root of this
// Merkle tree.
func (mt *MerkleTree) Verify(sum []byte) bool {
	if len(*mt) < 1 {
		return false
	}
	return bytes.Compare((*mt)[0], sum) == 0
}

// Root returns the root of this MerkleTree.
func (mt *MerkleTree) Root() []byte {
	if len(*mt) < 1 {
		return nil
	}
	return (*mt)[0]
}

// MerkleProofNode represents a node in a merkle tree that is used
// when proving membership of a leaf node.
type MerkleProofNode struct {
	Sum          []byte
	Sibling      *MerkleProofNode
	Parent       *MerkleProofNode
	SiblingFirst bool // True if the sibling is the left childe of the
	// parent (should be first part of the hash).
}

// Proof returns a proof for the given leaf node. If sum is not a leaf
// node, nil is returned. Otherwise, the lineage of that leaf to the
// root node is returned.
func (mt *MerkleTree) Proof(sum []byte) *MerkleProofNode {
	// Find the start of the leaf nodes.
	height := int(math.Ceil(math.Log2(float64(len(*mt)+1)))) - 1
	pos := int(math.Pow(2.0, float64(height))) - 1
	for ; pos < len(*mt) && (*mt)[pos] != nil; pos++ {
		if bytes.Compare((*mt)[pos], sum) == 0 {
			break
		}
	}

	// If we didn't find the sum, we can't continue.
	if pos >= len(*mt) || (*mt)[pos] == nil {
		return nil
	}

	// Fill out the lineage.
	cur := &MerkleProofNode{
		Sum: sum,
	}
	leaf := cur
	for pos != 0 {
		// Find cur's parent and create the node.
		pos = (pos - 1) / 2
		cur.Parent = &MerkleProofNode{
			Sum: (*mt)[pos],
		}

		// Find cur's sibling and create the node.
		cur.Sibling = &MerkleProofNode{
			Parent:  cur.Parent,
			Sibling: cur,
			Sum:     (*mt)[2*pos+2],
		}
		// If the sums are equal, the sibling is on the left and we need
		// to update our data structure.
		if bytes.Compare(cur.Sum, cur.Sibling.Sum) == 0 {
			cur.Sibling.Sum = (*mt)[2*pos+1]
			cur.SiblingFirst = true
		}
		cur = cur.Parent
	}

	return leaf
}

// VerifyMerkleProof uses the given proof to verify that the lineage
// given is valid. In order for the proof to succeed the leaf of the
// proof must be equal to the given leaf, the hashes up the proof must
// be align with the hashes built with the given hash, and the
// resulting root of the proof must be equal to the given root.
func VerifyMerkleProof(proof *MerkleProofNode, leaf, root []byte, h hash.Hash) bool {
	// Sanity Checks.
	if proof == nil {
		return false
	}
	if bytes.Compare(proof.Sum, leaf) != 0 {
		return false
	}

	// Traverse the proof verifying the sums as we go up.
	var sum []byte
	for cur := proof; cur.Parent != nil; cur = cur.Parent {
		h.Reset()
		// If the Sibling is on the left, it should be summed first.
		if cur.SiblingFirst {
			h.Write(cur.Sibling.Sum)
		}
		h.Write(cur.Sum)
		// Add the sibling, if we have one and didn't do it first.
		if cur.Sibling != nil && !cur.SiblingFirst {
			h.Write(cur.Sibling.Sum)
		}
		// Verify this sum against the parent sum.
		sum = h.Sum(nil)
		if bytes.Compare(sum, cur.Parent.Sum) != 0 {
			return false
		}
	}
	// Make sure the root is equal.
	if bytes.Compare(sum, root) != 0 {
		return false
	}
	return true
}
