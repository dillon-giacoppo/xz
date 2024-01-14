# xz

The xz package implements reading of xz format compressed data implemented as a
cgo shim over `liblzma`. It aims to reduce allocations and buffer copying to
limit overhead where possible and remain performant.

### Install

```sh
go get dill.foo/xz
```

##### liblzma Dependency

This module dynamically links to `liblzma` which must be installed on the system.

`pkg-config` is used to identify the compiler options but can be disabled with
build tag `nopkgconfig`.

###### Ubuntu/Debian

```sh
sudo apt-get install liblzma-dev
```

###### MacOS

```sh
brew install xz
```

### Example

```go
package main

import (
	"fmt"
	"io"
	"os"

	"dill.foo/xz"
)

func main() {
	xr := xz.NewReader(os.Stdin)
	defer xr.Close()

	_, err := io.Copy(os.Stdout, xr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```
