package lbpool

import (
	"math"
	"strconv"
	"sync/atomic"
)

type storage interface {
	get() (any, string)
	put(any) (bool, string)
}

type single struct {
	ch chan any
}

func newSingle(size uint) storage {
	return &single{ch: make(chan any, size)}
}

func (s *single) get() (any, string) {
	select {
	case x := <-s.ch:
		return x, "0"
	default:
		return nil, "0"
	}
}

func (s *single) put(x any) (bool, string) {
	select {
	case s.ch <- x:
		return true, "0"
	default:
		return false, "0"
	}
}

type sharded struct {
	buf     []shard
	l, r, w uint64
}

type shard struct {
	ch   chan any
	name string
}

func newSharded(size uint, shards uint) storage {
	ssize := uint64(math.Ceil(float64(size) / float64(shards)))
	s := &sharded{
		buf: make([]shard, shards),
		r:   math.MaxUint64,
		w:   math.MaxUint64,
		l:   ssize,
	}
	for i := uint(0); i < shards; i++ {
		s.buf[i].ch = make(chan any, ssize)
		s.buf[i].name = strconv.Itoa(int(i))
	}
	return s
}

func (s *sharded) get() (any, string) {
	_ = s.buf[s.l-1]
	ss := s.buf[atomic.AddUint64(&s.r, 1)%s.l]
	select {
	case x := <-ss.ch:
		return x, ss.name
	default:
		return nil, ss.name
	}
}

func (s *sharded) put(x any) (bool, string) {
	_ = s.buf[s.l-1]
	ss := s.buf[atomic.AddUint64(&s.w, 1)%s.l]
	select {
	case ss.ch <- x:
		return true, ss.name
	default:
		return false, ss.name
	}
}
