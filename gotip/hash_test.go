package main

import (
	"strings"
	"bufio"
	"fmt"
	"testing"
)

func TestSimpleCheck(t *testing.T) {
	S := strings.NewReader("Hello world")
	r := hashInterceptReader{
		sourceReader: S,
		sha256:       nil,
	}
	b := bufio.NewReader(&r)
	fmt.Println(b.ReadByte())
	if res := fmt.Sprintf("%x", r.sha256.Sum(nil)); res != "64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f3c" {
		t.Fatal("Hash reader failed with result:", res)
	}
}
