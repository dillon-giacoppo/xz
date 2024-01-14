// Copyright 2024 Dillon Giacoppo
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"os"

	"dill.foo/xz"
)

func main() {
	dec := xz.NewReader(os.Stdin)
	_, err := io.Copy(os.Stdout, dec)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
