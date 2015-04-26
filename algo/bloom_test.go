package algo

import "testing"

func TestNewBloomFilter(t *testing.T) {
	bf := NewBloomFilter(10, 2)
	if bf.m != 10 {
		t.Errorf("NewBloomFilter(10, 2) failed at bf.m: %v", bf.m)
	}
	if bf.k != 2 {
		t.Errorf("NewBloomFilter(10, 2) failed at bf.k: %v", bf.k)
	}
	if len(bf.bs) != 1 {
		t.Errorf("NewBloomFilter(10, 2) failed at len(bf.bs): %v", len(bf.bs))
	}
	bf = NewBloomFilterEstimate(1000, .01)
	if bf.m != 9585 {
		t.Errorf("NewBloomFilterEstimate(1000, .01) failed at bf.m: %v", bf.m)
	}
	if bf.k != 7 {
		t.Errorf("NewBloomFilterEstimate(1000, .01) failed at bf.k: %v", bf.k)
	}
	if len(bf.bs) != 150 {
		t.Errorf("NewBloomFilterEstimate(1000, .01) failed at len(bf.bs): %v", len(bf.bs))
	}
}

func TestBloomFilterFalsePositives(t *testing.T) {
	bf := BloomFilter{
		k: 7,
		n: 8500,
		m: 9585,
	}
	p := bf.FalsePositives()
	if p != 0.9827772314927979 {
		t.Errorf("bf.FalsePositive() with (k: 7, n: 8500, m: 9585) failed with: %v, "+
			"expected 0.9827772314927979", p)
	}
}

func TestBloomFilterAddExists(t *testing.T) {
	bf := NewBloomFilter(100, 5)
	for _, s := range []string{"Dog", "Cat", "Mouse", "Elephant", "Lion", "Giraffe"} {
		bf.Add([]byte(s))
		if !bf.Exists([]byte(s)) {
			t.Errorf("Add(%v) did not produce a true value for Exists(%v).", s, s)
		}
	}
	if bf.Exists([]byte("Garbage")) {
		t.Errorf("Exists(Garbage) was found, but shouldn't be.")
	}
}
