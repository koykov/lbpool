package lbpool

import (
	"math/rand"
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
	Size          uint
	ReleaseFactor float32
	ch            chan interface{}
	state         int
	once          sync.Once
	New           func() interface{}
}

var (
	// Suppress go vet warnings.
	_ = NewPool
)

// Init new pool with given size.
func NewPool(size uint) *Pool {
	p := Pool{Size: size}
	p.initPool()
	return &p
}

// Prepare pool for work.
func (p *Pool) initPool() {
	// Check bounds of release factor first.
	if p.ReleaseFactor < 0 {
		p.ReleaseFactor = 0
	}
	if p.ReleaseFactor > 1.0 {
		p.ReleaseFactor = 1.0
	}

	// Check size and init the storage.
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
func (p *Pool) Put(x Releaser) bool {
	// Check release factor first
	if p.ReleaseFactor > 0 {
		if rand.Float32() <= p.ReleaseFactor {
			// ... and release x.
			x.Release()
			return false
		}
	}

	// Implement leaky buffer logic.
	select {
	case p.ch <- x:
		return true
	default:
		x.Release()
	}
	return false
}
