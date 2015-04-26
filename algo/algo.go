// Package algo provides common computer science algorithms and data
// structures implemented in pure Go.
package algo

import (
	"errors"
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
