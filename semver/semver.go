// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package semver implements structs and functions that adhere to
// Semantic Versioning (http://semver.org/).
package semver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrParse is returned when New is unable to parse the given string
// into a semantic version.
var ErrParse = errors.New("unable to parse given string into a semantic version")

// SemanticVersion is a handy struct to handle with versioning. You can create
// one from a string and then find compatible versions and compare it
// to other versions. For more information see: http://semver.org/.
type SemanticVersion struct {
	Major int
	Minor int
	Patch int
}

// New creates a new semantic version from the given string.
func New(v string) (SemanticVersion, error) {
	nv := SemanticVersion{}
	// Verify it starts with a v.
	if v[:1] != "v" {
		return nv, ErrParse
	}
	// Split it out by it constituent parts, parse it, and then set the
	// right value.
	for i, part := range strings.Split(v[1:], ".") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nv, ErrParse
		}
		switch i {
		case 0:
			nv.Major = n
		case 1:
			nv.Minor = n
		case 2:
			nv.Patch = n
		default:
			return nv, nil
		}
	}
	return nv, nil
}

// GreaterEqual returns true if v is greater than or equal to o.
func (v SemanticVersion) GreaterEqual(o SemanticVersion) bool {
	if o.Major > v.Major {
		return false
	} else if o.Minor > v.Minor {
		return false
	} else if o.Patch > v.Patch {
		return false
	}
	return true
}

// Compatible returns true if v is compatible with o.
func (v SemanticVersion) Compatible(o SemanticVersion) bool {
	return v.Major == o.Major && v.GreaterEqual(o)
}

// String returns the version as a string.
func (v SemanticVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}
