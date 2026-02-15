package lbpool

import (
	"sync/atomic"
)

const (
	// Default size of the pool.
	defaultPoolSize = 64

	// Default release factor.
	defaultReleaseFactor float32 = 0
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
	// RF should be in range [0.0, 1.0]. Note, that RF value around or equal 1.0 is senseless since in that case poll
	// will store only small piece of the data.
	// Usually RF <= 0.05 is enough.
	releaseFactor float32
	rfCounter     uint32
	// RF base allows you to specify the precision of release factor. For example, combination of:
	// * RF == 0.05
	// * RF base == 100
	// , means that 5% of items will be dropped on the floor.
	rfBase uint32
	// Function to make new object if pool didn't deliver existing.
	newfn func() any
	// Internal storage and status flag.
	ch chan any
}

// NewPool inits new pool with given size.
func NewPool(size uint, releaseFactor float32, options ...Option) Pool {
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
		return x
	default:
		// Use New() function to make new object.
		if p.newfn != nil {
			x = p.newfn()
			return x
		}
	}
	return nil
}

func (p *pool) Put(x Releaser) bool {
	// Check release factor first.
	if p.releaseFactor > 0 && p.rfBase > 0 {
		rfc := atomic.AddUint32(&p.rfCounter, 1)
		if rfc >= p.rfBase {
			// Release factor counter limit reached, reset it.
			atomic.StoreUint32(&p.rfCounter, 0)
		} else if float32(rfc)/float32(p.rfBase) <= p.releaseFactor {
			// Drop x on the floor.
			x.Release()
			return false
		}
	}

	// Implement leaky buffer logic.
	select {
	case p.ch <- x:
		return true
	default:
		// Storage is full, release object manually and leak it.
		x.Release()
	}
	return false
}

func (p *pool) init() {
	// Check bounds of release factor first.
	if p.releaseFactor < 0 {
		p.releaseFactor = defaultReleaseFactor
	}
	if p.releaseFactor > 1.0 {
		p.releaseFactor = 1.0
	}
	if p.rfBase == 0 {
		p.rfBase = 1
	}
	if p.releaseFactor > defaultReleaseFactor && p.releaseFactor < 1 {
		for float32(p.rfBase)*p.releaseFactor < 1 {
			p.rfBase *= 10
		}
	}

	// Check size and init the storage.
	if p.size == 0 {
		p.size = defaultPoolSize
	}
	p.ch = make(chan any, p.size)
}
