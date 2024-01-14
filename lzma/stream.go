// Copyright 2024 Dillon Giacoppo
// SPDX-License-Identifier: MIT

// Package lzma decompresses data with C-lzma library.
package lzma

/*
#cgo !nopkgconfig pkg-config: liblzma

#include <stdlib.h>
#include <lzma.h>

// Alias the LZMA_STREAM_INIT macro.
lzma_stream stream_init() {
	return (lzma_stream) LZMA_STREAM_INIT;
}

lzma_ret safe_lzma_code(lzma_stream *stream, lzma_action action) {
	lzma_ret ret = lzma_code(stream, action);

	// lzma_code advances the pointers which is not safe in go if it exceeds the
	// original slice bounds. Therefore, if we reach the end of stream->avail_*
	// assume we have gone off the end of the slice and therefore must null the
	// now invalid reference out.
	if (stream->avail_out == 0) {
      stream->next_out = NULL;
	}
    if (stream->avail_in == 0) {
      stream->next_in = NULL;
	}
    return ret;
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

type Stream struct {
	internal C.lzma_stream
	pinner   runtime.Pinner
}

// Return values used by several functions in liblzma.
type Return int

const (
	Ok               Return = iota // operation completed successfully
	StreamEnd                      // end of stream was reached.
	NoCheck                        // input stream has no integrity check
	UnsupportedCheck               // cannot calculate the integrity check
	GetCheck                       // integrity check type is now available
	MemError                       // cannot allocate memory
	MemLimitError                  // memory usage limit was reached
	FormatError                    // file format not recognized
	OptionsError                   // invalid or unsupported options
	DataError                      // data is corrupt
	BufError                       // no progress is possible
	ProgError                      // programming error
	SeekNeeded                     // request to change the input file position
)

// Action used by Stream.Code.
type Action int

const (
	Run         Action = iota // continue coding
	SyncFlush                 // make all the input available at output
	FullFlush                 // finish encoding of the current block
	Finish                    // finish the coding operation
	FullBarrier               // finish encoding of the current block
)

// A DecoderOpt can be passed in when initializing a decoder.
type DecoderOpt int32

const (
	TellNoCheck          DecoderOpt = 1 << iota // enables NoCheck
	TellUnsupportedCheck                        // enables UnsupportedCheck
	TellAnyCheck                                // enables GetCheck
	Concatenated                                // enables concatenated file support
	IgnoreCheck                                 // disables DataError for invalid integrity checks. Since liblzma 5.1.4beta
	FailFast                                    // enables eagerly returning errors in threaded decoding. Since liblzma 5.3.3alpha
)

// NewStreamDecoder initializes an .xz Stream decoder.
func NewStreamDecoder(memlimit uint64, flags ...DecoderOpt) (*Stream, error) {
	var decoderFlag int32
	for _, flag := range flags {
		decoderFlag |= int32(flag)
	}
	stream := Stream{
		internal: C.stream_init(),
	}
	ret := Return(
		C.lzma_stream_decoder(
			(*C.lzma_stream)(&stream.internal),
			C.uint64_t(memlimit),
			C.uint32_t(decoderFlag),
		),
	)
	if ret != Ok {
		return nil, fmt.Errorf("error init stream decoder code=%d", ret)
	}
	return &stream, nil
}

func (stream *Stream) SetNextIn(in []byte) {
	stream.internal.next_in = (*C.uint8_t)(unsafe.SliceData(in))
	stream.internal.avail_in = C.size_t(len(in))
}

func (stream *Stream) AvailableIn() int {
	return int(stream.internal.avail_in)
}

func (stream *Stream) SetNextOut(out []byte) {
	stream.internal.next_out = (*C.uint8_t)(unsafe.SliceData(out))
	stream.internal.avail_out = C.size_t(len(out))
}

func (stream *Stream) AvailableOut() int {
	return int(stream.internal.avail_out)
}

// Code encodes or decodes data based on how the Stream has been initialized,
// and it's current state as set by Stream.SetNextIn and Stream.SetNextOut.
func (stream *Stream) Code(action Action) Return {
	stream.pin()
	defer stream.pinner.Unpin()

	return Return(C.safe_lzma_code((*C.lzma_stream)(&stream.internal), C.lzma_action(action)))
}

// End frees memory allocated for the coder data structures used internally.
func (stream *Stream) End() {
	stream.pin()
	defer stream.pinner.Unpin()

	C.lzma_end((*C.lzma_stream)(&stream.internal))
}

func (stream *Stream) pin() {
	if stream.internal.next_in != nil {
		stream.pinner.Pin(stream.internal.next_in)
	}
	if stream.internal.next_out != nil {
		stream.pinner.Pin(stream.internal.next_out)
	}
}
