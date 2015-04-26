// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package etcdutil

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// testNodes are used in some of the tests. It shouldn't be modified.
var testNodes = etcd.Nodes{
	&etcd.Node{
		Key: "/myport/test/k0-0",
		Dir: true,
		Nodes: etcd.Nodes{
			&etcd.Node{
				Key: "/myport/test/k0-0/k1-0",
				Dir: true,
				Nodes: etcd.Nodes{
					&etcd.Node{
						Key:   "/myport/test/k0-0/k1-0/k2-0",
						Value: "v2-0",
					},
					&etcd.Node{
						Key:   "/myport/test/k0-0/k1-0/k2-1",
						Value: "v2-1",
					},
					&etcd.Node{
						Key:   "/myport/test/k0-0/k1-0/k2-2",
						Value: "v2-2",
					},
				},
			},
			&etcd.Node{
				Key:   "/myport/test/k0-0/k1-1",
				Value: "v1-1",
			},
			&etcd.Node{
				Key:   "/myport/test/k0-0/k1-2",
				Value: "v1-2",
			},
		},
	},
	&etcd.Node{
		Key:   "/myport/test/k0-1",
		Value: "v0-1",
	},
}

func TestNew(t *testing.T) {
	e := NewFromString("test,test1,test2", "/myport/app")
	if e.p != "/myport/app" {
		t.Errorf("prefix not updated: %v %v", e.p, "/myport/app")
	}
	if e.s == nil {
		t.Errorf("stop channel nil")
	}
	if e.c == nil {
		t.Errorf("etcd client nil")
	}
}

func TestClient(t *testing.T) {
	ec := &EtcdUtil{c: &ecs{}}
	if ec.Client() != nil {
		t.Errorf("Client(): expected nil for non etcd.Client")
	}
	ec = &EtcdUtil{c: etcd.NewClient([]string{"https://localhost:4001"})}
	if ec.Client() != ec.c {
		t.Errorf("Client(): expected non-nil for non etcd.Client")
	}
}

// The Get*/MustGet* functions are fairly similar, so the first is
// commented and the rest do the same things.

func TestGet(t *testing.T) {
	tests := []struct {
		key string // The key to search for.
		def string // The default value.
		val string // The value we expect to get back.
		err error  // The error we expect to be returned.
		e   ecs    // The fake etcd client.
	}{
		// A normal get.
		{
			key: "key0",
			def: "bad",
			val: "val0",
			err: nil,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
			},
		},

		// An error condition.
		{
			key: "key0",
			def: "good",
			val: "val0",
			err: etcd.ErrWatchStoppedByUser,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
	}

	for k, test := range tests {
		// Setup the util and call Get.
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val, _, err := ec.Get(test.key, test.def)
		if test.err != nil {
			// If we are expecting an error, we need to test for it and the
			// default value.
			if err != test.err {
				t.Errorf("Test %v: wanted error '%v' but got '%v'", k, test.err, err)
			}
			if val != test.def {
				t.Errorf("Test %v: wanted default value '%v' with error but got '%v'", k, test.def, val)
			}
			continue
		}

		// Otherwise, make sure we don't get an error and we get the
		// expected value.
		if err != nil {
			t.Errorf("Test %v: got unexpected non-nil err: %v", k, err)
			continue
		}
		if val != test.val {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestMustGet(t *testing.T) {
	tests := []struct {
		key string // The key to search for.
		val string // The expected value.
		p   bool   // Whether or not a panic is expected.
		e   ecs    // The fake etcd client.
	}{
		// Normal get.
		{
			key: "key0",
			val: "val0",
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
			},
		},
		// An error condition.
		{
			key: "key0",
			val: "val0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
	}

	// Recover from any panics and update our panic state.
	p := false
	defer func() {
		if r := recover(); r != nil {
			p = true
		}
	}()

	for k, test := range tests {
		// Reset our panic state, setup the util, and call MustGet.
		p = false
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val, _ := ec.MustGet(test.key)
		// Make sure we did/didn't panic based on the test.
		if test.p && !p {
			t.Errorf("Test %v: expected panic, but didn't get it.", k)
			continue
		}
		// Make sure we got back the expected value.
		if val != test.val {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		key string
		def int
		val int
		err error
		e   ecs
	}{
		{
			key: "key0",
			def: -1,
			val: 1234,
			err: nil,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "1234"}},
			},
		},
		{
			key: "key0",
			def: -1,
			val: -1,
			err: etcd.ErrWatchStoppedByUser,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "1234"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
		{
			key: "key0",
			def: -1,
			val: -1,
			err: errors.New("strconv.ParseInt: parsing \"$@$@#\": invalid syntax"),
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "$@$@#"}},
			},
		},
	}

	for k, test := range tests {
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val, _, err := ec.GetInt(test.key, test.def)
		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("Test %v: wanted error '%v' but got '%v'", k, test.err, err)
			}
			if val != test.def {
				t.Errorf("Test %v: wanted default value '%v' with error but got '%v'", k, test.def, val)
			}
			continue
		}

		if err != nil {
			t.Errorf("Test %v: got unexpected non-nil err: %v", k, err)
			continue
		}
		if val != test.val {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestMustGetInt(t *testing.T) {
	tests := []struct {
		key string
		val int
		p   bool
		e   ecs
	}{
		{
			key: "key0",
			val: 123,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "123"}},
			},
		},
		{
			key: "key0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
		{
			key: "key0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "#&#&#&#"}},
			},
		},
	}

	p := false
	defer func() {
		if r := recover(); r != nil {
			p = true
		}
	}()

	for k, test := range tests {
		p = false
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val, _ := ec.MustGetInt(test.key)
		if test.p && !p {
			t.Errorf("Test %v: expected panic, but didn't get it.", k)
			continue
		}
		if val != test.val {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestGetDuration(t *testing.T) {
	tests := []struct {
		key string
		def time.Duration
		val time.Duration
		err error
		e   ecs
	}{
		{
			key: "key0",
			def: -1,
			val: 2 * time.Minute,
			err: nil,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "2m"}},
			},
		},
		{
			key: "key0",
			def: -1,
			val: -1,
			err: etcd.ErrWatchStoppedByUser,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "2m"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
		{
			key: "key0",
			def: -1,
			val: -1,
			err: errors.New("time: invalid duration $@$@#"),
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "$@$@#"}},
			},
		},
	}

	for k, test := range tests {
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val, _, err := ec.GetDuration(test.key, test.def)
		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("Test %v: wanted error '%v' but got '%v'", k, test.err, err)
			}
			if val != test.def {
				t.Errorf("Test %v: wanted default value '%v' with error but got '%v'", k, test.def, val)
			}
			continue
		}

		if err != nil {
			t.Errorf("Test %v: got unexpected non-nil err: %v", k, err)
			continue
		}
		if val != test.val {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestMustGetDuration(t *testing.T) {
	tests := []struct {
		key string
		val time.Duration
		p   bool
		e   ecs
	}{
		{
			key: "key0",
			val: 2 * time.Minute,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "2m"}},
			},
		},
		{
			key: "key0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
		{
			key: "key0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "#&#&#&#"}},
			},
		},
	}

	p := false
	defer func() {
		if r := recover(); r != nil {
			p = true
		}
	}()

	for k, test := range tests {
		p = false
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val, _ := ec.MustGetDuration(test.key)
		if test.p && !p {
			t.Errorf("Test %v: expected panic, but didn't get it.", k)
			continue
		}
		if val != test.val {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestGetJSON(t *testing.T) {
	type tv struct {
		Name string
		Age  int
	}

	tests := []struct {
		key string
		val tv
		err error
		e   ecs
	}{
		{
			key: "key0",
			val: tv{Name: "Test", Age: 33},
			err: nil,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: `{"Name": "Test", "Age": 33}`}},
			},
		},
		{
			key: "key0",
			err: etcd.ErrWatchStoppedByUser,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: `{"Name": "Test", "Age": 33}`}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
		{
			key: "key0",
			err: errors.New("invalid character '$' looking for beginning of value"),
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "$@$@#"}},
			},
		},
	}

	for k, test := range tests {
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		mtv := tv{}
		_, err := ec.GetJSON(test.key, &mtv)
		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("Test %v: wanted error '%v' but got '%v'", k, test.err, err)
			}
			if !reflect.DeepEqual(mtv, test.val) {
				t.Errorf("Test %v: wanted value '%v' with error but got '%v'", k, test.val, mtv)
			}
			continue
		}

		if err != nil {
			t.Errorf("Test %v: got unexpected non-nil err: %v", k, err)
			continue
		}
		if !reflect.DeepEqual(mtv, test.val) {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, mtv)
		}
	}
}

func TestMustGetJSON(t *testing.T) {
	type tv struct {
		Name string
		Age  int
	}

	tests := []struct {
		key string
		val tv
		p   bool
		e   ecs
	}{
		{
			key: "key0",
			val: tv{Name: "Test", Age: 33},
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: `{"Name": "Test", "Age": 33}`}},
			},
		},
		{
			key: "key0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "val0"}},
				err:   etcd.ErrWatchStoppedByUser,
			},
		},
		{
			key: "key0",
			p:   true,
			e: ecs{
				nodes: etcd.Nodes{&etcd.Node{Key: "/myport/test/key0", Value: "#&#&#&#"}},
			},
		},
	}

	p := false
	defer func() {
		if r := recover(); r != nil {
			p = true
		}
	}()

	for k, test := range tests {
		p = false
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		val := tv{}
		ec.MustGetJSON(test.key, &val)
		if test.p && !p {
			t.Errorf("Test %v: expected panic, but didn't get it.", k)
			continue
		}
		if !reflect.DeepEqual(val, test.val) {
			t.Errorf("Test %v: Expected value %v but got %v", k, test.val, val)
		}
	}
}

func TestWatch(t *testing.T) {
	// Chaing the startWait so our test doesn't last forever.
	startWait = 1 * time.Millisecond
	e := ecs{nodes: testNodes, c: make(chan *etcd.Response), r: make(chan ret)}
	ec := &EtcdUtil{
		p: "/myport/test",
		c: &e,
		s: make(chan bool),
	}

	// The WaitGroup will help this goroutine hold until everything is
	// processed.
	var wg sync.WaitGroup
	wg.Add(2)
	var res []string
	f := func(key, val string) {
		res = append(res, key+"|"+val)
		wg.Done()
	}

	// Start the watch.
	ec.Watch("/myport/test/k0-0", 0, true, f)

	// Send a return value so we can test the retry.
	e.r <- ret{nil, errors.New("retry")}
	time.Sleep(5 * time.Millisecond)

	// Send a could responses to our Watch.
	e.c <- &etcd.Response{
		Node: &etcd.Node{
			Key:   "/myport/test/k0-0/k1-0",
			Value: "val1",
		},
	}
	e.c <- &etcd.Response{
		Node: &etcd.Node{
			Key:   "/myport/test/k0-0/k1-1",
			Value: "val2",
		},
	}

	// Cleanup and check the results.
	wg.Wait()
	ec.Close()
	exp := []string{
		"/myport/test/k0-0/k1-0|val1",
		"/myport/test/k0-0/k1-1|val2",
	}
	if !reflect.DeepEqual(res, exp) {
		t.Errorf("Expecting %v but got %v", exp, res)
	}
}

func TestWatchRetryClose(t *testing.T) {
	// We do the same above, we just need to check the return works while we are in the retry loop.
	startWait = 1 * time.Second
	e := ecs{nodes: testNodes, c: make(chan *etcd.Response), r: make(chan ret)}
	ec := &EtcdUtil{
		p: "/myport/test",
		c: &e,
		s: make(chan bool),
	}

	var res []string
	f := func(key, val string) {
		res = append(res, key+"|"+val)
	}

	ec.Watch("/myport/test/k0-0", 0, true, f)

	// Send a return value so we can test the retry.
	e.r <- ret{nil, errors.New("retry")}
	go ec.Close()
	time.Sleep(5 * time.Millisecond)
}

func TestWalk(t *testing.T) {
	tests := []struct {
		key      string   // The key to walk through
		e        ecs      // The fake etcd client.
		err      error    // The error to return.
		errCount int      // When to return the above error.
		exp      []string // The expected results.
		expErr   error    // The expected error.
	}{
		// Failed get.
		{
			key: "doesn't matter",
			e: ecs{
				nodes: testNodes,
				err:   etcd.ErrWatchStoppedByUser,
			},
			expErr: etcd.ErrWatchStoppedByUser,
		},
		// Stop part way through.
		{
			key: "k0-0",
			e: ecs{
				nodes: testNodes,
			},
			err:      etcd.ErrWatchStoppedByUser,
			errCount: 2,
			exp: []string{
				"/myport/test/k0-0/k1-0/k2-0|v2-0",
				"/myport/test/k0-0/k1-0/k2-1|v2-1",
			},
			expErr: etcd.ErrWatchStoppedByUser,
		},
		// No errors.
		{
			key: "k0-0",
			e: ecs{
				nodes: testNodes,
			},
			exp: []string{
				"/myport/test/k0-0/k1-0/k2-0|v2-0",
				"/myport/test/k0-0/k1-0/k2-1|v2-1",
				"/myport/test/k0-0/k1-0/k2-2|v2-2",
				"/myport/test/k0-0/k1-1|v1-1",
				"/myport/test/k0-0/k1-2|v1-2",
			},
		},
		// Find a deep key.
		{
			key: "k0-0/k1-0/k2-0",
			e: ecs{
				nodes: testNodes,
			},
			exp: []string{
				"/myport/test/k0-0/k1-0/k2-0|v2-0",
			},
		},
	}

	for k, test := range tests {
		ec := &EtcdUtil{p: "/myport/test", c: &test.e, s: make(chan bool)}
		count := 0
		var res []string
		wf := func(key, value string) error {
			if count == test.errCount && test.err != nil {
				return test.err
			}
			res = append(res, key+"|"+value)
			count++
			return nil
		}

		// Walk and then test the results.
		_, err := ec.Walk(test.key, false, wf)
		if err != test.expErr {
			t.Errorf("Test %v: Unexpected error, wanted '%v' but got '%v'",
				k, test.expErr, err)
		}
		if !reflect.DeepEqual(res, test.exp) {
			t.Errorf("Test %v: Unexpected error, wanted '%v' but got '%v'",
				k, test.exp, res)
		}
	}
}

// ret is used to send on a channel to force a return.
type ret struct {
	r   *etcd.Response
	err error
}

type ecs struct {
	nodes etcd.Nodes // The nodes we are going to search through.
	err   error      // The error to return on the next call.
	c     chan *etcd.Response
	r     chan ret
}

func (e *ecs) Close() {
	e.nodes = nil
}

// We can ignore sort because we don't really use it in the
// util. We'll also ignore recur becase the test should setup the
// nodes for us. We just need to find it.
func (e *ecs) Get(key string, sort, recur bool) (*etcd.Response, error) {
	if e.err != nil {
		err := e.err
		e.err = nil
		return nil, err
	}

	// Our testing doesn't require anything else.
	return &etcd.Response{
		Node: findNode(key, e.nodes),
	}, nil
}

func findNode(key string, nodes etcd.Nodes) *etcd.Node {
	for _, node := range nodes {
		if node.Key == key {
			return node
		} else if len(node.Nodes) > 0 {
			found := findNode(key, node.Nodes)
			if found != nil {
				return found
			}
		}
	}
	return nil
}

// We can ignore the index because we don't really use it. We can
// ignore recur as well because the test should setup the nodes for
// us.
func (e *ecs) Watch(key string, index uint64, recur bool, out chan *etcd.Response, stop chan bool) (*etcd.Response, error) {
	for {
		select {
		case r := <-e.r:
			return r.r, r.err
		case <-stop:
			return nil, etcd.ErrWatchStoppedByUser
		case r := <-e.c:
			select {
			case r := <-e.r:
				return r.r, r.err
			case <-stop:
				return nil, etcd.ErrWatchStoppedByUser
			case out <- r:
			}
		}
	}
}
