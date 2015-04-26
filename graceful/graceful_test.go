// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package graceful

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	// This is the buffer we'll write the response to.
	b := &bytes.Buffer{}

	// We won't send responses until this is done (close has been called).
	waitToRespond := sync.WaitGroup{}
	waitToRespond.Add(1)

	// We won't check the responses until this is done (all requests
	// completed).
	waitForResponses := sync.WaitGroup{}

	// We won't close the server until this is done (all the requests
	// have been made).
	queued := sync.WaitGroup{}
	s := NewServer(&http.Server{
		Addr: ":38765",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We don't want to close until all of them have been queued.
			queued.Done()
			// We want to wait to respond until we've queued everything up.
			waitToRespond.Wait()
			w.Write([]byte("1"))
		}),
	})

	// Start listening.
	go s.ListenAndServe()

	// We need to wait for the server to be setup and running.
	for {
		if s.l == nil {
			time.Sleep(1 * time.Millisecond)
		} else {
			break
		}
	}
	for x := 0; x < 5; x++ {
		// We need to wait for another.
		waitForResponses.Add(1)
		// We've queued up another, so don't close until the server is in
		// the handler func.
		queued.Add(1)
		go func(y int) {
			// When we finish, we mark this work complete.
			defer waitForResponses.Done()
			resp, err := http.Get("http://localhost:38765/")
			if err != nil {
				t.Errorf("Unexpected error on get %v: %v", y, err)
				return
			}
			b.ReadFrom(resp.Body)
			resp.Body.Close()
		}(x)
	}

	// Wait for everything to queue up before closing. Then signal that
	// the server is closed. Finally, wait for all the responses to come
	// back.
	queued.Wait()
	s.Close()
	// Ensure that we can't connect again  before we send the others through.
	_, err := http.Get("http://localhost:38765/")
	if err == nil {
		t.Errorf("didn't get connection error when trying to connect after close.")
	}

	waitToRespond.Done()
	waitForResponses.Wait()

	// Check the results.
	if r := b.String(); r != "11111" {
		t.Errorf("failed to get all the responses: 11111 %v", r)
	}
}
