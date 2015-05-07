// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package gopool implements a concurrent work processing model. It is
// a similar to thread pools in other languages, but it uses
// goroutines and channels. A pool is formed wherein several
// goroutines get tasks from a channel. Various sources can be used to
// schedule tasks and given some coordination gopools on various
// systems can work from the same source.
package gopool

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// ErrStopped is used to signal a goroutine that it should stop.
var ErrStopped = errors.New("stop requested")

// Task is a some type of work that the gopool should perform. The
// Stringer interface is used to aid in logging.
type Task interface {
	fmt.Stringer

	// Run performs the work for this task. When the context is done,
	// processing should stop as soon as reasonably possible. Long
	// running tasks should make sure it's watching the context's Done()
	// channel.
	Run(context.Context)
}

// GoPool is a group of goroutines that work on Tasks. Each goroutine
// gets work from a channel until the context signals that it's done.
type GoPool struct {
	name    string
	src     <-chan Task
	wg      sync.WaitGroup
	ctx     context.Context
	verbose bool
}

// New creates a new GoPool with the given number of goroutines. The
// name is used for logging purposes. The goroutines are started as
// part of calling New().
//
// The goroutines will stop when the given context is done. If you
// want to make sure all of the tasks have got the signal and stopped
// cleanly, you should use Wait().
//
// The src channel is where the goroutines look for tasks. If verbose
// is true, information about the work being performed is logged to
// the default logger. Otherwise only unexpected closures or errors
// are logged.
func New(name string, goroutines int, verbose bool, ctx context.Context,
	src <-chan Task) *GoPool {
	p := &GoPool{
		name:    name,
		src:     src,
		ctx:     ctx,
		verbose: verbose,
	}
	for x := 0; x < goroutines; x++ {
		go p.worker(x)
	}
	p.wg.Add(goroutines)
	return p
}

// Wait blocks until all of the workers have stopped. This won't ever
// return if the context for this gopool is never done.
func (p *GoPool) Wait() {
	p.wg.Wait()
}

// String implements the fmt.Stringer interface. It just prints the
// name given to New().
func (p *GoPool) String() string {
	return p.name
}

// Worker is the function each goroutine uses to get and perform
// tasks. It stops when the stop channel is closed. It also stops if
// the source channel is closed but logs a message in addition.
func (p *GoPool) worker(ID int) {
	for {
		select {
		case <-p.ctx.Done():
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
			t.Run(p.ctx)
			if p.verbose {
				log.Printf("[gopool %v %v] finished task (duration %v): %v", p, ID,
					time.Now().Sub(start), t)
			}
		}
	}
}
