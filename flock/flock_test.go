package flock

import (
	"errors"
	"os"
	"testing"
	"time"
)

var ErrTimeout = errors.New("timeout")

// LockorTimeout is a helper function. It calls the given locker
// function and returns the error it got back from it. If the lock
// waits and doesn't recover within 100ms, a timeout occurs and
// ErrTimeout is returned.
func LockOrTimeout(f func() error) chan error {
	fc := make(chan struct{})
	tc := make(chan struct{})
	errc := make(chan error)
	var err error
	go func() {
		err = f()
		close(fc)
	}()
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(tc)
	}()
	go func() {
		select {
		case <-fc:
			errc <- err
		case <-tc:
			errc <- ErrTimeout
		}
	}()
	return errc
}

func TestNew(t *testing.T) {
	// Test something that won't likely exist.
	f, err := New("/some/file/that/likely/wont/exist")
	if f != nil || err == nil {
		t.Errorf(`New("/some/file/that/likely/wont/exist"): ` +
			"expected error, but didn't get one.")
		f.Close()
	}
}

func TestSharedWaitLocks(t *testing.T) {
	defer os.Remove("/tmp/flock_test")

	flocks := make([]*Flock, 2)
	errcs := make([]chan error, 2)
	errs := make([]error, 2)
	for x := 0; x < 2; x++ {
		f, err := New("/tmp/flock_test")
		if err != nil {
			t.Fatalf(`f[%v] = New("/tmp/flock_test"): %v`, x, err)
		}
		defer f.Close()
		defer f.Unlock()
		flocks[x] = f
	}

	// A shared lock shouldn't cause any waits.
	errcs[0] = LockOrTimeout(flocks[0].LockSharedWait)
	errcs[1] = LockOrTimeout(flocks[1].LockSharedWait)
	errs[0] = <-errcs[0]
	errs[1] = <-errcs[1]
	if errs[0] != nil || errs[1] != nil {
		t.Errorf("shared locks returned errors: %v | %v", errs[0], errs[1])
	}
}

func TestExclusiveWaitLocks(t *testing.T) {
	defer os.Remove("/tmp/flock_test")

	flocks := make([]*Flock, 2)
	errcs := make([]chan error, 2)
	errs := make([]error, 2)
	for x := 0; x < 2; x++ {
		f, err := New("/tmp/flock_test")
		if err != nil {
			t.Fatalf(`f[%v] = New("/tmp/flock_test"): %v`, x, err)
		}
		defer f.Close()
		defer f.Unlock()
		flocks[x] = f
	}

	// An exclusive lock should cause one to timeout
	errcs[0] = LockOrTimeout(flocks[0].LockExclusiveWait)
	errcs[1] = LockOrTimeout(flocks[1].LockExclusiveWait)
	errs[0] = <-errcs[0]
	errs[1] = <-errcs[1]
	if errs[0] == nil && errs[1] == nil {
		t.Errorf("both exclusive locks returned nil: %v | %v", errs[0], errs[1])
	}
}

func TestSharedLocks(t *testing.T) {
	defer os.Remove("/tmp/flock_test")

	flocks := make([]*Flock, 2)
	errcs := make([]chan error, 2)
	errs := make([]error, 2)
	for x := 0; x < 2; x++ {
		f, err := New("/tmp/flock_test")
		if err != nil {
			t.Fatalf(`f[%v] = New("/tmp/flock_test"): %v`, x, err)
		}
		defer f.Close()
		defer f.Unlock()
		flocks[x] = f
	}

	// A shared lock shouldn't cause any would blocks.
	errcs[0] = LockOrTimeout(flocks[0].LockShared)
	errcs[1] = LockOrTimeout(flocks[1].LockShared)
	errs[0] = <-errcs[0]
	errs[1] = <-errcs[1]
	if errs[0] != nil || errs[1] != nil {
		t.Errorf("shared locks returned errors: %v | %v", errs[0], errs[1])
	}
}

func TestExclusiveLocks(t *testing.T) {
	defer os.Remove("/tmp/flock_test")

	flocks := make([]*Flock, 2)
	errcs := make([]chan error, 2)
	errs := make([]error, 2)
	for x := 0; x < 2; x++ {
		f, err := New("/tmp/flock_test")
		if err != nil {
			t.Fatalf(`f[%v] = New("/tmp/flock_test"): %v`, x, err)
		}
		defer f.Close()
		defer f.Unlock()
		flocks[x] = f
	}

	// An exclusive lock should cause one to return ErrWouldBlock
	errcs[0] = LockOrTimeout(flocks[0].LockExclusive)
	errcs[1] = LockOrTimeout(flocks[1].LockExclusive)
	errs[0] = <-errcs[0]
	errs[1] = <-errcs[1]
	if errs[0] != ErrWouldBlock && errs[1] != ErrWouldBlock {
		t.Errorf("neither exclusive locks returned ErrWouldBlock: %v | %v", errs[0], errs[1])
	}
}

func TestUnlock(t *testing.T) {
	defer os.Remove("/tmp/flock_test")

	flocks := make([]*Flock, 2)
	errcs := make([]chan error, 2)
	errs := make([]error, 2)
	for x := 0; x < 2; x++ {
		f, err := New("/tmp/flock_test")
		if err != nil {
			t.Fatalf(`f[%v] = New("/tmp/flock_test"): %v`, x, err)
		}
		defer f.Close()
		defer f.Unlock()
		flocks[x] = f
	}

	// We are going to do an exclusive lock that waits, but the first
	// will try to release before things are up, so neither should
	// timeout.
	errcs[0] = LockOrTimeout(func() error {
		flocks[0].LockExclusiveWait()
		time.Sleep(50 * time.Millisecond)
		return flocks[0].Unlock()
	})
	errcs[1] = LockOrTimeout(flocks[1].LockExclusiveWait)
	errs[0] = <-errcs[0]
	errs[1] = <-errcs[1]
	if errs[0] != nil || errs[1] != nil {
		t.Errorf("unlocking returned some errors: %v | %v", errs[0], errs[1])
	}
}
