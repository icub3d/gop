// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package gopool

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestGoPool(t *testing.T) {
	// TODO this isn't very rigorous.
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	var l sync.Mutex
	items := []int{}
	exp := []int{}
	ai := func(i int) {
		l.Lock()
		defer l.Unlock()
		items = append(items, i)
	}

	// Do the setup.
	pq := NewPriorityQueue("test-queue")
	wakeup := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	ms := NewManagedSource(pq, true, wakeup, ctx)
	pool := New("test-pool", 5, true, ctx, ms.Source)

	// Add a bunch of tasks and wait for them to finish.
	for x := 0; x < 500; x++ {
		ms.Add <- &tt{f: ai, i: x}
		exp = append(exp, x)
	}
	for pq.q.Len() > 1 {
		time.Sleep(20 * time.Millisecond)
	}
	for x := 500; x < 1000; x++ {
		ms.Add <- &tt{f: ai, i: x}
		exp = append(exp, x)
	}
	for pq.q.Len() > 1 {
		time.Sleep(20 * time.Millisecond)
	}

	// Cleanup
	cancel()
	pool.Wait()
	ms.Wait()

	// Verify the list.
	sort.Ints(items)
	if !reflect.DeepEqual(items, exp) {
		t.Errorf("Didn't get all the items expected: %v, %v", len(items), len(exp))
	}

	// Verify we got some expected log values.
	log := buf.String()
	for x := 0; x < 4; x++ {
		e := fmt.Sprintf("[gopool test-pool %d] stop channel closed: stopping", x)
		if !strings.Contains(log, e) {
			t.Errorf("Didn't get an entry for: %v", e)
		}
	}
}

func TestGoPoolInputSourceClosed(t *testing.T) {
	// TODO this isn't very rigorous and relies on time.Sleep.
	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	// Do the setup.
	src := make(chan Task)
	pool := New("test-pool", 5, true, context.Background(), src)

	close(src)
	time.Sleep(10 * time.Millisecond)

	// Cleanup
	pool.Wait()

	// Verify we go some expected log values.
	log := buf.String()
	e := " input source closed: stopping"
	if !strings.Contains(log, e) {
		t.Errorf("Didn't get an entry for: %v", e)
	}
}

type tt struct {
	f func(int)
	i int
}

func (t *tt) String() string { return strconv.Itoa(t.i) }
func (t *tt) Run(ctx context.Context) {
	t.f(t.i)
}
