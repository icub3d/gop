// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package gopool implements a concurrent work processing model. It is
// a similar to thread pools in other languages, but it uses
// goroutines and channels. A pool is formed wherein several
// goroutines get tasks from a channel. Various sources can be used to
// schedule tasks and given some coordination workgroups on various
// systems can work from the same source.
package gopool

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// ErrStopped is used to signal a goroutine that it should stop.
var ErrStopped = errors.New("stop requested")

// Task is a some type of work that the gopool should perform. The
// Stringer interface is used to aid in logging.
type Task interface {
	fmt.Stringer

	// Run performs the work for this task. When the stop channel is
	// closed, processing should stop as soon as reasonably
	// possible. Long running tasks should make sure it's watching that
	// channel.
	//
	// Normally, the only error returned should be ErrStopped when the
	// stop channel is closed. If something fatal happens though and the
	// goroutine doing the work should stop, any other error can be
	// returned. The error is logged to the default logger.
	Run(stop chan struct{}) error
}

// GoPool is a group of goroutines that work on Tasks. Each goroutine
// gets work from a channel until Stop() is called, the source channel
// is closed or the current task tells it to close.
type GoPool struct {
	name    string
	src     <-chan Task
	wg      sync.WaitGroup
	stop    chan struct{}
	verbose bool
}

// New creates a new GoPool with the given number of goroutines. The
// name is used for logging purposes. The goroutines are started as
// part of calling New().
//
// To shut them down call Stop(). Once Stop() returns, all of the
// currently running tasks have finished. It's a good idea to wait if
// you don't want any tasks be stopped at an unexpected state.
//
// The src channel is where the goroutines look for tasks. If verbose
// is true, information about the work being performed is
// logged. Otherwise only unexpected closures or errors are logged.
func New(name string, goroutines int, verbose bool, src <-chan Task) *GoPool {
	p := &GoPool{
		name:    name,
		src:     src,
		stop:    make(chan struct{}),
		verbose: verbose,
	}
	for x := 0; x < goroutines; x++ {
		go p.worker(x)
	}
	p.wg.Add(goroutines)
	return p
}

// Stop gracefully signals all the goroutines to stop. It blocks
// until all of them are done.
func (p *GoPool) Stop() {
	close(p.stop)
	p.wg.Wait()
}

// String implements the fmt.Stringer interface. It just prints the
// name given to New().
func (p *GoPool) String() string {
	return p.name
}

// Worker is the function each goroutine uses to get and perform
// work. It stops when the stop channel is closed. It also stops if
// the source channel is closed but logs a message in addition.
func (p *GoPool) worker(ID int) {
	for {
		select {
		case <-p.stop:
			if p.verbose {
				log.Printf("[gopool %v %v] stop channel closed: stopping", p, ID)
			}
			p.wg.Done()
			return
		case t, ok := <-p.src:
			if !ok {
				log.Printf("[gopool %v %v] input source closed: stopping", p, ID)
				p.wg.Done()
				return
			}
			if p.verbose {
				log.Printf("[gopool %v %v] starting task: %v", p, ID, t)
			}
			start := time.Now()
			err := t.Run(p.stop)
			if err != nil {
				log.Printf("[gopool %v %v] running %v stopped: %v", p, ID, t, err)
				p.wg.Done()
				return
			}
			if p.verbose {
				log.Printf("[gopool %v %v] finished task (duration %v): %v", p, ID,
					time.Now().Sub(start), t)
			}
		}
	}
}
