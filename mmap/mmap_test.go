package mmap

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
	"testing"
)

func TestMmap(t *testing.T) {
	// Get a temp file name.
	file, err := ioutil.TempFile("", "test_mmap")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("closing temp file: %v", err)
	}
	if err = os.Remove(file.Name()); err != nil {
		t.Fatalf("removing temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("removing mapped file: %v", err)
		}
	}()

	// Create a file.
	m, err := New(file.Name(), 0644, os.O_CREATE|os.O_APPEND|os.O_RDWR, math.MaxInt32, false)
	if err != nil {
		t.Fatalf("New(%v, 0644, os.O_CREATE|os.O_APPEND|os.O_RDWR, math.MaxInt32, false): %v",
			file.Name(), err)
	}

	// Write some data to the middle of it, sync, and close.
	i := copy(m.Buf[1029010:], []byte("Hello, world!"))
	if i != 13 {
		t.Fatalf("didn't write 13")
	}
	if bytes.Compare(m.Buf[1029010:1029023], []byte("Hello, world!")) != 0 {
		t.Errorf("data was not written in buf properly: %v",
			string(m.Buf[1029010:1029023]))
	}

	if err := m.Sync(); err != nil {
		t.Fatalf("Sync(): %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}

	// Open it again.
	m, err = New(file.Name(), 0644, os.O_CREATE|os.O_APPEND|os.O_RDWR, math.MaxInt32, true)
	if err != nil {
		t.Fatalf("New(%v, 0644, os.O_CREATE|os.O_APPEND|os.O_RDWR, math.MaxInt32, true)): %v",
			file.Name(), err)
	}

	// Verify the data is still in the middle and cleanup.
	if bytes.Compare(m.Buf[1029010:1029023], []byte("Hello, world!")) != 0 {
		t.Errorf("data was not written properly: %v", string(m.Buf[1029010:1029023]))
	}
	if err := m.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
}
