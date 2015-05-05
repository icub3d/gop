// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package algo provides common computer science algorithms and data
// structures implemented in pure Go.
package algo

import (
	"errors"
	"strconv"
	"strings"
)

var (
	// ErrNotFound means that the operation was unable to find what you
	// were looking for.
	ErrNotFound = errors.New("not found")

	// ErrOutOfRange means that the operation was unable to to do what
	// you asked because the parameters for the request were outside of
	// the limits. You should check that you are working within the
	// right range specified by the method or function.
	ErrOutOfRange = errors.New("out of range")

	// ErrWrongType means that the wrong type was given. This usually
	// occurs when an interface{} is being converted and the type isn't
	// what was needed.
	ErrWrongType = errors.New("wrong type")

	// ErrNotImplemented means that the operation is not implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidParams means that you gave the function something
	// unexpected.
	ErrInvalidParams = errors.New("invalid params")
)

// MinInt returns the smallest integer among all of the given
// integers. 0 is returned when no integers are given.
func MinInt(is ...int) int {
	if len(is) < 1 {
		return 0
	}
	min := is[0]
	for _, i := range is {
		if i < min {
			min = i
		}
	}
	return min
}

// MaxInt returns the largest integer among all of the given
// integers. 0 is returned when no integers are given.
func MaxInt(is ...int) int {
	if len(is) < 1 {
		return 0
	}
	min := is[0]
	for _, i := range is {
		if i > min {
			min = i
		}
	}
	return min
}

// Levenshtein calculates the levenshtein distance between the two
// given strings. For more information on what a levenshtein distance
// is, see: http://en.wikipedia.org/wiki/Levenshtein_distance.
func Levenshtein(s, t string) int {
	// Sanity checks.
	if s == t {
		return 0
	}
	if len(s) == 0 {
		return len(t)
	}
	if len(t) == 0 {
		return len(s)
	}

	// Create two rows for tracking.
	v := make([][]int, 2)
	v[0] = make([]int, len(t)+1)
	v[1] = make([]int, len(t)+1)
	// Initialize.
	for i := 0; i < len(v[0]); i++ {
		v[0][i] = i
	}

	// Iterate and return.
	for i := 0; i < len(s); i++ {
		v[1][0] = i + 1
		for j := 0; j < len(t); j++ {
			c := 1
			if s[i] == t[j] {
				c = 0
			}
			v[1][j+1] = MinInt(v[1][j]+1, v[0][j+1]+1, v[0][j]+c)
		}
		for j := 0; j < len(v[0]); j++ {
			v[0][j] = v[1][j]
		}
	}
	return v[1][len(t)]
}

// LuhnCheck verifies the checksum (the last digit) of the given
// number using the Luhn algorithm:
// http://en.wikipedia.org/wiki/Luhn_algorithm.
func LuhnCheck(s string) bool {
	if len(s) < 2 {
		return false
	}
	c := s[len(s)-1:]
	l, err := Luhn(s[:len(s)-1])
	if err != nil {
		return false
	}
	return l == c
}

// Luhn calculates the Luhn checksum of the given number using the
// Luhn algorithm: http://en.wikipedia.org/wiki/Luhn_algorithm. It
// should not have a checksum at the end of it.
func Luhn(s string) (string, error) {
	if len(s) < 2 {
		return "", ErrInvalidParams
	}
	double := true
	sum := 0
	for x := len(s) - 1; x >= 0; x-- {
		i, err := strconv.Atoi(s[x : x+1])
		if err != nil {
			return "", err
		}
		if double {
			i *= 2
			if i >= 10 {
				i = 1 + i%10
			}
			sum += i
		} else {
			sum += i
		}
		double = !double
	}
	sum *= 9
	return strconv.Itoa(sum % 10), nil
}

// LuhnAppend appends the Luhn checksum to the given number.
func LuhnAppend(n string) (string, error) {
	l, err := Luhn(n)
	if err != nil {
		return "", err
	}
	return n + l, nil
}

// NPIChecksum returns the Luhn checksum for the given number. It
// differs from a normal Luhn in that if the number doesn't begin with
// 80840, 80840 will be prepended to the number for determining the
// checksum. It should not have a checksum at the end of it.
func NPIChecksum(s string) (string, error) {
	if !strings.HasPrefix(s, "80840") {
		s = "80840" + s
	}
	return Luhn(s)
}

// NPIChecksumCheck calculates and verifies the checksum using the
// Luhn algorithm. If the number doesn't begin with 80840, 80840 will
// be prepended to it before checking.
func NPIChecksumCheck(s string) bool {
	if !strings.HasPrefix(s, "80840") {
		s = "80840" + s
	}
	return LuhnCheck(s)
}

// NPIChecksumAppend calculates the checksum for the given partial NPI
// and appends the checksum to it. If the number doesn't begin with
// 80840, 80840 will be prepended to it before determing the checksum,
// but won't be included in the result.
func NPIChecksumAppend(s string) (string, error) {
	cs, err := NPIChecksum(s)
	if err != nil {
		return "", err
	}
	return s + cs, nil
}
