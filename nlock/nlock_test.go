package nlock

import "testing"

func TestNamedLock(t *testing.T) {
	nl := New()
	nl.Lock("a")
	nl.Lock("b")
	nl.Lock("c")
	nl.Unlock("d")
	nl.Unlock("a")
	nl.Unlock("b")
	nl.Unlock("c")
}
