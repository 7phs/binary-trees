# Boost performance "binary trees" of a benchmark game more than 3x times

[The Computer Language Benchmarks Game](https://benchmarksgame-team.pages.debian.net/benchmarksgame/index.html)
is an excellent synthetic test for compilers. There is different kind of algorithmic tasks implemented using popular
programming languages. 

[Go language](https://golang.org/) is one of my favorite language which I use for my pet projects and
production-ready projects.

I was surprised that Go is slower than several "native" languages (native against VM/JIT) on several tests,
for ex. [Binary trees](https://benchmarksgame-team.pages.debian.net/benchmarksgame/performance/binarytrees.html).

Binary trees is a memory-intensive using task. It approved by profiling.
I'v profiled the best implementation for Go - [Go #8](https://benchmarksgame-team.pages.debian.net/benchmarksgame/program/binarytrees-go-8.html).
A result - bottleneck is memory allocations.

## Buffered allocation

A source implementation has once allocation-intensive point:  

```go
func bottomUpTree(depth uint32) *Tree {
    tree := &Tree{}
    // .....
}
```

A little bit refactoring move that piece of code to an "allocator" to easy using a different kind of allocators in a project:

```go
func (o *AllocatorNaive) NewTree() *Tree {
	return &Tree{}
}

func NewTree(depth uint32, allocator Allocator) (tree *Tree) {
	tree = allocator.NewTree()
    // .....
}
```

This function was calling about 600 000 000 times of a "naive" allocator for 21 depth, and it was a place for 97.9% memory allocations.

The improvement is based on a task-specific prediction of allocation count.
A binary tree has __2^(depth+1)__ nodes including a root node. The buffered allocator is allocating this count of nodes once for a tree:

```go
func NewAllocatorBuffered(depth uint32) Allocator {
	return &AllocatorBuffered{
		buffer: make([]Tree, 1<<(depth+1)),
	}
}
```

A result - most calling allocation is that function for 5 580 000 times in ~107 times lesser than a naive allocator.

## Build and run 
 
You need Go 1.12+ with module support (GO111MODULE=on) to build a project:

```shell script
go build
```

A command supports options to choose allocator and depth of binary trees:

```shell script
./binary-trees -allocator 0 21
```

Supported allocators:
* 0 - naive (default)
* 1 - buffered

Another option is run without compilation stage:

```shell script
go run . -allocator 1 21
```

## Estimate CPU performance of allocators

An environment:

* Go 1.13 (tip)
* MacBook Pro 15 2012, Intel Core i7-3615QM (4 Cores), 16 Gb
* 5 times run for each allocator
* Estimate using a time command:
   ```shell script
   time ./binary-trees -allocator 1 21
   ```

The result:

 Allocator | Min (sec.) | Max (sec.) | Avg (sec.) 
---------- | ---------- | ---------- | ----------
 Naive     |     15.452 |     15.520 |    15.4796
 Buffered  |      4.816 |      4.895 |     4.8432 
 **Diff**  |            |            | **3.196x**
 
## A conclusion

* Profile your project to understand a performance bottleneck.
* Using a task-specific memory allocations strategy might help to boost the performance of data processing tasks in memory allocation cases.

## P.S.

Go standard library include [sync.Pool](https://golang.org/pkg/sync/#Pool).
It is [your best friend](https://github.com/valyala/fasthttp#fasthttp-best-practices) in general cases, as the author of the [fasthttp](https://github.com/valyala/fasthttp) said.   
