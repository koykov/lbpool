package lbpool

import (
	"sync"
	"sync/atomic"
)

const (
	// Default size of the pool.
	defaultPoolSize = 64

	// Default release factor
	defaultReleaseFactor float32 = 0
	// and precision.
	defaultReleaseFactorBase uint32 = 100

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
	// Maximum size of the pool.
	Size uint
	// Release factor (RF) value and internal counter.
	// RF is a value that indicates how big part of items should be released even if pool may store them.
	// This feature need for gradual refresh of pool data and avoid to bloating objects stored in the pool.
	// RF should be in range [0.0, 1.0]. Note, that RF value around or equal 1.0 is senseless since in that case poll
	// will store only small piece of the data.
	// Usually RF <= 0.05 is enough.
	ReleaseFactor float32
	rfCounter     uint32
	// RF base allows you to specify the precision of release factor. For example, combination of:
	// * RF == 0.05
	// * RF base == 100
	// , means that 5% of items will be drop on the floor.
	ReleaseFactorBase uint32
	// Function to make new object if pool didn't deliver existing.
	New func() interface{}
	// Internal storage and status flag.
	ch    chan interface{}
	state int
	// Once helper that guarantee only one init of the pool.
	once sync.Once
}

var (
	// Suppress go vet warnings.
	_ = NewPool
)

// Init new pool with given size.
func NewPool(size uint, releaseFactor float32) *Pool {
	p := Pool{
		Size:          size,
		ReleaseFactor: releaseFactor,
	}
	p.initPool()
	return &p
}

// Prepare pool for work.
func (p *Pool) initPool() {
	// Check bounds of release factor first.
	if p.ReleaseFactor < 0 {
		p.ReleaseFactor = defaultReleaseFactor
	}
	if p.ReleaseFactor > 1.0 {
		p.ReleaseFactor = 1.0
	}
	if p.ReleaseFactorBase == 0 {
		p.ReleaseFactorBase = defaultReleaseFactorBase
	}

	// Check size and init the storage.
	if p.Size == 0 {
		p.Size = defaultPoolSize
	}
	p.ch = make(chan interface{}, p.Size)
	p.state = stateInit
}

// Get selects an arbitrary item from the Pool, removes it from the
// Pool, and returns it to the caller.
func (p *Pool) Get() interface{} {
	// Implement once logic if pool isn't inited yet.
	if p.state == stateNil {
		p.once.Do(func() { p.initPool() })
	}

	var x interface{}
	select {
	case x = <-p.ch:
		// Return existing object.
		return x
	default:
		// Use New() function to make new object.
		if p.New != nil {
			x = p.New()
			return x
		}
	}
	return nil
}

// Put adds x to the pool.
func (p *Pool) Put(x Releaser) bool {
	// Check release factor first.
	if p.ReleaseFactor > 0 && p.ReleaseFactorBase > 0 {
		rfc := atomic.AddUint32(&p.rfCounter, 1)
		if rfc >= p.ReleaseFactorBase {
			// Release factor counter limit reached, reset it.
			atomic.StoreUint32(&p.rfCounter, 0)
		} else if float32(rfc)/float32(p.ReleaseFactorBase) <= p.ReleaseFactor {
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
