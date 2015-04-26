// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package flock provides a simple file locking mechanism for
// linux/unix based on unix.Flock.
package flock

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// ErrWouldBlock is returned by the non-blocking locks when it would
// have blocked.
var ErrWouldBlock = errors.New("would block")

// Flock is a file based lock mechanism.
type Flock struct {
	f *os.File
}

// New creates a new Flock for the given path name.
func New(name string) (*Flock, error) {
	var err error
	f := &Flock{}
	f.f, err = os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// LockSharedWait attempts to get a shared lock and waits until that
// lock is acquired or an error occurs.
func (f *Flock) LockSharedWait() error {
	return f.call(unix.LOCK_SH)
}

// LockExclusiveWait attempts to get an exclusive lock and waits until
// that lock is acquired or an error occurs.
func (f *Flock) LockExclusiveWait() error {
	return f.call(unix.LOCK_EX)
}

// LockShared attempts to get a shared lock but won't block if it
// can't be immediately acquired. In this case, the return error is
// ErrWouldBlock.
func (f *Flock) LockShared() error {
	return f.call(unix.LOCK_SH | unix.LOCK_NB)
}

// LockExclusive attempts to get an exclusive lock but won't block if
// it can't be immediately acquired. In this case, the return error is
// ErrWouldBlock.
func (f *Flock) LockExclusive() error {
	return f.call(unix.LOCK_EX | unix.LOCK_NB)
}

// Unlock attempts to release the lock you have
func (f *Flock) Unlock() error {
	return f.call(unix.LOCK_UN)
}

// Close closes the open file. This should be called when the lock is
// no longer needed.
func (f *Flock) Close() error {
	return f.f.Close()
}

func (f *Flock) call(flags int) error {
	err := unix.Flock(int(f.f.Fd()), flags)
	if err == unix.EWOULDBLOCK {
		return ErrWouldBlock
	}
	return err
}
