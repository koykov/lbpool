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
	// Shards count.
	shards uint
	// Internal storage.
	strg storage
	// Counter of incoming items to store.
	c uint64
	// Sampler to equal should item be stored or not.
	// Dependents of releaseFactor.
	smpl    *sampler
	sampler Sampler
	// Metrics writer component.
	mw MetricsWriter
}

// NewPool inits new pool with given size.
// Deprecated: use New instead.
func NewPool(size uint, releaseFactor float64, options ...Option) Pool {
	return New(size, releaseFactor, options...)
}

// New inits new pool with given size.
func New(size uint, releaseFactor float64, options ...Option) Pool {
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
	x, s := p.strg.get()
	if x != nil {
		// Return existing object.
		p.mw.Hit(s)
		return x
	} else if p.newfn != nil {
		x = p.newfn()
		p.mw.New(s)
		return x
	}
	return nil
}

func (p *pool) Put(x Releaser) bool {
	// Check sampler first.
	if !p.sampler.Sample() {
		// Drop x on the floor.
		x.Release()
		p.mw.Leak("-1", "factor")
		return false
	}

	// Implement leaky buffer logic.
	ok, s := p.strg.put(x)
	if ok {
		p.mw.Store(s)
		return true
	} else {
		// Storage is full, release object manually and leak it.
		x.Release()
		p.mw.Leak(s, "overflow")
		return false
	}
}

func (p *pool) init() {
	p.smpl = newSampler(base, p.releaseFactor)

	// Check size and init the storage.
	if p.size == 0 {
		p.size = defaultPoolSize
	}
	if p.shards > 0 {
		p.strg = newSharded(p.size, p.shards)
	} else {
		p.strg = newSingle(p.size)
	}
	atomic.StoreUint64(&p.c, math.MaxUint64)

	if p.mw == nil {
		p.mw = dummyMetricsWriter{}
	}
}
