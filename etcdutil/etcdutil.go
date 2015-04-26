// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package etcdutil provides a helper structure that simplifies common
// etcd operations. The *Get* operations return a uint64 which is the
// etcd index that can be used as a wait index.
package etcdutil

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

var (
	// This is for testing. We may want to expose these in the future.
	maxWait   = 2 * time.Minute
	startWait = 1 * time.Second
)

// ec is the interface to functions we need for our etcd client. It's
// primarily used to make testing without etcd possible.
type ec interface {
	Close()
	Get(string, bool, bool) (*etcd.Response, error)
	Watch(string, uint64, bool, chan *etcd.Response, chan bool) (*etcd.Response, error)
}

// EtcdUtil is the primary structure used in the package. Instantiate
// it with New or NewFromString.
type EtcdUtil struct {
	c ec        // The etcd client.
	p string    // The prefix.
	s chan bool // the watch stop channel.
}

// New creates a utilitiy structure that connects to etcd on the
// following machines. The prefix will be prepended to all keys during
// any requests.
func New(machines []string, prefix string) *EtcdUtil {
	return &EtcdUtil{
		p: prefix,
		c: etcd.NewClient(machines),
		s: make(chan bool),
	}
}

// NewFromString is like new, but takes a comman separated list of
// machines instead of an array.
func NewFromString(machines, prefix string) *EtcdUtil {
	return New(strings.Split(machines, ","), prefix)
}

// Client returns the underlying client if there is one.
func (u *EtcdUtil) Client() *etcd.Client {
	if ec, ok := u.c.(*etcd.Client); ok {
		return ec
	}
	return nil
}

// Get returns the value for the given prefix+key or the default value
// given.
func (u *EtcdUtil) Get(key, def string) (string, uint64, error) {
	k := strings.Join([]string{u.p, key}, "/")
	r, err := u.c.Get(k, false, false)
	if err != nil {
		return def, 0, err
	}
	return r.Node.Value, r.EtcdIndex, nil
}

// MustGet is like Get but panics with the error if an error occurs.
func (u *EtcdUtil) MustGet(key string) (string, uint64) {
	v, i, err := u.Get(key, "")
	if err != nil {
		panic(fmt.Sprintf("MustGet(%v): %v\n", key, err))
	}
	return v, i
}

// GetInt is like Get but returns an integer.
func (u *EtcdUtil) GetInt(key string, def int) (int, uint64, error) {
	s, i, err := u.Get(key, "")
	if err != nil {
		return def, i, err
	}
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return def, i, err
	}
	return int(v), i, nil
}

// MustGetInt is like MustGet but returns an integer.
func (u *EtcdUtil) MustGetInt(key string) (int, uint64) {
	v, i, err := u.GetInt(key, 0)
	if err != nil {
		panic(fmt.Sprintf("MustGetInt(%v): %v\n", key, err))
	}
	return v, i
}

// GetDuration is like Get but returns a duration.
func (u *EtcdUtil) GetDuration(key string, def time.Duration) (time.Duration, uint64, error) {
	s, i, err := u.Get(key, "")
	if err != nil {
		return def, i, err
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def, i, err
	}
	return d, i, nil
}

// MustGetDuration is like MustGet but returns a duration.
func (u *EtcdUtil) MustGetDuration(key string) (time.Duration, uint64) {
	v, i, err := u.GetDuration(key, 0)
	if err != nil {
		panic(fmt.Sprintf("MustGetDuration(%v): %v\n", key, err))
	}
	return v, i
}

// GetJSON is like get but decodes the JSON to dst.
func (u *EtcdUtil) GetJSON(key string, dst interface{}) (uint64, error) {
	s, i, err := u.Get(key, "")
	if err != nil {
		return i, err
	}
	if err := json.Unmarshal([]byte(s), dst); err != nil {
		return i, err
	}
	return i, nil
}

// MustGetJSON is like MustGet but decodes the JSON to dst.
func (u *EtcdUtil) MustGetJSON(key string, dst interface{}) uint64 {
	i, err := u.GetJSON(key, dst)
	if err != nil {
		panic(fmt.Sprintf("MustGetJSON(%v): %v\n", key, err))
	}
	return i
}

// Close closes the etcd client and stops any watches.
func (u *EtcdUtil) Close() {
	close(u.s)
	u.c.Close()
}

// Watch watches for any changes to the given prefix+key. If recursive
// is true, any changes to that directory or any sub-directory are
// also watched. Whenever a change is received, the given function
// will be called for the changed key and the changed value.
//
// Using the value 0 for the waitIndex only returns new
// changes. Otherwise, you probably want it to be a value returned by
// one of the *Get* commands.
//
// Watching will continue to retry until Close() is called. Multiple
// Watch's may be started.
func (u *EtcdUtil) Watch(key string, waitIndex uint64, recursive bool, f func(key, value string)) {
	k := strings.Join([]string{u.p, key}, "/")
	c := make(chan *etcd.Response)

	// This is the goroutine that receives updates and calls f.
	go func() {
		for {
			select {
			case <-u.s:
				return
			case r := <-c:
				f(r.Node.Key, r.Node.Value)
			}
		}
	}()

	// This is the goroutine that watches until Close() is called.
	go func() {
		wait := startWait
		for {
			_, err := u.c.Watch(k, waitIndex, recursive, c, u.s)
			if err == etcd.ErrWatchStoppedByUser {
				return
			} else if err != nil {
				log.Printf("Watch(%v): %v - Retrying in %v\n", k, err, wait)
				select {
				case <-u.s:
					return
				case <-time.After(wait):
					if wait < maxWait {
						wait *= 2
					}
				}
			}
		}
	}()
}

// Walk gets all the values under prefix+key and calls f on each
// key/value pair. If sorted is true, calls to f will be done in order
// of the key. If f returns an error, walking is halted and the error
// returned.
func (u *EtcdUtil) Walk(key string, sorted bool, f func(key, value string) error) (uint64, error) {
	k := strings.Join([]string{u.p, key}, "/")
	r, err := u.c.Get(k, sorted, true)
	if err != nil {
		return 0, err
	}
	return r.EtcdIndex, u.walkHelper(r.Node, f)
}

func (u *EtcdUtil) walkHelper(n *etcd.Node, f func(key, value string) error) error {
	if n.Dir {
		for _, node := range n.Nodes {
			if err := u.walkHelper(node, f); err != nil {
				return err
			}
		}
		return nil
	}
	return f(n.Key, n.Value)
}
