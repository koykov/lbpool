package lbpool

import (
	"math"
	"sync/atomic"
)

const (
	// Default size of the pool.
	defaultPoolSize = 64
)

// A Pool is a set of temporary objects.
// Object must implement release logic.
type Pool interface {
	// Get selects an arbitrary item from the pool, removes it from the pool, and returns it to the caller.
	Get() any
	// Put adds x to the pool.
	Put(x Releaser) bool
}

// Releaser is the interface that wraps the basic Release method.
type Releaser interface {
	Release()
}

type pool struct {
	// Maximum size of the pool.
	size uint
	// Release factor (RF) value and internal counter.
	// RF is a value that indicates how big part of items should be released even if pool may store them.
	// This feature need for gradual refresh of pool data and avoid to bloating objects stored in the pool.
	// RF should be in range [0.00001, 0.99999]. Note, that RF value around or equal 1.0 is senseless since in that case
	// poll will store only small piece of the data or event drop all data.
	// Usually RF <= 0.05 is enough.
	releaseFactor float64
	// Function to make new object if pool didn't deliver existing.
	newfn func() any
	// Internal storage and status flag.
	ch chan any
	// Counter of incoming items to store.
	c uint64
	// Sampler to equal should item be stored or not.
	// Dependents of releaseFactor.
	smpl *bsampler
	// Metrics writer component.
	mw MetricsWriter
}

// NewPool inits new pool with given size.
func NewPool(size uint, releaseFactor float64, options ...Option) Pool {
	p := &pool{
		size:          size,
		releaseFactor: releaseFactor,
	}
	for _, fn := range options {
		fn(p)
	}
	p.init()
	return p
}

func (p *pool) Get() any {
	var x any
	select {
	case x = <-p.ch:
		// Return existing object.
		p.mw.Hit()
		return x
	default:
		// Use New() function to make new object.
		if p.newfn != nil {
			x = p.newfn()
			p.mw.New()
			return x
		}
	}
	return nil
}

func (p *pool) Put(x Releaser) bool {
	// Check bsampler first.
	if p.smpl.shouldDrop(atomic.AddUint64(&p.c, 1)) {
		// Drop x on the floor.
		x.Release()
		p.mw.Leak("factor")
		return false
	}

	// Implement leaky buffer logic.
	select {
	case p.ch <- x:
		p.mw.Store()
		return true
	default:
		// Storage is full, release object manually and leak it.
		x.Release()
		p.mw.Leak("overflow")
	}
	return false
}

func (p *pool) init() {
	p.smpl = newSampler(base, p.releaseFactor)

	// Check size and init the storage.
	if p.size == 0 {
		p.size = defaultPoolSize
	}
	p.ch = make(chan any, p.size)
	atomic.StoreUint64(&p.c, math.MaxUint64)

	if p.mw == nil {
		p.mw = dummyMetricsWriter{}
	}
}
