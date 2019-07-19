//
// Copyright (C) 2019 Aleksei Piianin <piyanin@gmail.com>
//

package main

import "testing"

// A test is helper to easy profiling a code using Goland
func TestRunNaive(t *testing.T) {
	run(21, NewAllocatorNaive)
}

// A test is helper to easy profiling a code using Goland
func TestRunBuffered(t *testing.T) {
	run(21, NewAllocatorBuffered)
}
