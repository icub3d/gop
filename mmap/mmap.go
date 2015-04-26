// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

// Package mmap provides a simplified interface to using mmap on
// linux/unix systems.
package mmap

import (
	"os"
	"reflect"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Mmap represents a single mapped file. It uses mmap to store the
// data. Sync() and Close() should be used to ensure data
// integrity. Since the actual byte array is available, no locking is
// done at this level.
type Mmap struct {
	// The underlying file handle. It will be cleaned up with Close().
	File *os.File

	// The byte array of the mmaped file.
	Buf []byte
}

// New maps a new file. If size > 0, then the file is increased to the
// given size if it is not already at least that size. The flags and
// perms are passed when opening the file and determine how the mmap
// will be opened. If private it true, then the map will be private.
func New(name string, perms os.FileMode, flags int, size int64, private bool) (*Mmap, error) {
	var err error
	m := &Mmap{}
	m.File, err = os.OpenFile(name, flags, os.FileMode(perms))
	if err != nil {
		return nil, err
	}
	fi, err := m.File.Stat()
	if err != nil {
		m.File.Close()
		return nil, err
	}
	len := fi.Size()
	if size > 0 {
		if fi.Size() < size {
			err = m.File.Truncate(size)
			if err != nil {
				m.File.Close()
				return nil, err
			}
			len = size
		}
	}

	mperms := 0
	if flags&os.O_RDONLY != 0 || flags&os.O_RDWR != 0 {
		mperms |= unix.PROT_READ
	}
	if flags&os.O_WRONLY != 0 || flags&os.O_RDWR != 0 {
		mperms |= unix.PROT_WRITE
	}

	t := unix.MAP_SHARED
	if private {
		t = unix.MAP_PRIVATE
	}

	m.Buf, err = unix.Mmap(int(m.File.Fd()), 0, int(len), mperms, t)
	if err != nil {
		m.File.Close()
		return nil, err
	}
	return m, nil
}

// Sync ensures that any unwritten changes to the buffer are written
// to disk. It will block until completed or an error occurs.
func (m *Mmap) Sync() error {
	sh := *(*reflect.SliceHeader)(unsafe.Pointer(&m.Buf))
	_, _, err := unix.Syscall(unix.SYS_MSYNC,
		sh.Data, uintptr(sh.Len), unix.MS_SYNC)
	if err != 0 {
		return err
	}
	return nil
}

// Close closes the associated mmap and file handles for this mmap. It
// should not be used after this.
func (m *Mmap) Close() error {
	mErr := unix.Munmap(m.Buf)
	cErr := m.File.Close()
	if mErr != nil {
		return mErr
	}
	return cErr
}
