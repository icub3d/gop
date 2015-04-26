// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

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
