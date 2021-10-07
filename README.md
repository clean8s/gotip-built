### gotip precompiled builds

Compiling Go with `gotip` can get slow, which makes it unusable for CI/unit tests that want to use it, like
projects that need fuzzing or Go generics.

This repo periodically builds go `master` via GitHub Actions into an archive you can easily download with `wget`: 

* Windows: https://github-releases.fikisipi.workers.dev/windows
* Linux: https://github-releases.fikisipi.workers.dev/linux
* Mac: https://github-releases.fikisipi.workers.dev/mac

Unzip the archive, set `$GOROOT` and you are ready to go.
