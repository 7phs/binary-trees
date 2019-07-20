//
// Copyright (C) 2019 Aleksei Piianin <piyanin@gmail.com>
//

package main

import "log"

const (
	Naive    AllocatorType = 0
	Buffered AllocatorType = 1
)

var AllocatorSelected AllocatorType

type AllocatorType int

type Allocator interface {
	NewTree() *Tree
}

type AllocatorFactory func(uint32) Allocator

type AllocatorNaive struct{}

func NewAllocatorNaive(uint32) Allocator {
	return &AllocatorNaive{}
}

func (o *AllocatorNaive) NewTree() *Tree {
	return &Tree{}
}

type AllocatorBuffered struct {
	buffer []Tree
	index  int
}

func NewAllocatorBuffered(depth uint32) Allocator {
	return &AllocatorBuffered{
		buffer: make([]Tree, 1<<(depth+1)),
	}
}

func (o *AllocatorBuffered) NewTree() (tree *Tree) {
	tree = &o.buffer[o.index]
	o.index++
	return
}

func GetAllocator() AllocatorFactory {
	switch AllocatorSelected {
	case Naive:
		return NewAllocatorNaive

	case Buffered:
		return NewAllocatorBuffered

	default:
		log.Fatal("unsupported allocator: ", AllocatorSelected)
		return nil
	}
}
