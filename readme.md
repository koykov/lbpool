# Leaky Buffer Pool

An implementation of `Pool` based on the [leaky buffer](https://golang.org/doc/effective_go.html#leaky_buffer) pattern.
Actually, this pattern is more accurately called [leaky bucket](https://en.wikipedia.org/wiki/Leaky_bucket), but in Go terminology,
this specific name is used.

This solution has both advantages and disadvantages in comparison to native [sync.Pool](https://golang.org/src/sync/pool.go) implementation:
the native solution is faster, but it has a significant drawback:
> Any item stored in the Pool may be removed automatically at any time without
notification. If the Pool holds the only reference when this happens, the
item might be deallocated.
> - https://golang.org/src/sync/pool.go

If you need to store something in the pool that cannot be automatically cleaned up by the GC, you will encounter a memory leak.
Such objects include, for example, [cbyte](https://github.com/koykov/cbyte) arenas, which require manual release.

Outside of this feature, this pool works similarly to the native one.

## API

The `Get` method of the pool is full analogue to the native pool's method and has the same signature:
```go
func Get() any
```

The `Put` method can only accept objects that implement the `Releaser` interface:
```go
type Releaser interface {
    Release()
}

func Put(x Releaser) bool
```
It also returns a `bool` indicating whether the object leaked past the bucket or not.

The pool requires initialization via the `New` constructor:
```go
func New(size uint, releaseFactor float64, options ...Option) Pool
```
with arguments:
* `size` - The maximum size of the pool. Once this size is reached, all newly arriving elements will leak.
* `releaseFactor` - The proportion of incoming elements that will be forcibly leaked past the pool. This is needed for gradual renewal of elements in the pool.
* `options` - A list of optional functions that allow setting additional pool parameters.

Currently, three optional functions are available:
* `WithNewFn` - Sets the New function for constructing an object to return instead of a `nil` value.
* `WithShards` - The number of pool shards, needed to reduce lock contention on the internal channel's mutex.
* `WithMetricsWriter` - Connects a component for writing metrics to the pool.

## Usage

```go
import (
    "github.com/koykov/lbpool"
    "github.com/koykov/lbpool/metrics/victoria"
)

const (
    poolSize      = 100 // max pool size
    releaseFactor = .05 // part of items that will leak independent of leaky logic
)

type item struct{...}

func (item) Release() {...}

var pool = lbpool.New(poolSize, releaseFactor, lbpool.WithNewFn(func() any { return &item{} }),
lbpool.WithMetricsWriter(victoria.NewWriter("my_pool")))

itm := pool.Get().(*item)
...
pool.Put(itm)
```

## Performance

The pool is implemented over Go channels and works several times slower than the native pool:
```
BenchmarkPool-8                 	23870334	        50.07 ns/op	       0 B/op	       0 allocs/op
BenchmarkPoolParallel-8         	100000000	        11.62 ns/op	       0 B/op	       0 allocs/op
BenchmarkPoolNative-8           	100000000	        10.59 ns/op	       0 B/op	       0 allocs/op
BenchmarkPoolNativeParallel-8   	424538383	        2.920 ns/op	       0 B/op	       0 allocs/op
```
Consider this when choosing tools for your tasks.
