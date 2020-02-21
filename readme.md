# Leaky Buffer pool

A pool solution that implements [leaky buffer](https://golang.org/doc/effective_go.html#leaky_buffer) template.

It's slowly than vanilla pool but implements release logic in other hand. [sync/pool](https://golang.org/src/sync/pool.go) is a great pool solution
but it has a big inconvenience
> Any item stored in the Pool may be removed automatically at any time without
  notification. If the Pool holds the only reference when this happens, the
  item might be deallocated.
> - https://golang.org/src/sync/pool.go

This pool was made special for object like [cbyte](https://github.com/koykov/cbyte) that requires manual release.

Use it the same as vanilla pools.

## Benchmarks

```
BenchmarkPool-8                 20000000        81.8 ns/op       0 B/op       0 allocs/op
BenchmarkPoolParallel-8         100000000       19.6 ns/op       0 B/op       0 allocs/op
BenchmarkPoolNative-8           50000000        25.6 ns/op       0 B/op       0 allocs/op
BenchmarkPoolNativeParallel-8   200000000       5.54 ns/op       0 B/op       0 allocs/op
```

LB pool is 4-5 slowest that vanilla since it based on channels, whereas native is based on system pins.
