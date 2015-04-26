// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package semver

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleSemanticVersion() {
	v1, _ := New("v1.0.0")
	v2, _ := New("v1.0.1")
	v3, _ := New("v1.1.3")
	v4, _ := New("v2.0.0")

	if v2.Compatible(v1) {
		fmt.Printf("%s is compatible with %s.\n", v2, v1)
	}
	if v3.Compatible(v1) {
		fmt.Printf("%s is compatible with %s.\n", v3, v1)
	}
	if !v4.Compatible(v1) {
		fmt.Printf("%s is not compatible with %s.\n", v4, v1)
	}

	// Output:
	// v1.0.1 is compatible with v1.0.0.
	// v1.1.3 is compatible with v1.0.0.
	// v2.0.0 is not compatible with v1.0.0.
}

func TestNewSemanticVersion(t *testing.T) {
	tests := []struct {
		v        string
		expected SemanticVersion
		err      error
	}{
		// Test a valid version.
		{
			v:        "v1.2.3",
			expected: SemanticVersion{1, 2, 3},
		},
		// Test a valid version with just major.
		{
			v:        "v3",
			expected: SemanticVersion{3, 0, 0},
		},
		// Test a valid version with major.minor.
		{
			v:        "v4.5",
			expected: SemanticVersion{4, 5, 0},
		},
		// Test a string that doesn't start with a v.
		{
			v:   "1.2.3",
			err: ErrParse,
		},
		// Test a bad major version.
		{
			v:   "va.2.3",
			err: ErrParse,
		},
		// Test a bad minor version.
		{
			v:   "v1.a.3",
			err: ErrParse,
		},
		// Test a bad patch version.
		{
			v:   "v1.2.a",
			err: ErrParse,
		},
	}

	for i, test := range tests {
		v, err := New(test.v)
		if err != test.err {
			t.Errorf("Test %v: New(%v) returned error %v, wanted %v", i,
				test.v, err, test.err)
			continue
		}
		if err == nil && !reflect.DeepEqual(v, test.expected) {
			t.Errorf("Test %v: New(%v) = %v, wanted %v", i,
				test.v, v, test.expected)
		}
	}
}

func TestSemanticVersionGreaterEqual(t *testing.T) {
	tests := []struct {
		v, o     SemanticVersion
		expected bool
	}{
		// Test a bunch of true values.
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 2, 3},
			expected: true,
		},
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 2, 2},
			expected: true,
		},
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 1, 3},
			expected: true,
		},
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{0, 2, 3},
			expected: true,
		},
		// Test a bunch of false values.
		{
			v: SemanticVersion{1, 2, 3},
			o: SemanticVersion{2, 2, 3},
		},
		{
			v: SemanticVersion{1, 2, 3},
			o: SemanticVersion{1, 3, 3},
		},
		{
			v: SemanticVersion{1, 2, 3},
			o: SemanticVersion{1, 2, 4},
		},
	}

	for i, test := range tests {
		result := test.v.GreaterEqual(test.o)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test %v: %v.GreaterEqual(%v) = %v, wanted %v", i,
				test.v, test.o, result, test.expected)
		}
	}
}

func TestSemanticVersionCompatible(t *testing.T) {
	tests := []struct {
		v, o     SemanticVersion
		expected bool
	}{
		// Test a bunch of true values.
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 2, 3},
			expected: true,
		},
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 2, 2},
			expected: true,
		},
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 1, 3},
			expected: true,
		},
		{
			v:        SemanticVersion{1, 2, 3},
			o:        SemanticVersion{1, 0, 0},
			expected: true,
		},
		// Test a bunch of false values.
		{
			v: SemanticVersion{1, 2, 3},
			o: SemanticVersion{2, 2, 3},
		},
		{
			v: SemanticVersion{1, 2, 3},
			o: SemanticVersion{1, 3, 3},
		},
		{
			v: SemanticVersion{1, 2, 3},
			o: SemanticVersion{1, 2, 4},
		},
	}

	for i, test := range tests {
		result := test.v.Compatible(test.o)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test %v: %v.Compatible(%v) = %v, wanted %v", i,
				test.v, test.o, result, test.expected)
		}
	}
}

func TestSemanticVersionString(t *testing.T) {
	tests := []struct {
		v        SemanticVersion
		expected string
	}{
		{
			v:        SemanticVersion{1, 2, 3},
			expected: "v1.2.3",
		},
		{
			v:        SemanticVersion{1, 2, 0},
			expected: "v1.2.0",
		},
		{
			v:        SemanticVersion{1, 0, 0},
			expected: "v1.0.0",
		},
	}

	for i, test := range tests {
		result := test.v.String()
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test %v: %v.String() = %v, wanted %v", i,
				test.v, result, test.expected)
		}
	}
}
