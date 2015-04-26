package signalhandler

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
)

func ExampleSignalHandler() {
	// We'll use a wait group here so we don't exit before we get the
	// signal. Under normal circumstances, this wouldn't be necessary.
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Create the handler and add a watch on SigHup.
	h := New()
	h.Watch(SigHup, func() {
		fmt.Println("reloading config")
		wg.Done()
	})

	// Send ourselves a signal.
	syscall.Kill(os.Getpid(), SigHup)

	// Wait for the signal. Again, this isn't necessary under normal
	// circumstatnces.
	wg.Wait()

	// Stop listening.
	h.Stop()

	// Output:
	// hangup 1
	// reloading config
}

func TestSignalHandler(t *testing.T) {
	tests := []struct {
		toWatch []struct {
			sig os.Signal
			f   func(*bytes.Buffer)
		}
		signals  []syscall.Signal
		expected string
		wait     int
	}{
		// Test a simple single signal.
		{
			toWatch: []struct {
				sig os.Signal
				f   func(*bytes.Buffer)
			}{
				{
					sig: SigHup,
					f: func(b *bytes.Buffer) {
						b.WriteString("a")
					},
				},
			},
			signals:  []syscall.Signal{SigHup, SigHup, SigHup},
			expected: "aaa",
			wait:     3,
		},
		// Test multiple signals one with a duplicate.
		{
			toWatch: []struct {
				sig os.Signal
				f   func(*bytes.Buffer)
			}{
				{
					sig: SigHup,
					f: func(b *bytes.Buffer) {
						b.WriteString("a")
					},
				},
				{
					sig: SigUsr1,
					f: func(b *bytes.Buffer) {
						b.WriteString("b")
					},
				},
				{
					sig: SigUsr1,
					f: func(b *bytes.Buffer) {
						b.WriteString("c")
					},
				},
			},
			signals: []syscall.Signal{
				SigHup, SigHup, SigHup,
				SigUsr1, SigUsr1, SigUsr1,
				SigTerm, SigTerm,
				SigHup, SigHup,
			},
			expected: "aaabcbcbcaa",
			wait:     11,
		},
	}

	for k, test := range tests {
		// Do some prep work.
		h := New()
		b := &bytes.Buffer{}
		// We want to wait because the signals are asynchronous.
		wg := &sync.WaitGroup{}
		wg.Add(test.wait)
		// Not sure why, but during testing, we need to watch it like this
		// as well. I've used the code in normal places and it works
		// great.
		c := make(chan os.Signal, 1)
		signal.Notify(c)
		for _, w := range test.toWatch {
			tw := w
			h.Watch(tw.sig, func() {
				tw.f(b)
				wg.Done()
			})
		}
		for _, sig := range test.signals {
			syscall.Kill(os.Getpid(), sig)
			<-c
		}
		wg.Wait()
		if s := b.String(); s != test.expected {
			t.Errorf("Test %v: expected output failed:\n%v\n%v", k,
				test.expected, s)
		}
		h.Stop()
	}
}
