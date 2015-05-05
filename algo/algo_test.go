// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package algo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"testing"
)

func ExampleLevenshtein() {
	h := "Happy Christmas"
	m := "Merry Christmas"
	l := Levenshtein(h, m)
	fmt.Println(l)
	// Output:
	// 4
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		a []int
		e int
	}{
		{e: 0, a: []int{}},
		{e: 1, a: []int{1}},
		{e: 30, a: []int{30, 100}},
		{e: 30, a: []int{100, 30}},
		{e: 10, a: []int{90, 40, 60, 12, 11, 10, 11, 14, 19, 100, 11, 11, 20}},
	}
	for k, test := range tests {
		r := MinInt(test.a...)
		if r != test.e {
			t.Errorf("Test %v: MinInt(%v) = %v, expected %v", k, test.a, r, test.e)
		}
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		a []int
		e int
	}{
		{e: 0, a: []int{}},
		{e: 1, a: []int{1}},
		{e: 100, a: []int{30, 100}},
		{e: 100, a: []int{100, 30}},
		{e: 100, a: []int{90, 40, 60, 12, 11, 10, 11, 14, 19, 100, 11, 11, 20}},
	}
	for k, test := range tests {
		r := MaxInt(test.a...)
		if r != test.e {
			t.Errorf("Test %v: MaxInt(%v) = %v, expected %v", k, test.a, r, test.e)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		s string
		t string
		e int
	}{
		{
			s: "",
			t: "test",
			e: 4,
		},
		{
			s: "test",
			t: "",
			e: 4,
		},
		{
			s: "Claredi",
			t: "Claredi",
			e: 0,
		},
		{
			s: "Claredi",
			t: "Clarity",
			e: 3,
		},
		{
			s: "Claredi",
			t: "Clardi",
			e: 1,
		},
	}
	for k, test := range tests {
		r := Levenshtein(test.s, test.t)
		t.Logf("Levenshtein(%v, %v) = %v", test.s, test.t, r)
		if r != test.e {
			t.Errorf("Test %v: Levenshtein(%v, %v) = %v, expected %v",
				k, test.s, test.t, r, test.e)
		}
	}
}

func BenchmarkLevenshtein(b *testing.B) {
	b.StopTimer()
	// Build a large sample set to test with.
	var dict []string
	var test = "Claredi"
	f, err := os.Open("/usr/share/dict/american-english")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	x := 0
	for s.Scan() {
		dict = append(dict, s.Text())
		x++
		if x > 10000 {
			break
		}
	}
	if s.Err() != nil {
		panic(err)
	}

	b.StartTimer()
	for _, t := range dict {
		Levenshtein(test, t)
	}
}

func TestNPIChecksum(t *testing.T) {
	// This also checks most of Luhn.
	tests := []struct {
		s        string
		expected string
		err      error
	}{
		{s: "123456789", expected: "3", err: nil},
		{s: "992739871", expected: "6", err: nil},
		{s: "000000000", expected: "6", err: nil},
		{s: "300000000", expected: "0", err: nil},
		{s: "000000060", expected: "0", err: nil},
		{s: "A0000", expected: "", err: errors.New(`strconv.ParseInt: parsing "A": invalid syntax`)},
	}

	for k, test := range tests {
		result, err := NPIChecksum(test.s)
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
			t.Errorf("Test %v: expected error '%v' but got '%v'", k, test.err, err)
		}
		if result != test.expected {
			t.Errorf("Test %v: expected result '%v' but got '%v'", k, test.expected, result)
		}
	}
}

func TestNPIChecksumAppend(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      error
	}{
		{s: "123456789", expected: "1234567893", err: nil},
		{s: "992739871", expected: "9927398716", err: nil},
		{s: "000000000", expected: "0000000006", err: nil},
		{s: "300000000", expected: "3000000000", err: nil},
		{s: "000000060", expected: "0000000600", err: nil},
		{s: "A0000", expected: "", err: errors.New(`strconv.ParseInt: parsing "A": invalid syntax`)},
	}

	for k, test := range tests {
		result, err := NPIChecksumAppend(test.s)
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
			t.Errorf("Test %v: expected error '%v' but got '%v'", k, test.err, err)
		}
		if result != test.expected {
			t.Errorf("Test %v: expected result '%v' but got '%v'", k, test.expected, result)
		}
	}
}

func TestNPIChecksumCheck(t *testing.T) {
	// This also checks most of LuhnCheck.
	tests := []struct {
		s        string
		expected bool
	}{
		{s: "1234567893", expected: true},
		{s: "9927398716", expected: true},
		{s: "0000000006", expected: true},
		{s: "3000000000", expected: true},
		{s: "0000000600", expected: true},
		{s: "A000000000", expected: false},
		{s: "1234567890", expected: false},
		{s: "1234567891", expected: false},
		{s: "1234567892", expected: false},
		{s: "1234567894", expected: false},
		{s: "1234567895", expected: false},
		{s: "1234567896", expected: false},
		{s: "1234567897", expected: false},
		{s: "1234567898", expected: false},
		{s: "1234567899", expected: false},
	}

	for k, test := range tests {
		result := NPIChecksumCheck(test.s)
		if result != test.expected {
			t.Errorf("Test %v: NPIChecksumCheck(%v) != %v", k, test.s, test.expected)
		}
	}
}

func TestLuhnErrors(t *testing.T) {
	if LuhnCheck("1") != false {
		t.Errorf(`LuhnCheck("1") != false`)
	}
	if s, err := Luhn("1"); s != "" || err != ErrInvalidParams {
		t.Errorf(`Luhn("1") != "", ErrInvalidParams`)
	}
}

func TestLuhnAppend(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      error
	}{
		{s: "123456789", expected: "1234567897", err: nil},
		{s: "992739871", expected: "9927398710", err: nil},
		{s: "000000000", expected: "0000000000", err: nil},
		{s: "300000000", expected: "3000000004", err: nil},
		{s: "000000060", expected: "0000000604", err: nil},
		{s: "A0000", expected: "", err: errors.New(`strconv.ParseInt: parsing "A": invalid syntax`)},
	}

	for k, test := range tests {
		result, err := LuhnAppend(test.s)
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
			t.Errorf("Test %v: expected error '%v' but got '%v'", k, test.err, err)
		}
		if result != test.expected {
			t.Errorf("Test %v: expected result '%v' but got '%v'", k, test.expected, result)
		}
	}

}
