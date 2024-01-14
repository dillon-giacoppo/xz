// Copyright 2024 Dillon Giacoppo
// SPDX-License-Identifier: MIT

package xz

import (
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReader(t *testing.T) {
	const base64Input = "/Td6WFoAAATm1rRGAgAhARYAAAB0L+WjAQAMSGVsbG8KV29ybGQhCgAAAADvLogRnT+WygABJQ1xGcS2H7bzfQEAAAAABFla"
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64Input))
	xr := NewReader(r)
	if err := iotest.TestReader(xr, []byte("Hello\nWorld!\n")); err != nil {
		t.Fatal(err)
	}
}

func TestReader_Read(t *testing.T) {
	tests := []struct {
		name, base64Input, want string
		wantErr                 bool
		srcReader, outReader    func(io.Reader) io.Reader
	}{
		{
			name:        "behaves with DataErrReader",
			base64Input: "/Td6WFoAAATm1rRGAgAhARYAAAB0L+WjAQAMSGVsbG8KV29ybGQhCgAAAADvLogRnT+WygABJQ1xGcS2H7bzfQEAAAAABFla",
			want:        "Hello\nWorld!\n",
			srcReader:   iotest.DataErrReader,
		},
		{
			name:        "behaves with OneByteReader",
			base64Input: "/Td6WFoAAATm1rRGAgAhARYAAAB0L+WjAQAMSGVsbG8KV29ybGQhCgAAAADvLogRnT+WygABJQ1xGcS2H7bzfQEAAAAABFla",
			want:        "Hello\nWorld!\n",
			srcReader:   iotest.OneByteReader,
		},
		{
			name:        "behaves with HalfReader",
			base64Input: "/Td6WFoAAATm1rRGAgAhARYAAAB0L+WjAQAMSGVsbG8KV29ybGQhCgAAAADvLogRnT+WygABJQ1xGcS2H7bzfQEAAAAABFla",
			want:        "Hello\nWorld!\n",
			srcReader:   iotest.HalfReader,
		},
		{
			name:        "behaves with ErrReader",
			base64Input: "/Td6WFoAAATm1rRGAgAhARYAAAB0L+WjAQAMSGVsbG8KV29ybGQhCgAAAADvLogRnT+WygABJQ1xGcS2H7bzfQEAAAAABFla",
			srcReader: func(r io.Reader) io.Reader {
				return iotest.ErrReader(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name:        "behaves with output to OneByteReader",
			base64Input: "/Td6WFoAAATm1rRGAgAhARYAAAB0L+WjAQAMSGVsbG8KV29ybGQhCgAAAADvLogRnT+WygABJQ1xGcS2H7bzfQEAAAAABFla",
			want:        "Hello\nWorld!\n",
			outReader:   iotest.OneByteReader,
		},
		// The remaining test cases ensures the reader is compliant with up
		// stream XZ-utils spec. Cases are from:
		// https://github.com/tukaani-project/xz/tree/fbb3ce541ef79cad1710e88a27a5babb5f6f8e5b/tests/files
		{
			// has one stream with no blocks.
			name:        "good-0-empty.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVo=",
		},
		{
			// has one stream with no blocks followed by four-byte stream padding.
			name:        "good-0pad-empty.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVoAAAAA",
		},
		{
			// has two zero-block streams concatenated without stream padding.
			name:        "good-0cat-empty.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVr9N3pYWgAAAWki3jYAAAAAHN9EIZBCmQ0BAAAAAAFZWg==",
		},
		{
			// has two zero-block streams concatenated with four-byte stream
			// padding between the streams
			name:        "good-0catpad-empty.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVoAAAAA/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVo=",
		},
		{
			// has one stream with one block with two uncompressed LZMA2 chunks
			// and no integrity check
			name:        "good-1-check-none.xz",
			base64Input: "/Td6WFoAAAD/EtlBAgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgAAASANNO2zywZynnoBAAAAAABZWg==",
			want:        "Hello\nWorld!\n",
		},
		{
			// has one stream with one block with two uncompressed LZMA2 chunks
			// and CRC32 check.
			name:        "good-1-check-crc32.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgBDo6IVAAEkDTAo36+QQpkNAQAAAAABWVo=",
			want:        "Hello\nWorld!\n",
		},
		{
			// is like good-1-check-crc32.xz but with CRC64.
			name:        "good-1-check-crc64.xz",
			base64Input: "/Td6WFoAAATm1rRGAgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgDvLogRnT+WygABKA08Z2oDH7bzfQEAAAAABFla",
			want:        "Hello\nWorld!\n",
		},
		{
			// is like good-1-check-crc32.xz but with SHA256.
			name:        "good-1-check-sha256.xz",
			base64Input: "/Td6WFoAAArh+wyhAgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgCOWTXn4TNozZaI/o9IoJVSk2dqAhViWCx+hI2v4T+wRgABQA2Thk6uGJtLmgEAAAAAClla",
			want:        "Hello\nWorld!\n",
		},
		{
			// has one stream with two blocks with one uncompressed LZMA2 chunk
			// in each block.
			name:        "good-2-lzma2.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAAAFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAhoGGwcAAAbc510+MA2LAgAAAAABWVo=",
			want:        "Hello\nWorld!\n",
		},
		{
			// has both Compressed Size and Uncompressed Size in the block
			// Header. This has also four extra bytes of Header padding.
			name:        "good-1-block_header-1.xz",
			base64Input: "/Td6WFoAAAFpIt42A8ARDSEBCAAAAAAAf9456wEADEhlbGxvCldvcmxkIQoAAAAAQ6OiFQABJQ1xGcS2kEKZDQEAAAAAAVla",
			want:        "Hello\nWorld!\n",
		},
		{
			// has known Compressed Size.
			name:        "good-1-block_header-2.xz",
			base64Input: "/Td6WFoAAAFpIt42AkARIQEIAAA6TIjhAQAMSGVsbG8KV29ybGQhCgAAAABDo6IVAAEhDXXcqNKQQpkNAQAAAAABWVo=",
			want:        "Hello\nWorld!\n",
		},
		{
			// has known Uncompressed Size.
			name:        "good-1-block_header-3.xz",
			base64Input: "/Td6WFoAAAFpIt42AoANIQEIAABREYFZAQAMSGVsbG8KV29ybGQhCgAAAABDo6IVAAEhDXXcqNKQQpkNAQAAAAABWVo=",
			want:        "Hello\nWorld!\n",
		},
		{
			// has two LZMA2 chunks, of which the second sets new properties.
			name:        "good-1-lzma2-1.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMT4ADiALZdACYbykZnWvJ3uH2G2EHbBTXNg6V8EqUF25C9LxTTcXKWqIp9hFZxjWoimKuePZCALcdeDBJS0z8HCHscpHfzE7gXwO6RgTmzh/D/ALNqUkHtLrDyZJekmp5joa4ZdA2p1Vts7rHgLNxh3Mudhs/h3Ap6gRRf0EDIfg2XRM61wvwsWQi/A4Dc10SOs9Qt3uUWIW5HgqwIWdjkZilh1dH6SWOQET4g0Kni1RSB2SPQj0OuRVU2aaoAwADlAK0LAIzxnUAr0H0dme7k3GN0ZEakoEpkZbL2TsHIaJ8nVK27pjQ8d+wPLhuOQiflaL9g9As68Jsx698/2K+lVZJGBVgiCY+oYAgLo+k+vLQW28ejosAW1RSnIugv6LTQdxfFi+Tyu2vW75qBNE4d3Ow25kRyvym1PAUxYGa6LAMP1kfGfYXUxV5OV3PDQWm+DYyctRWp59J4UUvVKdD5NRrFXfSMenDVXqgxV4DIpdjgAAAA+0dI2wABggPJAwAACwSO3j4wDYsCAAAAAAFZWg==",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \nconsequat. Duis aute irure dolor in reprehenderit \nin voluptate velit esse cillum dolore eu \nfugiat nulla pariatur. Excepteur sint occaecat cupidatat \nnon proident, sunt in culpa qui officia \ndeserunt mollit anim id est laborum. \n",
		},
		{
			// has two LZMA2 chunks, of which the second resets the state
			// without specifying new properties.
			name:        "good-1-lzma2-2.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMT4ADiALZdACYbykZnWvJ3uH2G2EHbBTXNg6V8EqUF25C9LxTTcXKWqIp9hFZxjWoimKuePZCALcdeDBJS0z8HCHscpHfzE7gXwO6RgTmzh/D/ALNqUkHtLrDyZJekmp5joa4ZdA2p1Vts7rHgLNxh3Mudhs/h3Ap6gRRf0EDIfg2XRM61wvwsWQi/A4Dc10SOs9Qt3uUWIW5HgqwIWdjkZilh1dH6SWOQET4g0Kni1RSB2SPQj0OuRVU2aaoAoADlAK8AjPGdQH2CTyRyFPGdhMtaMmyXakCDi/CvMcK0ZW+J/fvYi1RBghZUEtFN1YbFwFr6SWOREf7/9Y8UAoVheThKS09BY/iHLyzm4ukxj4sU06F+gehVAu8hMaJ7BcwfpGDngaqn2XiC5hiyqxyqGS/ChxTF2cs/0BimzSpLXajHXwFnKEws5MzVUp6TAn4QXfUDsZgvJu2Ge1Z/E3lYj0QQ2dkPluk7v7W42ivh1oHxyQAA+0dI2wABgwPJAwAArtfSFT4wDYsCAAAAAAFZWg==",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \nconsequat. Duis aute irure dolor in reprehenderit \nin voluptate velit esse cillum dolore eu \nfugiat nulla pariatur. Excepteur sint occaecat cupidatat \nnon proident, sunt in culpa qui officia \ndeserunt mollit anim id est laborum. \n",
		},
		{
			// has two LZMA2 chunks, of which the first is uncompressed and the
			// second is LZMA. The first chunk resets dictionary and the second
			// sets new properties.
			name:        "good-1-lzma2-3.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQA0TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2ljaW5nIArAAZMBJF0AMpsJbFTXLpVs+fc31RziRgKCdftJdo1zU7b9bdvK29lEC7EuvhO2uqji8+11VNxBIMy/NlsgmV0PIaEGo5Ytt5ec8Hv+4hKMLVHw23Z3faR705Xp+wXm9ZePYunbMLu0cD0WeAN3Oot61bj4Sicl9Y6qJBSmKShrL3PgoXG0e6SAUEDK79u0lf27wYyOYJfby38h7cAQcRp9y80J0Nn/bYDAZ30/xpTPW91REdHL1CDXK4ROqEW7QngaaEBfJF6JOjZ925gozPmD7DIGMUdHO2wc9GI0QLMou1Q23XoOHDYlOFgG+BWjzhjI/ZYeaSkDw70n8+eP23O0K084WCS/gxQ5fnPu/s/KvfMhaiiAyI5dgce8F9Ask7UIlboOkoJmrv+4AwD7R0jbAAH0AskDAABnw5U+PjANiwIAAAAAAVla",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \nconsequat. Duis aute irure dolor in reprehenderit \nin voluptate velit esse cillum dolore eu \nfugiat nulla pariatur. Excepteur sint occaecat cupidatat \nnon proident, sunt in culpa qui officia \ndeserunt mollit anim id est laborum. \n",
		},
		{
			// has three LZMA2 chunks: First is LZMA, second is uncompressed
			// with dictionary reset, and third is LZMA with new properties but
			// without dictionary reset.
			name:        "good-1-lzma2-4.xz",
			base64Input: "/Td6WFoAAATm1rRGAgAhAQgAAADYDyMT4AC7AKFdACYbykZnWvJ3uH2G2EHbBTXNg6V8EqUF25C9LxTTcXKWqIp9hFZxjWoimKuePZCALcdeDBJS0z8HCHscpHfzE7gXwO6Rc8q8z+s0ZqxIm2nZkweuzlCvaAkvW4gfwgiiLFhFsP9iCevu22NPb+DzH88SN5iWTvbysvtur0QC4iLe1eY0lzmjRS+umS95aY/pN4lI/sx+6qkorcPm3LnaqhZ+AQAmbGFib3JpcyBuaXNpIHV0IGFsaXF1aXAgZXggZWEgY29tbW9kbyAKwADlAL1dADGbyhnFVOy2VOexfcRXnmyJrUptFtg8BZQQFpk4IaO5xYD//O7U1T/djNc9j3bsiKoyq2XUOO/3+Yq/9/ilVtdt1z+FC54/4kdoIggFNbhBcvnbvreOhr9DS44NQy9Bad9hDMToNwhK3sJ2FrhITp65U1AfM4PoKaBnyGY6fyISYvtH5Lz0UQ8ViEnYygsli17o2v04wM5Mcxv/0JvoTLcT+DeZ4tqcL7XquKWN6leCmyXK+/aICpvfQQNuAAAAsgdE6RczS4QAAasDyQMAAPVQLf6xxGf7AgAAAAAEWVo=",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \nconsequat. Duis aute irure dolor in reprehenderit \nin voluptate velit esse cillum dolore eu \nfugiat nulla pariatur. Excepteur sint occaecat cupidatat \nnon proident, sunt in culpa qui officia \ndeserunt mollit anim id est laborum. \n",
		},
		{
			// has an empty LZMA2 stream with only the end of payload marker.
			// XZ Utils 5.0.1 and older incorrectly see this file as corrupt.
			name:        "good-1-lzma2-5.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhARAAAACocI6GAAAAAAAAAAAAAREAO5Zfc5BCmQ0BAAAAAAFZWg==",
		},
		{
			// has three Delta filters and LZMA2.
			name:        "good-1-3delta-lzma2.xz",
			base64Input: "/Td6WFoAAATm1rRGBAMDAQADAQEDAQIhAQgAALwVZcYBAchMI7eE4glxT/q6ofdRYwisrvJYQg1m7qgBzWAuiFjXbts9JgAF8fuvNGcXwJ8/+fwNDgOk5q9psWKeR5dDwy9Ho6P1BFrAmz0BzFs6+rPCTJ1PV/27r1P/Bv/1p1FepJxjtLRi90egUG6v4wtSw6c3wFRJAbm0/ztfBK+7KMz/hGRxvjA/1VswqWF/pidTtb8AUz37urNeu/mBSbt0qaFO/bymTPG/VGbvpK1RIOMP7gwCpGM7/6jHVgKv3bFQwWf3S++0WkcGt1+jTarjF2W7qDAGtVJgp/TxFxX5Qa23OhW46p9mx1HRYRntCLz/W3Hxb3pnjgWmVZpx/pyiBF1g+6e28k5RvgfqUMKnSPse+O4R/Qae6bVmdJ4sVL+3VOIRCbZWMAmp0P4sXgyqWZZnBam7OLBGYA+srjfATGWuiFy/vELhe8E1SvW+oxZiNAKrtVsDA5/sf4bRZt88F+wKuEo8FLpflzgKwbxP8BGuNlEKt5pMMfD8p+e4WMT5OrX8p65aFgeo4JZfuGmlnVW2+wdLtJoHbkvoUxad/rG6UvK/751ewlboXfsEoltT/beqW7E2VgvBV4tRuwUKSVT5jRfNuUHdvAQ0AAAAALIHROkXM0uEAAHpA8kDAACS+728scRn+wIAAAAABFla",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \nconsequat. Duis aute irure dolor in reprehenderit \nin voluptate velit esse cillum dolore eu \nfugiat nulla pariatur. Excepteur sint occaecat cupidatat \nnon proident, sunt in culpa qui officia \ndeserunt mollit anim id est laborum. \n",
		},
		{
			// has one stream with no blocks followed by five-byte stream
			// padding. stream padding must be a multiple of four bytes, thus
			// this file is corrupt
			name:        "bad-0pad-empty.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVoAAAAAAA==",
			wantErr:     true,
		},
		{
			// has two zero-block streams concatenated with five-byte stream
			// padding between the streams.
			name:        "bad-0catpad-empty.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVoAAAAAAP03elhaAAABaSLeNgAAAAAc30QhkEKZDQEAAAAAAVla",
			wantErr:     true,
		},
		{
			// is good-0-empty.xz concatenated with an empty LZMA_Alone file.
			name:        "bad-0cat-alone.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVr9N3pYWQAAAWki3jYAAAAAHN9EIZBCmQ0BAAAAAAFZWg==",
			wantErr:     true,
		},
		{
			// is good-0cat-empty.xz but with one byte wrong in the Header
			// Magic Bytes field of the second stream. liblzma gives
			// LZMA_DATA_ERROR for this. (LZMA_FORMAT_ERROR is used only if the
			// first stream of a file has invalid Header Magic Bytes.)
			name:        "bad-0cat-header_magic.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVpdAAABAP//////////AIP/+///wAAAAA==",
			wantErr:     true,
		},
		{
			// is good-0-empty.xz but with one byte wrong in the Header Magic
			// Bytes field. liblzma gives LZMA_FORMAT_ERROR for this.
			name:        "bad-0-header_magic.xz",
			base64Input: "/Td6WFkAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVo=",
			wantErr:     true,
		},
		{
			// is good-0-empty.xz but with one byte wrong in the Footer Magic
			// Bytes field. liblzma gives LZMA_DATA_ERROR for this.
			name:        "bad-0-footer_magic.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWVg=",
			wantErr:     true,
		},
		{
			// is good-0-empty.xz without the last byte of the file.
			name:        "bad-0-empty-truncated.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCGQQpkNAQAAAAABWQ==",
			wantErr:     true,
		},
		{
			// has no blocks but Index claims that there is one block.
			name:        "bad-0-nonempty_index.xz",
			base64Input: "/Td6WFoAAAFpIt42AAEAACu1hiCQQpkNAQAAAAABWVo=",
			wantErr:     true,
		},
		{
			// has wrong Backward Size in stream Footer.
			name:        "bad-0-backward_size.xz",
			base64Input: "/Td6WFoAAAFpIt42AAAAABzfRCE1kcXGAAAAAAABWVo=",
			wantErr:     true,
		},
		{
			// has different stream Flags in stream Header and stream Footer.
			name:        "bad-1- stream_flags-1.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgBDo6IVAAEkDTAo368qE5CUAQAAAAACWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has wrong CRC32 in stream Header.
			name:        "bad-1- stream_flags-2.xz",
			base64Input: "/Td6WFoAAAFpIt52AgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgBDo6IVAAEkDTAo36+QQpkNAQAAAAABWVo=",
			wantErr:     true,
		},
		{
			// has wrong CRC32 in stream Footer.
			name:        "bad-1- stream_flags-3.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgBDo6IVAAEkDTAo36+QQpgNAQAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has two-byte variable-length integer in the Uncompressed Size
			// field in block Header while one-byte would be enough for that
			// value. It's important that the file gets rejected due to too big
			// integer encoding instead of due to Uncompressed Size not matching
			// the value stored in the block Header. That is, the decoder must
			// not try to decode the Compressed Data field.
			name:        "bad-1-vli-1.xz",
			base64Input: "/Td6WFoAAAFpIt42A4CNACEBCAAAAAAAoEipFwEADEhlbGxvCldvcmxkIQoAAAAAQ6OiFQABJQ1xGcS2kEKZDQEAAAAAAVla",
			wantErr:     true,
		},
		{
			// as ten-byte variable-length integer as Uncompressed Size in block
			// Header. It's important that the file gets rejected due to too big
			// integer encoding instead of due to Uncompressed Size not matching
			// the value stored in the block Header. That is, the decoder must
			// not try to decode the Compressed Data field.
			name:        "bad-1-vli-2.xz",
			base64Input: "/Td6WFoAAAFpIt42BICNgICAgICAgIABIQEIANJk8FwBAAxIZWxsbwpXb3JsZCEKAAAAAEOjohUAASkNfVZxGpBCmQ0BAAAAAAFZWg==",
			wantErr:     true,
		},
		{
			// has block Header that ends in the middle of the Filter Flags field.
			name:        "bad-1-block_header-1.xz",
			base64Input: "/Td6WFoAAAFpIt42AQAhAQydYGIBAAVIZWxsbwoCAAZXb3JsZCEKAEOjohUAASQNMCjfr5BCmQ0BAAAAAAFZWg==",
			wantErr:     true,
		},
		{
			// has block Header that has Compressed Size and Uncompressed Size
			// but no List of Filter Flags field.
			name:        "bad-1-block_header-2.xz",
			base64Input: "/Td6WFoAAAFpIt42AcAEDYCXihIBAAVIZWxsbwoCAAZXb3JsZCEKAEOjohUAASQNMCjfr5BCmQ0BAAAAAAFZWg==",
			wantErr:     true,
		},
		{
			// has wrong CRC32 in block Header.
			name:        "bad-1-block_header-3.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMzAQAFSGVsbG8KAgAGV29ybGQhCgBDo6IVAAEkDTAo36+QQpkNAQAAAAABWVo=",
			wantErr:     true,
		},
		{
			// has too big Compressed Size in block Header (2^63 - 1 bytes while
			// maximum is a little less, because the whole block must stay
			// smaller than 2^63). It's important that the file gets rejected
			// due to invalid Compressed Size value; the decoder must not try
			// decoding the Compressed Data field.
			name:        "bad-1-block_header-4.xz",
			base64Input: "/Td6WFoAAAFpIt42BED//////////38hAQgAAGPiOnABAAxIZWxsbwpXb3JsZCEKAAAAAEOjohUAASkNfVZxGpBCmQ0BAAAAAAFZWg==",
			wantErr:     true,
		},
		{
			// has zero as Compressed Size in block Header.
			name:        "bad-1-block_header-5.xz",
			base64Input: "/Td6WFoAAAFpIt42A8AADSEBCAAAAAAAqTRVIwEADEhlbGxvCldvcmxkIQoAAAAAQ6OiFQABJQ1xGcS2kEKZDQEAAAAAAVla",
			wantErr:     true,
		},
		{
			// has corrupt block Header which may crash xz -lvv in XZ Utils 5.0.3 and earlier.
			name:        "bad-1-block_header-6.xz",
			base64Input: "/Td6WFoAAAFpIt42AMARDSEBCAAAAAAAf9456wEADEhlbGxvCldvcmxkIQoAAAAAQ6OiFQABJQ1xGcS2kEKZDQEAAAAAAVla",
			wantErr:     true,
		},
		{
			// has wrong Unpadded Sizes in Index.
			name:        "bad-2-index-1.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAAAFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAhsGGgcAAMZoBy4+MA2LAgAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has wrong Uncompressed Sizes in Index.
			name:        "bad-2-index-2.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAAAFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAhoNGwAAAJL7eC8+MA2LAgAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has non-null byte in Index padding.
			name:        "bad-2-index-3.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAAAFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAhoGGwcAAZDs4Co+MA2LAgAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// wrong CRC32 in Index.
			name:        "bad-2-index-4.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAAAFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAhoGGwcAAAbc51w+MA2LAgAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has zero as Unpadded Size. It is important that the file gets
			// rejected specifically due to Unpadded Size having an invalid value.
			name:        "bad-2-index-5.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAAAFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAjUGAAcAAHu7BSw+MA2LAgAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has Index whose Uncompressed Size fields have huge values whose
			// sum exceeds the maximum allowed size of 2^63 - 1 bytes. In this
			// file the sum is exactly 2^64. lzma_index_append() in
			// liblzma <= 5.2.6 lacks the integer overflow check for the
			// uncompressed size and thus doesn't catch the error when decoding
			// the Index field in this file. This makes "xz -l" not detect the
			// error and will display 0 as the uncompressed size. Note that
			// regular decompression isn't affected by this bug because it uses
			// lzma_index_hash_append() instead.
			name:        "bad-3-index-uncomp-overflow.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQwAAACPmEGcAQAFSGVsbG8KAAAAFjWWMQIAIQEMAAAAj5hBnAEABFdvcmxkAAAAAEc+tvsCACEBDAAAAI+YQZwBAAEhCgAAAALuky0AAxr//////////38Z//////////9/FgIyic40KHKcEAYAAAAAAVla",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has non-null byte in the padding of the Compressed Data field of
			// the first block.
			name:        "bad-2-compressed_data_padding.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAAABFjWWMQIAIQEIAAAA2A8jEwEABldvcmxkIQoAAN3RylMAAhoGGwcAAAbc510+MA2LAgAAAAABWVo=",
			want:        "Hello\n",
			wantErr:     true,
		},
		{
			// has wrong Check (CRC32).
			name:        "bad-1-check-crc32.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgBDo6IUAAEkDTAo36+QQpkNAQAAAAABWVo=",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has Compressed Size and Uncompressed Size in block Header but
			// wrong Check (CRC32) in the actual data. This file differs by one
			// byte from good-1-block_header-1.xz: the last byte of the Check
			// field is wrong. This file is useful for testing error detection
			// in the threaded decoder when a worker thread is configured to
			// pass base64Input one byte at a time to the block decoder.
			name:        "bad-1-check-crc32-2.xz",
			base64Input: "/Td6WFoAAAFpIt42A8ARDSEBCAAAAAAAf9456wEADEhlbGxvCldvcmxkIQoAAAAAQ6Oi/wABJQ1xGcS2kEKZDQEAAAAAAVla",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has wrong Check (CRC64).
			name:        "bad-1-check-crc64.xz",
			base64Input: "/Td6WFoAAATm1rRGAgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgDvLogRnT+WywABKA08Z2oDH7bzfQEAAAAABFla",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has wrong Check (SHA-256).
			name:        "bad-1-check-sha256.xz",
			base64Input: "/Td6WFoAAArh+wyhAgAhAQgAAADYDyMTAQAFSGVsbG8KAgAGV29ybGQhCgCOWTXn4TNozZaI/o9IoJVSk2dqAhViWCx+hI2v4T+wRwABQA2Thk6uGJtLmgEAAAAAClla",
			want:        "Hello\nWorld!\n",
			wantErr:     true,
		},
		{
			// has LZMA2 stream whose first chunk (uncompressed) doesn't reset
			// the dictionary.
			name:        "bad-1-lzma2-1.xz",
			base64Input: "/Td6WFoAAAD/EtlBAgAhAQgAAADYDyMTAgAFSGVsbG8KAgAGV29ybGQhCgAAASANNO2zywZynnoBAAAAAABZWg==",
			wantErr:     true,
		},
		{
			// has two LZMA2 chunks, of which the second chunk indicates
			// dictionary reset, but the LZMA compressed data tries to repeat
			// data from the previous chunk.
			name:        "bad-1-lzma2-2.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMT4ADiALZdACYbykZnWvJ3uH2G2EHbBTXNg6V8EqUF25C9LxTTcXKWqIp9hFZxjWoimKuePZCALcdeDBJS0z8HCHscpHfzE7gXwO6RgTmzh/D/ALNqUkHtLrDyZJekmp5joa4ZdA2p1Vts7rHgLNxh3Mudhs/h3Ap6gRRf0EDIfg2XRM61wvwsWQi/A4Dc10SOs9Qt3uUWIW5HgqwIWdjkZilh1dH6SWOQET4g0Kni1RSB2SPQj0OuRVU2aaoA4ADlAK0LAIzxnUAr0H0dme7k3GN0ZEakoEpkZbL2TsHIaJ8nVK27pjQ8d+wPLhuOQiflaL9g9As68Jsx698/2K+lVZJGBVgiCY+oYAgLo+k+vLQW28ejosAW1RSnIugv6LTQdxfFi+Tyu2vW75qBNE4d3Ow25kRyvym1PAUxYGa6LAMP1kfGfYXUxV5OV3PDQWm+DYyctRWp59J4UUvVKdD5NRrFXfSMenDVXqgxV4DIpdjgAAAA+0dI2wABggPJAwAACwSO3j4wDYsCAAAAAAFZWg==",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \n",
			wantErr:     true,
		},
		{
			// sets new invalid properties (lc=8, lp=0, pb=0) in the middle of
			// block.
			name:        "bad-1-lzma2-3.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMT4ADiALZdACYbykZnWvJ3uH2G2EHbBTXNg6V8EqUF25C9LxTTcXKWqIp9hFZxjWoimKuePZCALcdeDBJS0z8HCHscpHfzE7gXwO6RgTmzh/D/ALNqUkHtLrDyZJekmp5joa4ZdA2p1Vts7rHgLNxh3Mudhs/h3Ap6gRRf0EDIfg2XRM61wvwsWQi/A4Dc10SOs9Qt3uUWIW5HgqwIWdjkZilh1dH6SWOQET4g0Kni1RSB2SPQj0OuRVU2aaoAwADlAK0IAIzxnUAr0H0dme7k3GN0ZEakoEpkZbL2TsHIaJ8nVK27pjQ8d+wPLhuOQiflaL9g9As68Jsx698/2K+lVZJGBVgiCY+oYAgLo+k+vLQW28ejosAW1RSnIugv6LTQdxfFi+Tyu2vW75qBNE4d3Ow25kRyvym1PAUxYGa6LAMP1kfGfYXUxV5OV3PDQWm+DYyctRWp59J4UUvVKdD5NRrFXfSMenDVXqgxV4DIpdjgAAAA+0dI2wABggPJAwAACwSO3j4wDYsCAAAAAAFZWg==",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \n",
			wantErr:     true,
		},
		{
			// has two LZMA2 chunks, of which the first is uncompressed and the
			// second is LZMA. The first chunk resets dictionary as it should,
			// but the second chunk tries to reset state without specifying
			// properties for LZMA.
			name:        "bad-1-lzma2-4.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQA0TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2ljaW5nIAqgAZMBJAAymwlsVNculWz59zfVHOJGAoJ1+0l2jXNTtv1t28rb2UQLsS6+E7a6qOLz7XVU3EEgzL82WyCZXQ8hoQajli23l5zwe/7iEowtUfDbdnd9pHvTlen7Beb1l49i6dswu7RwPRZ4A3c6i3rVuPhKJyX1jqokFKYpKGsvc+ChcbR7pIBQQMrv27SV/bvBjI5gl9vLfyHtwBBxGn3LzQnQ2f9tgMBnfT/GlM9b3VER0cvUINcrhE6oRbtCeBpoQF8kXok6Nn3bmCjM+YPsMgYxR0c7bBz0YjRAsyi7VDbdeg4cNiU4WAb4FaPOGMj9lh5pKQPDvSfz54/bc7QrTzhYJL+DFDl+c+7+z8q98yFqKIDIjl2Bx7wX0CyTtQiVug6Sgmau/7gDAAD7R0jbAAHzAskDAADf85AjPjANiwIAAAAAAVla",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \n",
			wantErr:     true,
		},
		{
			// is like bad-1-lzma2-4.xz but doesn't try to reset anything in the
			// header of the second chunk.
			name:        "bad-1-lzma2-5.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQA0TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2ljaW5nIAqAAZMBJAAymwlsVNculWz59zfVHOJGAoJ1+0l2jXNTtv1t28rb2UQLsS6+E7a6qOLz7XVU3EEgzL82WyCZXQ8hoQajli23l5zwe/7iEowtUfDbdnd9pHvTlen7Beb1l49i6dswu7RwPRZ4A3c6i3rVuPhKJyX1jqokFKYpKGsvc+ChcbR7pIBQQMrv27SV/bvBjI5gl9vLfyHtwBBxGn3LzQnQ2f9tgMBnfT/GlM9b3VER0cvUINcrhE6oRbtCeBpoQF8kXok6Nn3bmCjM+YPsMgYxR0c7bBz0YjRAsyi7VDbdeg4cNiU4WAb4FaPOGMj9lh5pKQPDvSfz54/bc7QrTzhYJL+DFDl+c+7+z8q98yFqKIDIjl2Bx7wX0CyTtQiVug6Sgmau/7gDAAD7R0jbAAHzAskDAADf85AjPjANiwIAAAAAAVla",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \n",
			wantErr:     true,
		},
		{
			// has reserved LZMA2 control byte value (0x03).
			name:        "bad-1-lzma2-6.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQAFSGVsbG8KAwAGV29ybGQhCgBDo6IVAAEkDTAo36+QQpkNAQAAAAABWVo=",
			want:        "Hello\n",
			wantErr:     true,
		},
		{
			// has EOPM at LZMA level.
			name:        "bad-1-lzma2-7.xz",
			base64Input: "/Td6WFoAAAFpIt42AgAhAQgAAADYDyMTAQA0TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2ljaW5nIAqAAZMBJAAymwlsVNculWz59zfVHOJGAoJ1+0l2jXNTtv1t28rb2UQLsS6+E7a6qOLz7XVU3EEgzL82WyCZXQ8hoQajli23l5zwe/7iEowtUfDbdnd9pHvTlen7Beb1l49i6dswu7RwPRZ4A3c6i3rVuPhKJyX1jqokFKYpKGsvc+ChcbR7pIBQQMrv27SV/bvBjI5gl9vLfyHtwBBxGn3LzQnQ2f9tgMBnfT/GlM9b3VER0cvUINcrhE6oRbtCeBpoQF8kXok6Nn3bmCjM+YPsMgYxR0c7bBz0YjRAsyi7VDbdeg4cNiU4WAb4FaPOGMj9lh5pKQPDvSfz54/bc7QrTzhYJL+DFDl+c+7+z8q98yFqKIDIjl2Bx7wX0CyTtQiVug6Sgmau/7gDAAD7R0jbAAHzAskDAADf85AjPjANiwIAAAAAAVla",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \n",
			wantErr:     true,
		},
		{
			// is like good-1-lzma2-4.xz but doesn't set new properties in the
			// third LZMA2 chunk.
			name:        "bad-1-lzma2-8.xz",
			base64Input: "/Td6WFoAAATm1rRGAgAhAQgAAADYDyMT4AC7AKFdACYbykZnWvJ3uH2G2EHbBTXNg6V8EqUF25C9LxTTcXKWqIp9hFZxjWoimKuePZCALcdeDBJS0z8HCHscpHfzE7gXwO6Rc8q8z+s0ZqxIm2nZkweuzlCvaAkvW4gfwgiiLFhFsP9iCevu22NPb+DzH88SN5iWTvbysvtur0QC4iLe1eY0lzmjRS+umS95aY/pN4lI/sx+6qkorcPm3LnaqhZ+AQAmbGFib3JpcyBuaXNpIHV0IGFsaXF1aXAgZXggZWEgY29tbW9kbyAKoADlAL0AMZvKGcVU7LZU57F9xFeebImtSm0W2DwFlBAWmTgho7nFgP/87tTVP92M1z2PduyIqjKrZdQ47/f5ir/3+KVW123XP4ULnj/iR2giCAU1uEFy+du+t46Gv0NLjg1DL0Fp32EMxOg3CErewnYWuEhOnrlTUB8zg+gpoGfIZjp/IhJi+0fkvPRRDxWISdjKCyWLXuja/TjAzkxzG//Qm+hMtxP4N5ni2pwvteq4pY3qV4KbJcr79ogKm99BA24AAAAAsgdE6RczS4QAAaoDyQMAAFCDcTWxxGf7AgAAAAAEWVo=",
			want:        "Lorem ipsum dolor sit amet, consectetur adipisicing \nelit, sed do eiusmod tempor incididunt ut \nlabore et dolore magna aliqua. Ut enim \nad minim veniam, quis nostrud exercitation ullamco \nlaboris nisi ut aliquip ex ea commodo \n",
			wantErr:     true,
		},
		// Flaky test on ubuntu-latest liblzma v5.2.5
		// {
		// 	// as LZMA2 stream that is truncated at the end of a LZMA2 chunk (no end marker).
		// 	// The uncompressed size of the partial LZMA2 stream exceeds the
		// 	// value stored in the block Header.
		// 	name:        "bad-1-lzma2-9.xz",
		// 	base64Input: "/Td6WFoAAAFpIt42A8AUDSEBCAAAAAAAOxUQDQEADEhlbGxvCldvcmxkIQoC//94Q6OiFQABKA08Z2oDkEKZDQEAAAAAAVla",
		// 	want:        "Hello\nWorld!\n",
		// 	wantErr:     true,
		// },
		// Flaky test on ubuntu-latest liblzma v5.2.5
		// {
		// 	// has LZMA2 stream that, from point of view of a LZMA2 decoder,
		// 	// extends past the end of block (and even the end of the file).
		// 	// Uncompressed Size in block Header is bigger than the invalid
		// 	// LZMA2 stream may produce (even if a decoder reads until the end
		// 	// of the file). The Check type is None to nullify certain simple
		// 	// size-based sanity checks in a block decoder.
		// 	name:        "bad-1-lzma2-10.xz",
		// 	base64Input: "/Td6WFoAAAFpIt42A8AUDSEBCAAAAAAAOxUQDQEADEhlbGxvCldvcmxkIQoC//94Q6OiFQABKA08Z2oDkEKZDQEAAAAAAVla",
		// 	want:        "Hello\nWorld!\n",
		// 	wantErr:     true,
		// },
		{
			// has LZMA2 stream that lacks the end of payload marker. When
			// Compressed Size bytes have been decoded, Uncompressed Size bytes
			// of output will have been produced but the LZMA2 decoder doesn't
			// indicate end of stream.
			name:        "bad-1-lzma2-11.xz",
			base64Input: "/Td6WFoAAAD/EtlBA8AQDSEBDAAAAAAAV/dqnwEADEhlbGxvIFdvcmxkIQoAASANNO2zywZynnoBAAAAAABZWg==",
			want:        "Hello World!\n",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(tt.base64Input))
				if tt.srcReader != nil {
					r = tt.srcReader(r)
				}
				var xr io.Reader = NewReader(r)
				if tt.outReader != nil {
					xr = tt.outReader(xr)
				}
				got, err := io.ReadAll(xr)
				if (err != nil) != tt.wantErr {
					t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if string(got) != tt.want {
					t.Errorf("Read() got = '%v', want %v", string(got), tt.want)
				}
			},
		)
	}
}
