// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package wraphttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	_ = r.Body.Close()
	w.WriteHeader(400)
	h := w.Header()
	_, _ = io.WriteString(w, h.Get("Host"))
	_, _ = io.WriteString(w, " - echo: ")
	_, _ = w.Write(body)
})

func TestLogging(t *testing.T) {
	h := NewLogHandler(testHandler)

	// Setup out request and logging.
	data := bytes.NewBuffer([]byte("hello, server"))
	r, err := http.NewRequest("POST", "/", data)
	if err != nil {
		t.Fatalf("failed making request: %v", err)
	}
	rr := httptest.NewRecorder()
	ld := &bytes.Buffer{}
	log.SetOutput(ld)

	// Make the request and verify the values.
	h.ServeHTTP(rr, r)
	line := ld.String()
	parts := strings.Split(line, " ")
	if len(parts) != 14 {
		t.Fatalf("didn't get a full log line (14 parts): %v %v", len(parts), parts)
	}
	if parts[4] != "POST" {
		t.Errorf("Method was not POST: %v", parts[4])
	}
	if parts[5] != "/" {
		t.Errorf("URL was not '/': %v", parts[5])
	}
	if parts[6] != "400" {
		t.Errorf("code was not 400: %v", parts[6])
	}
	if parts[12] != "13" {
		t.Errorf("request size was not 13: %v", parts[12])
	}
	if parts[13] != "22\n" {
		t.Errorf("response was not 22: %v", parts[13])
	}
	if rr.Body.String() != " - echo: hello, server" {
		t.Errorf("response body was not ' - echo: hello, server': %v",
			rr.Body.String())
	}
}
