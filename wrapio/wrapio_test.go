package wrapio

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
)

func ExampleNewFuncReader() {
	// This is the string we'll read from.
	r := strings.NewReader("This is the sample data that we are going to test with.")
	// Create our wrappers and use them.
	fr := NewFuncReader(func(p []byte) {
		// Replace all spaces with a period. I'm assuming ascii here for
		// the simplicity of the example.
		for x, b := range p {
			if b == 0x20 {
				p[x] = 0x2E
			}
		}
	}, r)
	b := &bytes.Buffer{}
	io.Copy(b, fr)
	fmt.Println(b.String())
	// Output:
	// This.is.the.sample.data.that.we.are.going.to.test.with.
}

func ExampleNewHashReader() {
	// We'll read from this using io.Copy.
	r := strings.NewReader("This is the sample data that we are going to test with.")
	// Generate different hashes on each side.
	m := md5.New()
	s := sha256.New()
	// Create our wrappers and use them.
	hr := NewHashReader(m, r)
	hw := NewHashWriter(s, ioutil.Discard)
	io.Copy(hw, hr)
	// Use the Sum()s which in this case we'll just print it out.
	fmt.Println(hex.EncodeToString(m.Sum(nil)))
	fmt.Println(hex.EncodeToString(s.Sum(nil)))
	// Output:
	// 9bd2f8a51a7745e0e0af586736f93944
	// 52b846d6fedeb0a90acec7ce09f7d590ec4db0e5bd1884bc74c1d81e3c00b471
}

func TestHashReader(t *testing.T) {
	tests := []struct {
		data     string
		expected string
		hash     hash.Hash
	}{
		{
			data:     "this is a test.",
			expected: "09cba091df696af91549de27b8e7d0f6",
			hash:     md5.New(),
		},
		{
			data:     "I eat pizza for breakfast and there is nothing you can do to stop me.",
			expected: "9df01f46c48c1be6507a14b73da3ad7007f7815b0ddae3465707931403f32e46",
			hash:     sha256.New(),
		},
	}
	for k, test := range tests {
		sr := strings.NewReader(test.data)
		hr := NewHashReader(test.hash, sr)
		ioutil.ReadAll(hr)
		s := hex.EncodeToString(test.hash.Sum(nil))
		if s != test.expected {
			t.Errorf("Test %v: unexpected Sum(), got vs expected:\n%v\n%v",
				k, s, test.expected)
		}
	}
	// Test the special error cases.
	if NewHashReader(tests[0].hash, nil) != nil {
		t.Errorf("nil io.Reader didn't return nil.")
	}
	if NewHashReader(nil, strings.NewReader("")) != nil {
		t.Errorf("nil hash didn't return nil.")
	}
}

func TestHashWriter(t *testing.T) {
	tests := []struct {
		data     string
		expected string
		hash     hash.Hash
	}{
		{
			data:     "this is a test.",
			expected: "09cba091df696af91549de27b8e7d0f6",
			hash:     md5.New(),
		},
		{
			data:     "I eat pizza for breakfast and there is nothing you can do to stop me.",
			expected: "9df01f46c48c1be6507a14b73da3ad7007f7815b0ddae3465707931403f32e46",
			hash:     sha256.New(),
		},
	}
	for k, test := range tests {
		sr := strings.NewReader(test.data)
		hw := NewHashWriter(test.hash, ioutil.Discard)
		io.Copy(hw, sr)
		s := hex.EncodeToString(test.hash.Sum(nil))
		if s != test.expected {
			t.Errorf("Test %v: unexpected Sum(), got vs expected:\n%v\n%v",
				k, s, test.expected)
		}
	}
	// Test the special error cases.
	if NewHashWriter(tests[0].hash, nil) != nil {
		t.Errorf("nil io.Writer didn't return nil.")
	}
	if NewHashWriter(nil, ioutil.Discard) != nil {
		t.Errorf("nil hash didn't return nil.")
	}
}

func ExampleNewStatsReader() {
	// We'll read from this using io.Copy.
	sr := strings.NewReader("This is the sample data that we are going to test with.")
	// Create our wrappers and use them.
	s, r := NewStatsReader(iotest.OneByteReader(sr))
	io.Copy(ioutil.Discard, r)
	// Print out the statistics.
	s.Lock()
	defer s.Unlock()
	fmt.Println(s)
	// Output:
	// [Total: 55, Average: 1.000000, Calls: 55]
}

func TestStatsString(t *testing.T) {
	s := Stats{Total: 10, Average: 2.193, Calls: 5}
	if s.String() != "[Total: 10, Average: 2.193000, Calls: 5]" {
		t.Errorf("Stats.String() produced the wrong ouptut: %v", s)
	}
}

func TestStatsReader(t *testing.T) {
	tests := []struct {
		data     string
		expected *Stats
	}{
		{
			data:     "this is a test.",
			expected: &Stats{Total: 15, Average: 15, Calls: 1},
		},
	}
	for k, test := range tests {
		sr := strings.NewReader(test.data)
		s, hr := NewStatsReader(sr)
		ioutil.ReadAll(hr)
		if s.Total != test.expected.Total || s.Calls != test.expected.Calls ||
			s.Average != test.expected.Average {
			t.Errorf("Test %v: unexpected stats, got vs expected:\n%v\n%v",
				k, s, test.expected)
		}
	}
}

func TestStatsWriter(t *testing.T) {
	tests := []struct {
		data     string
		expected *Stats
	}{
		{
			data:     "this is a test.",
			expected: &Stats{Total: 15, Average: 15, Calls: 1},
		},
	}
	for k, test := range tests {
		sr := strings.NewReader(test.data)
		s, hw := NewStatsWriter(ioutil.Discard)
		io.Copy(hw, sr)
		if s.Total != test.expected.Total || s.Calls != test.expected.Calls ||
			s.Average != test.expected.Average {
			t.Errorf("Test %v: unexpected stats, got vs expected:\n%v\n%v",
				k, s, test.expected)
		}
	}
}

func ExampleNewBlockReader() {
	// This is the buffer that we'll read from.
	buf := strings.NewReader("0123456789")
	br := NewBlockReader(3, buf)
	p := make([]byte, 5)
	// Read until we get an error.
	var err error
	var n int
	for err == nil {
		n, err = br.Read(p)
		fmt.Println(n, err, string(p[:n]))
	}
	// Output:
	// 3 <nil> 012
	// 3 <nil> 345
	// 3 <nil> 678
	// 1 <nil> 9
	// 0 EOF
}

func TestBlockWriter(t *testing.T) {
	// We'll test various valid sizes and lengths against strings as a
	// sort of functional test.
	tests := []struct {
		size   int
		writes []string
		reads  []string
		ns     []int
		errs   []error
		last   string
	}{
		{
			size: 2,
			writes: []string{
				"01234", "5678", "90",
			},
			reads: []string{
				"0123", "4567", "89",
			},
			ns: []int{
				5, 4, 2,
			},
			errs: []error{
				nil, nil, nil,
			},
			last: "0",
		},
		{
			size: 2,
			writes: []string{
				"01234", "5678", "9",
			},
			reads: []string{
				"0123", "4567", "89",
			},
			ns: []int{
				5, 4, 1,
			},
			errs: []error{
				nil, nil, nil,
			},
			last: "",
		},
	}
	for k, test := range tests {
		bw := new(bytes.Buffer)
		w := NewBlockWriter(test.size, bw)
		for x, write := range test.writes {
			n, err := w.Write([]byte(write))
			if n != test.ns[x] {
				t.Errorf("Test %v(%v): n == %v, but wanted %v",
					k, x, n, test.ns[x])
			}
			if err != test.errs[x] {
				t.Errorf("Test %v(%v): err == %v, but wanted %v",
					k, x, err, test.errs[x])
			}
			s := bw.String()
			if s != test.reads[x] {
				t.Errorf("Test %v(%v): s == '%v', but wanted '%v'",
					k, x, s, test.reads[x])
			}
			bw.Reset()
		}
		w.Close()
		s := bw.String()
		if s != test.last {
			t.Errorf("Test %v: last == '%v', but wanted '%v'",
				k, s, test.last)
		}
	}
	// Test a nil writer.
	if NewBlockWriter(1, nil) != nil {
		t.Errorf("nil io.Writer didn't return nil.")
	}
	if NewBlockWriter(0, &bytes.Buffer{}) != nil {
		t.Errorf("zero size didn't return nil.")
	}
	// Test with the error writer.
	e := ew{err: fmt.Errorf("i did it")}
	w := NewBlockWriter(1, e)
	for x := 0; x < 2; x++ {
		n, err := w.Write([]byte("test"))
		if n != 0 || err == nil {
			t.Errorf("Test %v: bad error writer results: %v %v",
				x, n, err)
		}
	}
	err := w.Close()
	if err == nil {
		t.Errorf("bad error close results: %v", err)
	}
}

func TestBlockReaderFunctional(t *testing.T) {
	if NewBlockReader(0, er{}) != nil {
		t.Errorf("zero reader size didn't return nil")
	}
	if NewBlockReader(1, nil) != nil {
		t.Errorf("nil reader didn't return nil")
	}
	r := strings.NewReader("0123456789")
	br := NewBlockReader(3, r)
	s, sr := NewStatsReader(br)
	buf := &bytes.Buffer{}
	io.Copy(buf, sr)
	if buf.String() != "0123456789" {
		t.Errorf("expected output '%v' != results '%v'",
			"0123456789", buf.String())
	}
	if s.Calls != 2 {
		t.Errorf("expected calls %v != results %v",
			2, s.Calls)
	}
}

// This does some unit testing. It puts the block in an artificial
// state and checks the expected outcome.
func TestBlockReaderUnitTest(t *testing.T) {
	tests := []struct {
		p        []byte
		expected []byte
		block    block
		buf      []byte
		n        int
		err      error
	}{
		// Test being at the end with no data left.
		{
			p:        make([]byte, 10),
			expected: []byte{},
			block: block{
				err: io.EOF,
			},
			n:   0,
			err: io.EOF,
		},
		// Test where len(p) is less than block size.
		{
			p:        make([]byte, 10),
			expected: []byte{},
			block: block{
				size: 20,
			},
			n:   0,
			err: nil,
		},
		// Test where we have an error, but still some in the buffer. We
		// want to pull full blocks in this one.
		{
			p:        make([]byte, 10),
			expected: []byte{48, 49, 50, 51},
			block: block{
				buf:  []byte("01234"),
				size: 2,
				err:  io.EOF,
			},
			buf: []byte("4"),
			n:   4,
			err: nil,
		},
		// Same as above, but p is smaller.
		{
			p:        make([]byte, 3),
			expected: []byte{48, 49},
			block: block{
				buf:  []byte("01234"),
				size: 2,
				err:  io.EOF,
			},
			buf: []byte("234"),
			n:   2,
			err: nil,
		},
		// Same as above, but we only have a partial block.
		{
			p:        make([]byte, 4),
			expected: []byte{48, 49, 50},
			block: block{
				buf:  []byte("012"),
				size: 4,
				err:  io.EOF,
			},
			buf: []byte{},
			n:   3,
			err: nil,
		},
		// We aren't in an error condition, so we read and append
		{
			p:        make([]byte, 5),
			expected: []byte{48, 49, 50, 51},
			block: block{
				r: er{
					data: []byte("34567"),
					n:    5,
					err:  nil,
				},
				buf:  []byte("012"),
				size: 4,
				err:  nil,
			},
			buf: []byte("4567"),
			n:   4,
			err: nil,
		},
		// We don't get a full block after a read
		{
			p:        make([]byte, 5),
			expected: []byte{},
			block: block{
				r: er{
					data: []byte(""),
					n:    0,
					err:  nil,
				},
				buf:  []byte("012"),
				size: 4,
				err:  nil,
			},
			buf: []byte("012"),
			n:   0,
			err: nil,
		},
	}
	for k, test := range tests {
		n, err := test.block.Read(test.p)
		if !reflect.DeepEqual(test.p[:n], test.expected) {
			t.Errorf("Test %v: p (%v) != expected (%v)",
				k, test.p, test.expected)
		}
		if n != test.n {
			t.Errorf("Test %v: n (%v) != expected (%v)",
				k, n, test.n)
		}
		if err != test.err {
			t.Errorf("Test %v: err (%v) != expected (%v)",
				k, err, test.err)
		}
		if test.buf != nil {
			if !reflect.DeepEqual(test.buf, test.block.buf) {
				t.Errorf("Test %v: buf (%v) != expected (%v)",
					k, test.block.buf, test.buf)
			}
		}
	}
}

func ExampleNewLastFuncReader() {
	// This is the buffer that we'll read from.
	buf := strings.NewReader("0123456789")
	br := NewLastFuncReader(func(p []byte) []byte {
		// We are going to pad with zeros.
		for len(p) < 3 {
			p = append(p, 48)
		}
		return p
	}, buf)
	p := make([]byte, 3)
	// Read until we get an error.
	var err error
	var n int
	for err == nil {
		n, err = br.Read(p)
		fmt.Println(n, err, string(p[:n]))
	}
	// Output:
	// 3 <nil> 012
	// 3 <nil> 345
	// 3 <nil> 678
	// 3 EOF 900
}

func TestLastFuncReader(t *testing.T) {
	tests := []struct {
		ps       [][]byte
		expected [][]byte
		errs     []error
		data     io.Reader
		f        func([]byte) []byte
	}{
		// General case.
		{
			ps: [][]byte{
				make([]byte, 3),
				make([]byte, 3),
				make([]byte, 3),
				make([]byte, 3),
			},
			errs: []error{
				nil,
				nil,
				nil,
				io.EOF,
			},
			expected: [][]byte{
				[]byte("012"),
				[]byte("345"),
				[]byte("678"),
				[]byte("900"),
			},
			data: strings.NewReader("0123456789"),
			f: func(p []byte) []byte {
				for len(p) < 3 {
					p = append(p, 48)
				}
				return p
			},
		},
		// Our func returns a value less than p.
		{
			ps: [][]byte{
				make([]byte, 3),
				make([]byte, 3),
				make([]byte, 3),
			},
			errs: []error{
				nil,
				nil,
				io.EOF,
			},
			expected: [][]byte{
				[]byte("012"),
				[]byte("345"),
				[]byte("6"),
			},
			data: strings.NewReader("012345678"),
			f: func(p []byte) []byte {
				return p[:1]
			},
		},
		// p is too big
		{
			ps: [][]byte{
				make([]byte, 3),
				make([]byte, 100),
				make([]byte, 3),
			},
			errs: []error{
				nil,
				nil,
				io.EOF,
			},
			expected: [][]byte{
				[]byte("012"),
				[]byte("345"),
				[]byte("6"),
			},
			data: strings.NewReader("012345678"),
			f: func(p []byte) []byte {
				return p[:1]
			},
		},
		// Test the case where the first read returns an error.
		{
			ps: [][]byte{
				make([]byte, 3),
			},
			errs: []error{
				io.EOF,
			},
			expected: [][]byte{
				[]byte("100"),
			},
			data: &er{
				data: []byte("1"),
				n:    1,
				err:  io.EOF,
			},
			f: func(p []byte) []byte {
				for len(p) < 3 {
					p = append(p, 48)
				}
				return p
			},
		},
	}
	for k, test := range tests {
		r := NewLastFuncReader(test.f, test.data)
		for x, expected := range test.expected {
			n, err := r.Read(test.ps[x])
			if n != len(test.expected[x]) {
				t.Errorf("Test %v(%v): n (%v) != expected (%v)",
					k, x, n, len(test.expected[x]))
			}
			if err != test.errs[x] {
				t.Errorf("Test %v(%v): err (%v) != expected (%v)",
					k, x, err, test.errs[x])
			}
			if !reflect.DeepEqual(test.ps[x][:n], expected) {
				t.Errorf("Test %v(%v): written (%v) != expected (%v)",
					k, x, test.ps[x][:n], expected)
			}
		}
		// Do one last call and expect 0, EOF.
		if n, err := r.Read(test.ps[0]); n != 0 && err != io.EOF {
			t.Errorf("Test %v: final read didn't return 0, EOF: %v %v",
				k, n, err)
		}
	}
	// Test the special error cases.
	if NewLastFuncReader(tests[0].f, nil) != nil {
		t.Errorf("nil io.Writer didn't return nil.")
	}
	if NewLastFuncReader(nil, &bytes.Buffer{}) != nil {
		t.Errorf("nil func did't return nil.")
	}
}

func TestLastFuncWriter(t *testing.T) {
	tests := []struct {
		ps       []string
		expected []string
		last     string
		f        func([]byte) []byte
	}{
		{
			ps:       []string{"1", "22", "3", "4444", "5"},
			expected: []string{"", "1", "22", "3", "4444"},
			last:     "5: THE END",
			f: func(p []byte) []byte {
				return append(p, []byte(": THE END")...)
			},
		},
	}
	for k, test := range tests {
		buf := &bytes.Buffer{}
		w := NewLastFuncWriter(test.f, buf)
		for x, p := range test.ps {
			n, err := w.Write([]byte(p))
			if n != len(p) {
				t.Errorf("Test %v(%v): n (%v) != expected (%v)",
					k, x, n, len(p))
			}
			if err != nil {
				t.Errorf("Test %v(%v): err (%v) != expected (%v)",
					k, x, err, nil)
			}
			if !reflect.DeepEqual(buf.String(), test.expected[x]) {
				t.Errorf("Test %v(%v): written (%v) != expected (%v)",
					k, x, buf.String(), test.expected[x])
			}
			buf.Reset()
		}
		err := w.Close()
		if err != nil {
			t.Errorf("Test %v: err (%v) != expected (%v)",
				k, err, nil)
		}
		if !reflect.DeepEqual(buf.String(), test.last) {
			t.Errorf("Test %v: last (%v) != expected (%v)",
				k, buf.String(), test.last)
		}
		// Set the error and verify the return.
		w.(*last).err = io.EOF
		if n, err := w.Write(nil); n != 0 && err != io.EOF {
			t.Errorf("Test %v: final write didn't return 0, EOF: %v %v",
				k, n, err)
		}
	}
	// Test the special error cases.
	if NewLastFuncWriter(tests[0].f, nil) != nil {
		t.Errorf("nil io.Writer didn't return nil.")
	}
	if NewLastFuncWriter(nil, ioutil.Discard) != nil {
		t.Errorf("nil func did't return nil.")
	}
}

// Er is a helper for testing reads. It always writes the given data
// to p and returns the given values.
type er struct {
	data []byte
	n    int
	err  error
}

func (e er) Read(p []byte) (int, error) {
	copy(p, e.data)
	return e.n, e.err
}

// Ew is a helper for testing the writers that need to error out. Any
// call to Write() will produce the err.
type ew struct {
	err error
}

func (e ew) Write(p []byte) (int, error) {
	return 0, e.err
}
