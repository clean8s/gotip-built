// +build !gofuzzbeta

package fuzz

import (
	originalTest "testing"
)


// F is a type passed to fuzz targets.
//
// A fuzz target may add seed corpus entries using F.Add or by storing files in
// the testdata/fuzz/<FuzzTargetName> directory. The fuzz target must then
// call F.Fuzz once to provide a fuzz function. See the testing package
// documentation for an example, and see the F.Fuzz and F.Add method
// documentation for details
type F struct {
	NoopStruct
}

// Cleanup registers a function to be called after the fuzz function has been
// called on all seed corpus entries, and after fuzzing completes (if enabled).
// Cleanup functions will be called in last added, first called order
func (f *F) Cleanup(fn func()) {

}

// Error is equivalent to Log followed by Fail.
func (f *F) Error(args ...interface{}) {

}

// Errorf is equivalent to Logf followed by Fail.
func (f *F) Errorf(format string, args ...interface{}) {

}

// Fail marks the function as having failed but continues execution.
func (f *F) Fail() {

}

// FailNow marks the function as having failed and stops its execution
// by calling runtime.Goexit (which then runs all deferred calls in the
// current goroutine).
// Execution will continue at the next test, benchmark, or fuzz function.
// FailNow must be called from the goroutine running the
// fuzz target, not from other goroutines
// created during the test. Calling FailNow does not stop
// those other goroutines.
func (f *F) FailNow() {

}

// Fatal is equivalent to Log followed by FailNow.
func (f *F) Fatal(args ...interface{}) {

}

// Fatalf is equivalent to Logf followed by FailNow.
func (f *F) Fatalf(format string, args ...interface{}) {

}

// Helper marks the calling function as a test helper function.
// When printing file and line information, that function will be skipped.
// Helper may be called simultaneously from multiple goroutines.
func (f *F) Helper() {

}

// Setenv calls os.Setenv(key, value) and uses Cleanup to restore the
// environment variable to its original value after the test.
//
// When fuzzing is enabled, the fuzzing engine spawns worker processes running
// the test binary. Each worker process inherits the environment of the parent
// process, including environment variables set with F.Setenv.
func (f *F) Setenv(key, value string) {

}

// Skip is equivalent to Log followed by SkipNow.
func (f *F) Skip(args ...interface{}) {

}

// SkipNow marks the test as having been skipped and stops its execution
// by calling runtime.Goexit.
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Execution will continue at the next test or benchmark. See also FailNow.
// SkipNow must be called from the goroutine running the test, not from
// other goroutines created during the test. Calling SkipNow does not stop
// those other goroutines.
func (f *F) SkipNow() {

}

// Skipf is equivalent to Logf followed by SkipNow.
func (f *F) Skipf(format string, args ...interface{}) {

}

// TempDir returns a temporary directory for the test to use.
// The directory is automatically removed by Cleanup when the test and
// all its subtests complete.
// Each subsequent call to t.TempDir returns a unique directory;
// if the directory creation fails, TempDir terminates the test by calling Fatal.
func (f *F) TempDir() string {
	return ""
}

// Add will add the arguments to the seed corpus for the fuzz target. This will
// be a no-op if called after or within the Fuzz function. The args must match
// or be convertible to those in the Fuzz function.
func (f *F) Add(args ...interface{}) {

}

// A NoopStruct is used to simulate Go features that aren't available in stable branches.
type NoopStruct interface{}

type T = originalTest.T
type B = originalTest.B
type M = originalTest.M

type TB = originalTest.TB
type PB = originalTest.PB
type InternalTest = originalTest.InternalTest
type Cover = originalTest.Cover
type CoverBlock = originalTest.CoverBlock
type BenchmarkResult = originalTest.BenchmarkResult
type InternalExample = originalTest.InternalExample
