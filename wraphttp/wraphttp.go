// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package wraphttp provides functions that wrap some of the standard
// net/http interfaces.
package wraphttp

import (
	"io"
	"log"
	"net/http"
	"time"
)

// ResponseWriterStats stores information about data being written to
// an http.ResponseWriter. The information is not locked with any
// syncing, so you should only check the results when you are done
// writing.
type ResponseWriterStats struct {
	w            http.ResponseWriter
	ResponseCode int
	Total        int
}

// NewResponseWriterStats creates a new ResponseWriterStats that wraps
// the given http.ResponseWriter.
func NewResponseWriterStats(w http.ResponseWriter) *ResponseWriterStats {
	return &ResponseWriterStats{w: w, ResponseCode: 200}
}

// Header implements the ResponseWriter interface.
func (c *ResponseWriterStats) Header() http.Header {
	return c.w.Header()
}

// WriteHeader implements the ResponseWriter interface.
func (c *ResponseWriterStats) WriteHeader(h int) {
	c.ResponseCode = h
	c.w.WriteHeader(h)
}

// Write implements the ResponseWriter interface.
func (c *ResponseWriterStats) Write(data []byte) (int, error) {
	c.Total += len(data)
	return c.w.Write(data)
}

// RequestBodyStats stores information about data being read from an
// http.Request.Body. The information is not locked with any syncing,
// so you should only check the results when you are done reading.
type RequestBodyStats struct {
	r     io.ReadCloser
	Total int
}

// NewRequestBodyStats creates a new RequestBodyStats that wraps the
// given io.ReadCloser.
func NewRequestBodyStats(r io.ReadCloser) *RequestBodyStats {
	return &RequestBodyStats{r: r}
}

// Close implements the io.ReadCloser interface.
func (r *RequestBodyStats) Close() error {
	return r.r.Close()
}

// Read implements the io.ReadCloser interface.
func (r *RequestBodyStats) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	r.Total += n
	return n, err
}

// NewLogHandler wraps the given http.Handler. The default log.Logger
// in the log package is used for logging. It logs: remote address,
// HTTP protocol, HTTP method, the request URL, the response code, the
// start time, the duration, and the number of bytes received and the
// number of bytes sent.
func NewLogHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		out := NewResponseWriterStats(w)
		in := NewRequestBodyStats(r.Body)
		r.Body = in
		h.ServeHTTP(out, r)
		diff := time.Now().Sub(start)
		log.Printf("%v %v %v %v %v %v %v %v %v", r.RemoteAddr, r.Proto, r.Method, r.URL,
			out.ResponseCode, start, diff, in.Total, out.Total)
	})
}
