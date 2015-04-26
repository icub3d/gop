package algo

import (
	"bufio"
	"os"
	"testing"
)

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
