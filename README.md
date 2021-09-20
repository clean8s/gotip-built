### gofuzz

Go fuzzing is still not in stable Go branch: [go/tree/dev.fuzz](https://github.com/golang/go/tree/dev.fuzz),
so this is:

1. An up-to-date build of `dev.fuzz`
2. A mock implementation so that IDEs autocomplete it.

The mock is tagged as `!gofuzzbeta` so that you get the best of both worlds:
1. if you are actually running `gotip` you can fuzz but mock won't affect it
2. If you write fuzz code for stable Go in an IDE, it'll type check but not run it.

### Downloading compiled tips inside a GitHub Action

[![gotip compile](https://github.com/clean8s/gofuzz/actions/workflows/gotip-dw.yml/badge.svg)](https://github.com/clean8s/gofuzz/actions/workflows/gotip-dw.yml)

`gotip` is very slow so you can use this repo to obtain a precompiled build from the `dev.fuzz` branch for amd64.

```yaml
- name: fuzz download
  run: |
    GOTIP=$(curl -sL https://api.github.com/repos/clean8s/gofuzz/releases/latest | jq -r '.assets[].browser_download_url') \
    wget $GOTIP && tar xzf gofuzz_linux_amd64.tar.gz
- run: ./gotip/bin/go test -fuzz .
```

Add `gotipmock.go` and start writing a fuzzer in any `*_test.go` by naming it `FuzzX(f *F)`:

```go
func FuzzSomeFunction(f *F) {
	for i := 0; i < 1000; i++ {
		f.Add("corpus" + blah(i))
	}
  
	f.Fuzz(func(t *T, fuzzString string) {
		if yourFn(fuzzString).Output.Invalid {
			t.Skip()
		}
	})
}
```

`./gotip/bin/go test -fuzz .` will mutate `fuzzString` and try to find an input that crashes your function:
```sh
./gotip/bin/go test -fuzz .
fuzz: elapsed: 18s, execs: 424755 (23583/sec), workers: 4, interesting: 32
        --- FAIL: FuzzOne (0.00s)
            testing.go:1241: panic: Some panic
                goroutine 10134 [running]:
                runtime/debug.Stack()
```
