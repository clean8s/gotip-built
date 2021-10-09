package main

import (
	"time"
	"strings"
	"fmt"
)

type ProgressBar struct {
	currentSize, estimatedSize int
	initTime                   time.Time
	stop bool
}

func NewProgressBar(estimatedN int64) *ProgressBar {
	p := new(ProgressBar)
	p.initTime = time.Now()
	p.stop = false
	p.currentSize = 0
	p.estimatedSize = int(estimatedN)
	return p
}

func (bar *ProgressBar) ShouldRedraw() bool {
	if bar.currentSize >= bar.estimatedSize {
		return false
	}
	if bar.stop {
		return false
	}
	return true
}

func (bar *ProgressBar) Draw() float64 {
	dt := time.Now().Sub(bar.initTime)
	progress := float64(bar.currentSize) / float64(bar.estimatedSize)
	{
		dots := "."
		n := bar.currentSize * 20 / bar.estimatedSize
		dots += strings.Repeat(".", n)
		dots += strings.Repeat(" ", 20-n)
		K := fmt.Sprintf("Downloading & extracting (%.1f"+"%%"+" in %ds)", progress*100, int(dt.Seconds()))
		fmt.Printf("%-40s |%s|\r", K, dots)
	}
	return progress
}

func (bar *ProgressBar) Increment(increase int64) {
	bar.currentSize += int(increase)
	if bar.currentSize > bar.estimatedSize {
		bar.currentSize = bar.estimatedSize
	}
}
