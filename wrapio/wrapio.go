// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package wrapio implements wrappers for the io.Reader and io.Writer
// interfaces. These wrappers act as middlemen that allow you to do
// multiple things with a single stream of data. They are useful when
// requesting the data multiple times may be difficult or
// expensive. They also eliminate the need to track and maintain all
// of these items yourself and make functions like io.Copy and
// ioutil.ReadAll extremely useful.
package wrapio

import (
	"fmt"
	"hash"
	"io"
	"sync"
)

// Wrap implements the io.Closer, io.Reader, and io.Writer interface.
type wrap struct {
	handler func([]byte)
	r       io.Reader
	w       io.Writer
}

// Read implements the io.Reader interface.
func (w *wrap) Read(p []byte) (int, error) {
	n, err := w.r.Read(p)
	if n > 0 {
		w.handler(p[:n])
	}
	return n, err
}

// Write implements the io.Writer interface.
func (w *wrap) Write(p []byte) (int, error) {
	w.handler(p)
	return w.w.Write(p)
}

// NewFuncReader returns an io.Reader that wraps the given io.Reader
// with the given handler. Any Read() operations that read at least
// one byte will run through the handler before being returned. If
// either of the parameters are nil, nil is returned.
func NewFuncReader(handler func([]byte), r io.Reader) io.Reader {
	if handler == nil || r == nil {
		return nil
	}
	return &wrap{handler: handler, r: r}
}

// NewFuncWriter returns an io.Writer that wraps the given io.Writer
// with the given handler. Any Write() operations will run through the
// handler before being written. If either of the parameters are nil,
// nil is returned.
//
// Since the handler is called with all the data before the write, if
// an error occurs and not all of it is written, sending that data
// again will cause it to be sent to the handler again as well. This
// is a special case because most errors on write are fatal, but in
// cases where writing will continue, this must be taken into account.
func NewFuncWriter(handler func([]byte), w io.Writer) io.Writer {
	if handler == nil || w == nil {
		return nil
	}
	return &wrap{handler: handler, w: w}
}

// NewHashReader returns an io.Reader that wraps the given io.Reader
// with the given hash.Hash. Any Read() operations will also be
// written to the hash allowing you to simultaneously read something
// and get the hash of that thing. If either of the parameters are
// nil, nil is returned.
func NewHashReader(h hash.Hash, r io.Reader) io.Reader {
	if h == nil {
		return nil
	}
	return NewFuncReader(func(p []byte) {
		h.Write(p)
	}, r)
}

// NewHashWriter returns an io.Writer that wraps the given io.Writer
// with the given hash.Hash. Any Write() operations will also be
// written to the hash allowing you to simultaneously write something
// and get the hash of that thing. If either of the parameters are
// nil, nil is returned.
func NewHashWriter(h hash.Hash, w io.Writer) io.Writer {
	if h == nil {
		return nil
	}
	return NewFuncWriter(func(p []byte) {
		h.Write(p)
	}, w)
}

// Stats maintains the statistics about the I/O. It is updated with
// each read/write operation. If you are accessing the values, you
// should Lock() before accessing them and Unlock() after you are done
// to prevent possible race conditions.
type Stats struct {
	sync.Mutex
	Total   int     // The total number of bytes that have passed through.
	Average float64 // The average number of bytes read or written per call.
	Calls   int     // The number of calls made to Read or Write.
}

// String implements the fmt.Stringer interface.
func (s *Stats) String() string {
	return fmt.Sprintf("[Total: %d, Average: %f, Calls: %d]",
		s.Total, s.Average, s.Calls)
}

func (s *Stats) update(p []byte) {
	s.Lock()
	defer s.Unlock()
	s.Total += len(p)
	s.Calls++
	s.Average = float64(s.Total / s.Calls)
}

// NewStatsReader returns an io.Reader that wraps the given io.Reader
// with the returned statistical analyzer. Any Read() operations will
// be analyzed and the statistics updated. If either of the parameters
// are nil, nil is returned.
func NewStatsReader(r io.Reader) (*Stats, io.Reader) {
	s := &Stats{}
	return s, NewFuncReader(s.update, r)
}

// NewStatsWriter returns an io.Writer that wraps the given io.Writer
// with the returned statistical analyzer. Any Write() operations will
// be analyzed and the statistics updated. If either of the parameters
// are nil, nil is returned.
func NewStatsWriter(w io.Writer) (*Stats, io.Writer) {
	s := &Stats{}
	return s, NewFuncWriter(s.update, w)
}

type block struct {
	r    io.Reader
	w    io.Writer
	size int
	buf  []byte
	err  error // The non-nil error from the last Read().

}

// Read implements the io.Reader interface.
func (b *block) Read(p []byte) (int, error) {
	// If we've finished reading, we can quit.
	if b.err != nil && len(b.buf) == 0 {
		return 0, b.err
	}
	// We'll only fill p with full blocks.
	n := (len(p) / b.size) * b.size
	if n == 0 {
		return 0, nil
	}
	if b.err == nil {
		// Fill p temporarily and append it to our buffer.
		l, err := b.r.Read(p)
		b.err = err
		b.buf = append(b.buf, p[:l]...)
	}
	// If the size of p if bigger than what we have, only pull the
	// number of blocks that is in the buffer.
	if n > len(b.buf) {
		n = (len(b.buf) / b.size) * b.size
	}
	// If we've reached the end and we don't have a full block, make it
	// our last send.
	if b.err != nil && n == 0 {
		n = len(b.buf)
	}
	// Copy what we have to p.
	copy(p, b.buf[:n])
	copy(b.buf, b.buf[n:])
	b.buf = b.buf[:len(b.buf)-n]
	return n, nil
}

// Write implements the io.Writer interface.
func (b *block) Write(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	// We should first append p to our buffer.
	b.buf = append(b.buf, p...)
	// Write out any whole blocks.
	l := (len(b.buf) / b.size) * b.size
	if l > 0 {
		n, err := b.w.Write(b.buf[:l])
		// Move the unwritten portion to the beginning of the buffer and
		// reslice the buffer.
		copy(b.buf, b.buf[l:])
		b.buf = b.buf[:len(b.buf)-l]
		// In the error case, we want to report the actual written
		// information.
		if err != nil {
			b.err = err
			return n, err
		}
	}
	return len(p), nil
}

// Close implements the io.Closer interface.
func (b *block) Close() error {
	if b.err != nil {
		return b.err
	}
	// Write out any remaining data (which wouldn't have fit into a
	// block).
	if len(b.buf) > 0 {
		_, err := b.w.Write(b.buf)
		return err
	}
	return nil
}

// NewBlockReader returns a reader that sends data to the given reader
// in blocks that are a multiple of size. The one exception of this is
// the last Read() in which there may be an incomplete block. If p in
// Read(p) is not the length of a block, no data will be written to it
// (i.e it will return 0, nil). This may cause an infinite loop if you
// never give a slice larger than size.
func NewBlockReader(size int, r io.Reader) io.Reader {
	if r == nil || size < 1 {
		return nil
	}
	return &block{r: r, size: size}
}

// NewBlockWriter returns a writer that sends data to the given writer
// in blocks that are a multiple of size. Writes may be held if there
// is not enough data to Write() a complete block. To adhere to the
// io.Writer documentation though, the returned number of written
// bytes will always be the length of the given slice unless an error
// occurred in writing.
//
// Because it is impossible to tell when writing is completed, the
// returned writer is also a closer. The close operation should be
// called to flush out the remaining unwritten data that did not fit
// into a block size.
func NewBlockWriter(size int, w io.Writer) io.WriteCloser {
	if w == nil || size < 1 {
		return nil
	}
	return &block{w: w, size: size}
}

// Last implements the io.Closer, io.Reader, and io.Writer interface.
type last struct {
	handler func([]byte) []byte
	bufLen  int
	bufCap  int
	tmpCap  int
	tmpLen  int
	buf     []byte
	tmp     []byte
	err     error
	r       io.Reader
	w       io.Writer
}

// Read implements the io.Reader interface.
func (l *last) Read(p []byte) (int, error) {
	lp := len(p)
	// Check our error scenarios first.
	if l.err != nil && l.bufLen == 0 {
		// We've sent all of our buffer and have an error condition. We
		// are done.
		return 0, l.err
	} else if l.err != nil {
		// We have an error condition, but haven't sent all of our data.
		data := l.handler(l.buf[:l.bufLen])
		copy(p, data)
		n := lp
		if n > len(data) {
			n = len(data)
		}
		l.bufLen = 0
		return n, l.err
	}
	// On the first call, we'll have an empty buffer. Let's fill it.
	if l.buf == nil {
		l.buf = make([]byte, lp)
		l.bufCap = lp
		l.bufLen, l.err = l.r.Read(l.buf)
		if l.tmpLen == 0 || l.err != nil {
			// Our first read could be our last.
			return l.Read(p)
		}
	}
	// We should do a read at this point. If the read returns data, we
	// want to fill p, and copy the data to our buffer. If it doesn't
	// return data and an error, then we know our last read was it so we
	// should close out.
	if l.tmpCap < lp {
		l.tmp = make([]byte, lp)
		l.tmpCap = lp
	}
	l.tmpLen, l.err = l.r.Read(l.tmp)
	if l.tmpLen == 0 {
		return l.Read(p)
	}
	// Copy as much of our buffer as we can into p.
	n := lp
	if n > l.bufLen {
		n = l.bufLen
	}
	copy(p, l.buf[:n])
	// Resize our buffer if necessary and copy our temporary data to the
	// buffer.
	if l.bufCap < l.tmpCap {
		l.buf = make([]byte, l.tmpCap)
		l.bufCap = l.tmpCap
	}
	copy(l.buf, l.tmp[:l.tmpLen])
	l.bufLen = l.tmpLen
	return n, nil
}

// Write implements the io.Writer interface.
func (l *last) Write(p []byte) (int, error) {
	if l.err != nil {
		return 0, l.err
	}
	// Write out the current buffer if we have some.
	if l.bufLen > 0 {
		_, l.err = l.w.Write(l.buf[:l.bufLen])
	}
	// Resize the buffer if necessary.
	lp := len(p)
	if l.bufCap < lp {
		l.buf = make([]byte, lp)
		l.bufCap = lp
	}
	// Copy p into our buffer.
	copy(l.buf, p)
	l.bufLen = lp
	return lp, nil
}

// Close implements the io.Closer interface.
func (l *last) Close() error {
	if l.bufLen > 0 {
		_, l.err = l.w.Write(l.handler(l.buf[:l.bufLen]))
	}
	return l.err
}

// NewLastFuncReader returns an io.Reader that calls the given handler
// on the last Read() operation before passing it along. The last
// Read() operation is either the data returned with an error or if
// there is no data returned with the error, the data returned from
// the last call. If the slice passed to Read() is not consistent,
// data may be truncated.
func NewLastFuncReader(handler func([]byte) []byte, r io.Reader) io.Reader {
	if handler == nil || r == nil {
		return nil
	}
	return &last{handler: handler, r: r}
}

// NewLastFuncWriter returns an io.Writer that uses the given handler
// on the data from the very last Write() operation. It does this by
// holding onto the last Write()'s data without sending it. The first
// call to Write() won't send data to the given writer. Because it is
// impossible to tell when the last write is, the Close() function
// should be called after all the Write()s have been completed. This
// will cause the last write to be handed to the handler. The returned
// byte slice will be sent along.
func NewLastFuncWriter(handler func([]byte) []byte,
	w io.Writer) io.WriteCloser {
	if handler == nil || w == nil {
		return nil
	}
	return &last{handler: handler, w: w}
}
