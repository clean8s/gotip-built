### Precompiled dev versions of golang

Using [gotip](https://pkg.go.dev/golang.org/dl/gotip) to install dev versions of go can delay
CI/unit tests for **4+ minutes**
as it goes through the whole toolchain compiling process.

This repo fixes that by storing precompiled daily gotip builds.


💾**Installation**:
```bash
go install github.com/clean8s/gotip-built/gotip@master
gotip download
```
Then you can use it as usual `go`: `gotip install`, `gotip mod tidy`, ...

👷 As of October, the most useful merged feature is [fuzz testing](https://go.dev/blog/fuzz-beta), and of course, [Go generics](https://github.com/golang/go/labels/generics).

---

**Verifying download hash**: The `download` command outputs
the SHA-256 hash of the go source tar (it's calculated on-the-go as blocks get downloaded & uncompressed).
You can check whether it matches the artifact generated by
the [GitHub Action](https://github.com/clean8s/gotip-built/actions/workflows/gotip-dw.yml) - as well as verify the build steps.

**Supported platforms**: darwin_amd64, windows_amd64, linux_amd64 
