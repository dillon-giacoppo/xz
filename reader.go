// Copyright 2024 Dillon Giacoppo
// SPDX-License-Identifier: MIT

// Package xz decompresses data with C-lzma library.
package xz

import (
	"errors"
	"fmt"
	"io"
	"math"

	"dill.foo/xz/lzma"
)

const defaultBufferSize = 32 * 1024

type reader struct {
	src     io.Reader
	stream  *lzma.Stream
	buf     []byte
	action  lzma.Action
	lastErr error
}

// NewReader creates a XZ decoder reader from the given source.
func NewReader(src io.Reader) io.ReadCloser {
	stream, err := lzma.NewStreamDecoder(math.MaxUint64, lzma.Concatenated, lzma.TellUnsupportedCheck)
	return &reader{
		src:     src,
		stream:  stream,
		buf:     make([]byte, defaultBufferSize),
		action:  lzma.Run,
		lastErr: err,
	}
}

func (r *reader) Read(p []byte) (int, error) {
	if r.lastErr != nil || len(p) == 0 {
		return 0, r.lastErr
	}
	r.stream.SetNextOut(p)
	for {
		if r.stream.AvailableIn() == 0 {
			n, err := r.src.Read(r.buf)
			if err != nil && err != io.EOF {
				r.lastErr = err
				return 0, err
			}
			if err == io.EOF {
				r.action = lzma.Finish
			}
			r.stream.SetNextIn(r.buf[:n])
		}
		ret := r.stream.Code(r.action)
		written := len(p) - r.stream.AvailableOut()
		switch ret {
		case lzma.Ok:
			if r.stream.AvailableOut() == 0 {
				return written, nil
			}
		case lzma.StreamEnd:
			r.lastErr = io.EOF
			_ = r.stream.Close()
			return written, io.EOF
		default:
			r.lastErr = fmt.Errorf("lzma return error code=%d", ret)
			_ = r.stream.Close()
			return written, r.lastErr
		}
	}
}

// Close closes the reader. If the caller consumes the entire Reader until io.EOF
// (or other error) as is typical with methods such as io.ReadAll then the
// resources will have been freed from the terminal Read call and close will
// have no effect.
func (r *reader) Close() error {
	if r.lastErr == nil {
		r.lastErr = errors.New("reader is closed")
		_ = r.stream.Close()
	}
	return nil
}
