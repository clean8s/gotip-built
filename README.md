### gofuzz

Go fuzzing is not ready in stable branch [go/tree/dev.fuzz](https://github.com/golang/go/tree/dev.fuzz),
so this is a mock implementation so that IDEs autocomplete it.

The mock is tagged as `!gofuzzbeta` so that you get the best of both worlds:
1. if you are actually running `gotip` you can fuzz but mock won't affect it
2. If you write fuzz code for stable Go in an IDE, it'll type check but not run it.
