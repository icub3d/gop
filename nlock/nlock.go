// Package nlock provides mutext locking by name.
//
// A NamedLock is a set of mutex locks that are accessibly by a
// name. It is useful in cases where an undefined number of resources
// need to be synchronized. For example, concurrent access to a folder
// within a file system could use a NamedLock where the lock names are
// the file names. One goroutine can acquire a lock by name and then
// be sure that it's work on that file won't be interrupted by other
// files.
package nlock

import "sync"

// NamedLock is used for creating mutex locks by name. It is
// instantiated with the New() function.
type NamedLock struct {
	l sync.Mutex
	m map[string]*sync.Mutex
}

// New creates a new Named lock.
func New() *NamedLock {
	return &NamedLock{
		m: map[string]*sync.Mutex{},
	}
}

// Lock locks the given name. If name is already locked, it blocks
// until the mutex is available.
func (nl *NamedLock) Lock(name string) {
	nl.l.Lock()
	l, ok := nl.m[name]
	if !ok {
		l = &sync.Mutex{}
		nl.m[name] = l
	}
	nl.l.Unlock()
	l.Lock()
}

// Unlock unlocks the given name. It is a run-time error if the name
// is not locked when Unlock is called.
func (nl *NamedLock) Unlock(name string) {
	nl.l.Lock()
	defer nl.l.Unlock()
	l, ok := nl.m[name]
	if !ok {
		return
	}
	l.Unlock()
}
