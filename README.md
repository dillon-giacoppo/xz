# xz

The xz package implements reading of xz format compressed data implemented as a cgo shim over liblzma. It aims to reduce
allocations and buffer copying to limit overhead where possible and remain performant.

###### Install

```sh
go get dill.foo/xz
```
