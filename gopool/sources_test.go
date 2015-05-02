package gopool

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewSource(t *testing.T) {
	// TODO - this isn't very rigorous and uses time.Sleep.
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	// Setup
	pq := NewPriorityQueue("test")
	wakeup := make(chan struct{})
	src, add, stop := NewSource(pq, true, wakeup)

	// Make sure the top isn't selecting.
	select {
	case c, ok := <-src:
		t.Errorf("got task from empty: %v %v", ok, c)
	default:
	}

	// Send a wakeup.
	wakeup <- struct{}{}
	// Close a wakeup.
	close(wakeup)
	time.Sleep(20 * time.Millisecond)

	// Add two items.
	add <- NewPriorityTask(&sct{name: "first"}, 5)
	add <- NewPriorityTask(&sct{name: "third"}, 7)
	add <- NewPriorityTask(&sct{name: "second"}, 10)
	// Close the add.
	close(add)
	time.Sleep(20 * time.Millisecond)

	// Get one item
	c := <-src
	if c.String() != "first" {
		t.Errorf("Didn't get first task.")
	}

	// Cleanup
	done := make(chan struct{})
	stop <- done
	<-done

	// Make sure the task was added back:
	if pq.q.Len() != 2 {
		t.Errorf("pq.q.Len() != 2 after close: %v", pq.q.Len())
	}

	// Check the logs.
	entries := strings.Split(buf.String(), "\n")
	exp := []string{
		"no task available",
		"got a wakeup signal",
		"no task available",
		"wakeup closed",
		"no task available",
		"added task first",
		"added task third",
		"added task second",
		"add closed",
		"sent task first",
		"stop requested",
		"added back task second",
		"",
	}
	if len(entries) != len(exp) {
		for _, e := range entries {
			t.Logf("%s", e)
		}
		t.Fatalf("Didn't get back expected entries: %v %v", len(entries), len(exp))
	}
	for x, entry := range entries {
		if !strings.Contains(entry, exp[x]) {
			t.Errorf("Expected '%v' in entry, but got '%v'", exp[x], entry)
		}
	}
}

func TestPriorityTask(t *testing.T) {
	buf := &bytes.Buffer{}
	c := &sct{name: "test", w: buf}
	pt := NewPriorityTask(c, 42)
	s := make(chan struct{})
	pt.Run(s)
	if c.s != s {
		t.Errorf("stop channel didn't work.")
	}
	if buf.String() != "test" {
		t.Errorf(`Run() didn't run, buf.String() != "test": %v`, buf.String())
	}
	if pt.Priority() != 42 {
		t.Errorf(`pt.Priority() != 42: %v`, pt.Priority())
	}
	if pt.String() != "test" {
		t.Errorf(`pt.String() != "test": %v`, pt.String())
	}
}

func TestPriorityQueue(t *testing.T) {
	q := NewPriorityQueue("test")

	// Check the stringer.
	if q.String() != "test" {
		t.Errorf(`q.String() != "test": %v`, q.String())
	}
	// Verify empty is nil.
	if q.Next() != nil {
		t.Fatalf("q.Next() != nil after NewPriorityQueue()")
	}

	// Add one and get it.
	c := &sct{name: "test"}
	q.Add(c)
	p := q.Next().(PriorityTask)
	if p.Priority() != 0 {
		t.Fatalf("non-PriorityTask was given a non-zero priority: %v", p.Priority())
	}
	if p.(*pt).t != c {
		t.Fatalf("task wasn't properly initialized with given task")
	}
	// Verify empty is nil after reading them all.
	if q.Next() != nil {
		t.Fatalf("q.Next() != nil after last Next()")
	}

	// Add a bunch and make sure we get them back in the right order.
	buf := &bytes.Buffer{}
	for x := 0; x < 10; x += 2 {
		q.Add(NewPriorityTask(&sct{name: strconv.Itoa(x), w: buf}, x))
	}
	for x := 1; x < 10; x += 2 {
		q.Add(NewPriorityTask(&sct{name: strconv.Itoa(x), w: buf}, x))
	}
	// Read them all back and run them.
	for c := q.Next(); c != nil; c = q.Next() {
		c.Run(nil)
	}
	if buf.String() != "9876543210" {
		t.Errorf("priority wasn't properly applied. Expected 9876543210, but got %v", buf.String())
	}
}

// sct is a helper for testing that bascially just prints it's name to
// w and sets the stop channel.
type sct struct {
	name string
	s    chan struct{}
	w    io.Writer
}

func (t *sct) Run(s chan struct{}) error { t.s = s; t.w.Write([]byte(t.name)); return nil }
func (t *sct) String() string            { return t.name }
