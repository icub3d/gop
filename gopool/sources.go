// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package gopool

import (
	"container/heap"
	"fmt"
	"log"
	"sync"

	"golang.org/x/net/context"
)

// Sourcer is the interface that allows a type to be run as a source
// that communicates approptiately with a gopool. If this Sourcer is
// used by a managed source, the Next() and Add() methods are
// synchronized internally, so as long as no other places are calling
// them, they won't suffer from race conditions. If they might be
// called concurrently, it is the implementers responsibility to
// synchronize usage (e.g. through a mutex).
type Sourcer interface {
	fmt.Stringer

	// Next returns the next task from the source. It should return nil
	// if there is currently no work.
	Next() Task

	// Add schedules a task. It also reschedules a task during cleanup
	// if a task was taken but was unable to be sent. As such, it should
	// be available until the ManagedSource using it returns from a call
	// to Wait().
	Add(t Task)
}

// ManagedSource wraps a Sourcer in a goroutine that synchronizes
// access to the Sourcer.
type ManagedSource struct {
	// Source is the channel where tasks can be retrieved.
	Source <-chan Task

	// Add is the channel on which tasks can be added.
	Add chan<- Task

	wg *sync.WaitGroup
}

// Wait blocks until the ManagedSource is done. If you want to ensure
// that the managed source has cleaned up completely, you should call
// this.
func (ms *ManagedSource) Wait() {
	ms.wg.Wait()
}

// NewManagedSource creates a managed source using the given Sourcer and
// starts it. If the wakeup channel is non-nil, it can be used to force
// the goroutine to wakeup and look for new tasks. This may be useful
// if the source may be empty but is later filled.
//
// If verbose is true things happening in the channel are logged to
// the default logger.
func NewManagedSource(s Sourcer, verbose bool, wakeup chan struct{},
	ctx context.Context) *ManagedSource {
	source := make(chan Task)
	add := make(chan Task)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var src chan Task
		var top Task
		for {
			// Get the top task if we don't have one.
			if top == nil {
				top = s.Next()
			}
			// Setup the src channel based on the availability of a task.
			src = source
			if top == nil {
				if verbose {
					log.Printf("[source %v] no task available, none will be sent", s)
				}
				src = nil
			}
			select {
			case _, ok := <-wakeup:
				if !ok {
					wakeup = nil
					if verbose {
						log.Printf("[source %v] wakeup closed, no longer selecting with it", s)
					}
				} else {
					if verbose {
						log.Printf("[source %v] got a wakeup signal", s)
					}
				}
			case t, ok := <-add:
				if !ok {
					add = nil
					if verbose {
						log.Printf("[source %v] add closed, no longer selecting with it", s)
					}
				}
				if t != nil {
					s.Add(t)
					if verbose {
						log.Printf("[source %v] added task %v", s, t)
					}
				}
			case <-ctx.Done():
				if verbose {
					log.Printf("[source %v] stop requested", s)
				}
				if top != nil {
					s.Add(top)
					if verbose {
						log.Printf("[source %v] added back task %v", s, top)
					}
				}
				wg.Done()
				return
			case src <- top:
				if verbose {
					log.Printf("[source %v] sent task %v", s, top)
				}
				top = nil
			}
		}
	}()
	return &ManagedSource{Source: source, Add: add, wg: &wg}
}

// PriorityTask is a Task that has a priority.
type PriorityTask interface {
	Task
	Priority() int
}

// NewPriorityTask returns a PriorityTask with the given task
// and priority.
func NewPriorityTask(t Task, priority int) PriorityTask {
	return &pt{
		p: priority,
		t: t,
	}
}

// pt is an internal implementation of the PriorityTask.
type pt struct {
	p int
	t Task
}

func (t *pt) String() string        { return t.t.String() }
func (t *pt) Priority() int         { return t.p }
func (t *pt) Run(c context.Context) { t.t.Run(c) }

// PriorityQueue is an implementation of a Sourcer using a priority
// queue. Higher priority tasks will be done first.
type PriorityQueue struct {
	q    *pq
	name string
}

// NewPriorityQueue creates a new PriorityQueue.
func NewPriorityQueue(name string) *PriorityQueue {
	q := &PriorityQueue{q: &pq{}, name: name}
	heap.Init(q.q)
	return q
}

func (q *PriorityQueue) String() string {
	return q.name
}

// Next implements Sourcer.Next.
func (q *PriorityQueue) Next() Task {
	if q.q.Len() < 1 {
		return nil
	}
	return heap.Pop(q.q).(Task)
}

// Add implements Sourcer.Add.
func (q *PriorityQueue) Add(t Task) {
	if p, ok := t.(PriorityTask); ok {
		heap.Push(q.q, p)
	} else {
		heap.Push(q.q, NewPriorityTask(t, 0))
	}
}

// Our internal representation of priority queue.
type pq []PriorityTask

func (q pq) Len() int           { return len(q) }
func (q pq) Less(i, j int) bool { return q[i].Priority() > q[j].Priority() }
func (q pq) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *pq) Push(x interface{}) {
	*q = append(*q, x.(PriorityTask))
}

func (q *pq) Pop() interface{} {
	old := *q
	n := len(old)
	t := old[n-1]
	*q = old[0 : n-1]
	return t
}
