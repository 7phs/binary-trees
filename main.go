//
// The source code based on the Computer Language Benchmarks Game
// http://benchmarksgame.alioth.debian.org/
//
// Go adaptation of binary-trees Rust #4 program - https://benchmarksgame-team.pages.debian.net/benchmarksgame/program/binarytrees-go-8.html
// This release uses semaphores to match the number of workers with the CPU count
//
// contributed by Marcel Ibes
//
// This version extending the base code to use several memory allocators strategy.
//
// Copyright (C) 2019 Aleksei Piianin <piyanin@gmail.com>
//

package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"runtime"
	"sort"
	"strconv"

	"golang.org/x/sync/semaphore"
)

type Tree struct {
	Left  *Tree
	Right *Tree
}

func NewTree(depth uint32, allocator Allocator) (tree *Tree) {
	tree = allocator.NewTree()

	if depth > uint32(0) {
		tree.Right = NewTree(depth-1, allocator)
		tree.Left = NewTree(depth-1, allocator)
	}

	return
}

func (o *Tree) ItemCheck() uint32 {
	if o.Left != nil && o.Right != nil {
		return uint32(1) + o.Right.ItemCheck() + o.Left.ItemCheck()
	}

	return 1
}

type Message struct {
	Pos  uint32
	Text string
}

func inner(depth, iterations uint32, allocatorFabric AllocatorFactory) string {
	chk := uint32(0)
	for i := uint32(0); i < iterations; i++ {
		chk += NewTree(
			depth,
			allocatorFabric(depth),
		).ItemCheck()
	}
	return fmt.Sprintf("%d\t trees of depth %d\t check: %d",
		iterations, depth, chk)
}

const minDepth = uint32(4)

func main() {
	n := 0

	flag.IntVar((*int)(&AllocatorSelected), "allocator", 0, "allocator type: 0 - naive, 1 - buffered")

	flag.Parse()

	if flag.NArg() > 0 {
		n, _ = strconv.Atoi(flag.Arg(0))
	}

	run(uint32(n), GetAllocator())
}

func run(n uint32, allocatorFabric AllocatorFactory) {
	cpuCount := runtime.NumCPU()
	sem := semaphore.NewWeighted(int64(cpuCount))

	maxDepth := n
	if minDepth+2 > n {
		maxDepth = minDepth + 2
	}

	depth := maxDepth + 1

	messages := make(chan *Message, cpuCount)
	expected := uint32(2) // initialize with the 2 summary messages we're always outputting

	go func() {
		for halfDepth := minDepth / 2; halfDepth < maxDepth/2+1; halfDepth++ {
			depth := halfDepth * 2
			iterations := uint32(1 << (maxDepth - depth + minDepth))
			expected++

			func(d, i, pos uint32) {
				if err := sem.Acquire(context.TODO(), 1); err == nil {
					go func() {
						defer sem.Release(1)
						messages <- &Message{pos, inner(d, i, allocatorFabric)}
					}()
				} else {
					panic(err)
				}
			}(depth, iterations, expected)
		}

		if err := sem.Acquire(context.TODO(), 1); err == nil {
			go func() {
				defer sem.Release(1)

				messages <- &Message{0,
					fmt.Sprintf("stretch tree of depth %d\t check: %d",
						depth, NewTree(
							depth,
							allocatorFabric(depth),
						).ItemCheck())}
			}()
		} else {
			panic(err)
		}

		if err := sem.Acquire(context.TODO(), 1); err == nil {
			go func() {
				defer sem.Release(1)

				messages <- &Message{math.MaxUint32,
					fmt.Sprintf("long lived tree of depth %d\t check: %d",
						maxDepth, NewTree(
							maxDepth,
							allocatorFabric(maxDepth),
						).ItemCheck())}
			}()
		} else {
			panic(err)
		}
	}()

	sortedMsg := make([]*Message, 0, len(messages))
	for m := range messages {
		sortedMsg = append(sortedMsg, m)
		expected--
		if expected == 0 {
			close(messages)
		}
	}

	sort.Slice(sortedMsg, func(i, j int) bool { return sortedMsg[i].Pos < sortedMsg[j].Pos })

	for _, m := range sortedMsg {
		fmt.Println(m.Text)
	}

	// MEM STATISTICS
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)

	fmt.Printf("Alloc: %.3f/%.3f\n", float64(memStat.Alloc)/(1024*1024), float64(memStat.TotalAlloc)/(1024*1024))
}
