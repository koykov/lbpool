package lbpool

import (
	"math"
	"sync/atomic"
)

type storage interface {
	get() any
	put(any) bool
}

type single struct {
	ch chan any
}

func newSingle(size uint) storage {
	return &single{ch: make(chan any, size)}
}

func (s *single) get() any {
	select {
	case x := <-s.ch:
		return x
	default:
		return nil
	}
}

func (s *single) put(x any) bool {
	select {
	case s.ch <- x:
		return true
	default:
		return false
	}
}

type sharded struct {
	buf  []chan any
	r, w uint64
}

func newSharded(size uint, shards uint) storage {
	ssize := int(math.Ceil(float64(size) / float64(shards)))
	s := &sharded{
		buf: make([]chan any, shards),
		r:   math.MaxUint64,
		w:   math.MaxUint64,
	}
	for i := uint(0); i < shards; i++ {
		s.buf[i] = make(chan any, ssize)
	}
	return s
}

func (s *sharded) get() any {
	c := s.buf[atomic.AddUint64(&s.r, 1)%uint64(len(s.buf))]
	select {
	case x := <-c:
		return x
	default:
		return nil
	}
}

func (s *sharded) put(x any) bool {
	c := s.buf[atomic.AddUint64(&s.w, 1)%uint64(len(s.buf))]
	select {
	case c <- x:
		return true
	default:
		return false
	}
}
