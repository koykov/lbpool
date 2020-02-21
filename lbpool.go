package lbpool

import (
	"sync"
)

const (
	// Default size of the pool.
	poolDefSize = 64

	// Pool status code.
	stateNil  = 0
	stateInit = 1
)

// Releaser is the interface that wraps the basic Release method.
type Releaser interface {
	Release()
}

// A Pool is a set of temporary objects.
// Object must implement release logic.
type Pool struct {
	Size  uint
	ch    chan interface{}
	state int
	once  sync.Once
	New   func() interface{}
}

// Init new pool with given size.
func NewPool(size uint) *Pool {
	p := Pool{Size: size}
	p.initPool()
	return &p
}

// Prepare pool for work.
func (p *Pool) initPool() {
	if p.Size == 0 {
		p.Size = poolDefSize
	}
	p.ch = make(chan interface{}, p.Size)
	p.state = stateInit
}

// Get selects an arbitrary item from the Pool, removes it from the
// Pool, and returns it to the caller.
func (p *Pool) Get() interface{} {
	if p.state == stateNil {
		p.once.Do(func() { p.initPool() })
	}

	var x interface{}
	select {
	case x = <-p.ch:
		return x
	default:
		if p.New != nil {
			x = p.New()
			return x
		}
	}
	return nil
}

// Put adds x to the pool.
func (p *Pool) Put(x Releaser) {
	select {
	case p.ch <- x:
		return
	default:
		x.Release()
	}
}
