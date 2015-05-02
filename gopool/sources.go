package gopool

import (
	"container/heap"
	"fmt"
	"log"
)

// Sourcer is the interface that allows a type to be run as a source
// that communicates approptiately with a gopool. The Next() and Add()
// methods are synchronized internally, so as long as no other places
// are calling them, they won't suffer from race conditions. If they
// might be called concurrently, it is the implementers responsibility
// to synchronize usage (e.g. through a mutex).
type Sourcer interface {
	fmt.Stringer

	// Next returns the next task from the source. It should return nil
	// if there is currently no work.
	Next() Task

	// Add schedules a task. It aslo reschedules a task during cleanup
	// if a task was taken but was unable to be sent. As such, it should
	// be available until the goroutine managing it is done.
	Add(t Task)
}

// NewSource creates a managed source using the given Sourcer and
// starts a goroutine that synchronizes access to the given
// Interface. If a wakeup channel is non-nil, it can be used to force
// the goroutine to wakeup and look for new tasks. If verbose is true
// things happening in the channel are logged to default logger. The
// returned channels are the source, add and stop channels
// respectively. The returned source channel is used for getting
// tasks. The add channel is used to add tasks elsewhere.
//
// The stop channel is used to stop the running goroutine. When it is
// time to stop, simply send a new channel down that channel. When the
// goroutine has cleaned up, the given channel will be closed. For
// example:
//
//    src, add, stop := New(s, false, nil)
//    done := make(chan struct{})
//    stop <- done // Send the channel we are goign to wait on.
//    <- done      // Once this returns, thing are cleaned up.
func NewSource(s Sourcer, verbose bool, wakeup chan struct{}) (<-chan Task,
	chan<- Task, chan chan struct{}) {
	source := make(chan Task)
	stop := make(chan chan struct{})
	add := make(chan Task)
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
			case c := <-stop:
				if verbose {
					log.Printf("[source %v] stop requested", s)
				}
				if top != nil {
					s.Add(top)
					if verbose {
						log.Printf("[source %v] added back task %v", s, top)
					}
				}
				if c != nil {
					close(c)
				}
				return
			case src <- top:
				if verbose {
					log.Printf("[source %v] sent task %v", s, top)
				}
				top = nil
			}
		}
	}()
	return source, add, stop
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

func (t *pt) String() string            { return t.t.String() }
func (t *pt) Priority() int             { return t.p }
func (t *pt) Run(s chan struct{}) error { return t.t.Run(s) }

// PriorityQueue is an implementation of Interface using a priority
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
