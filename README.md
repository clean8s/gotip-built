### gotip precompiled builds

Go fuzzing is still not released: [go/tree/dev.fuzz](https://github.com/golang/go/tree/dev.fuzz),
so this is:

1. An up-to-date build of the ~~dev.fuzz~~ `master` branch **[edit: dev.fuzz has been merged to master]**
2. A mock implementation so that IDEs autocomplete it.

The mock is tagged as `!gofuzzbeta` so that you get the best of both worlds:
1. if you are actually running `gotip` you can fuzz but mock won't affect it
2. If you write fuzz code for stable Go in an IDE, it'll type check but not run it.
---

Since compiling takes more > 4 minutes reusing a build is faster:

```yaml
- name: fuzz download
  run: |
    FUZZREPO="https://api.github.com/repos/clean8s/gotip-builds/releases/latest"
    GOTIP=$(curl -sL "$FUZZREPO" | jq -r '.assets[].browser_download_url')
    wget $GOTIP && tar xzf gotip-amd64-ubuntu-latest.tar.gz
```
 

To use fuzzing: `$GOROOT="$HOME/gotip" $GOROOT/bin/go test -fuzz .`

Example: Add `gotipmock.go` and start writing a fuzzer in any `*_test.go` by naming it `FuzzX(f *F)`:

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
