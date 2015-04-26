// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package algo

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestMerkleTree(t *testing.T) {
	data := [][]byte{
		[]byte("cat"),
		[]byte("dog"),
		[]byte("mouse"),
		[]byte("parrot"),
		[]byte("hamster"),
		[]byte("goat"),
	}

	// Build the tree and expected values for testing.
	h := sha256.New()
	result := NewMerkleTree(data, h)
	expected := make([][]byte, 15)
	for i, d := range data {
		h.Reset()
		h.Write(d)
		expected[i+7] = h.Sum(nil)
	}
	sets := [][]int{
		[]int{3, 5},
		[]int{1, 2},
		[]int{0, 0},
	}
	for _, s := range sets {
		for x := s[0]; x <= s[1]; x++ {
			h.Reset()
			h.Write(expected[x*2+1])
			h.Write(expected[x*2+2])
			expected[x] = h.Sum(nil)
		}
	}

	// Verify the size and all the values in the tree.
	if len(result) != len(expected) {
		t.Errorf("merkle tree sizes failed: %v %v", len(result), len(expected))
	}
	for i, e := range expected {
		if bytes.Compare(e, result[i]) != 0 {
			t.Errorf("merkle tree failed at %v:", i)
			t.Errorf("%v", e)
			t.Errorf("%v", result[i])
		}
	}

	// Test a tree with a single element.
	h.Reset()
	h.Write([]byte("cat"))
	expected = [][]byte{h.Sum(nil)}
	result = NewMerkleTree([][]byte{[]byte("cat")}, h)

	if len(result) != len(expected) {
		t.Errorf("merkle tree sizes failed: %v %v", len(result), len(expected))
	}
	for i, e := range expected {
		if bytes.Compare(e, result[i]) != 0 {
			t.Errorf("merkle tree failed at %v:", i)
			t.Errorf("%v", e)
			t.Errorf("%v", result[i])
		}
	}

	// Test the Verify function against a good and bad value.
	if result.Verify([]byte("bad")) {
		t.Errorf("Verify succeeded when it should have failed.")
	}
	h.Reset()
	h.Write([]byte("cat"))
	if !result.Verify(h.Sum(nil)) {
		t.Errorf("Verify failed when it should have succeeded.")
	}

	// Test an empty tree.
	expected = MerkleTree{}
	result = NewMerkleTree([][]byte{}, h)
	if len(result) != len(expected) {
		t.Errorf("merkle tree sizes failed: %v %v", len(result), len(expected))
	}
	for i, e := range expected {
		if bytes.Compare(e, result[i]) != 0 {
			t.Errorf("merkle tree failed at %v:", i)
			t.Errorf("%v", e)
			t.Errorf("%v", result[i])
		}
	}
	if result.Verify([]byte("bad")) {
		t.Errorf("Verify succeeded when it should have failed.")
	}
}

func TestMerkleTreeRoot(t *testing.T) {
	mt := MerkleTree{}
	if mt.Root() != nil {
		t.Errorf("empty MerkleTree returned a non-nil root.")
	}
}

func TestMerkleTreeProofAndVerify(t *testing.T) {
	data := [][]byte{
		[]byte("cat"),
		[]byte("dog"),
		[]byte("mouse"),
		[]byte("parrot"),
		[]byte("hamster"),
		[]byte("goat"),
	}

	// Build the tree and expected values for testing.
	h := sha256.New()
	result := NewMerkleTree(data, h)
	root := result.Root()
	expected := make([][]byte, 15)
	for i, d := range data {
		h.Reset()
		h.Write(d)
		expected[i+7] = h.Sum(nil)
	}
	sets := [][]int{
		[]int{3, 5},
		[]int{1, 2},
		[]int{0, 0},
	}
	for _, s := range sets {
		for x := s[0]; x <= s[1]; x++ {
			h.Reset()
			h.Write(expected[x*2+1])
			h.Write(expected[x*2+2])
			expected[x] = h.Sum(nil)
		}
	}

	// Test for a non-leaf node.
	proof := result.Proof(expected[0])
	if proof != nil {
		t.Errorf("non-nil returned for non-leaf node.")
	}

	// Test all the leaf nodes (should all succeed).
	for x := 7; x <= 12; x++ {
		proof = result.Proof(expected[x])
		if !VerifyMerkleProof(proof, expected[x], root, h) {
			t.Errorf("proof of '%v' failed verification.", string(data[x-7]))
		}
	}

	// Muck with the proof to produce some falses.
	// nill proof.
	if VerifyMerkleProof(nil, expected[12], root, h) {
		t.Errorf("nil proof returned true.")
	}
	// differing leaf.
	if VerifyMerkleProof(proof, expected[11], root, h) {
		t.Errorf("proof with differing leaf returned true.")
	}
	// differing root.
	if VerifyMerkleProof(proof, expected[12], nil, h) {
		t.Errorf("proof with differing root returned true.")
	}
	// altered proof (one that shouldn't match).
	proof.Parent.Parent.Sum = nil
	if VerifyMerkleProof(proof, expected[12], nil, h) {
		t.Errorf("proof with differing middle value returned true.")
	}
}
