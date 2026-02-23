package lbpool

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
	// Function to make new object if pool didn't deliver existing.
	newfn func() any
	// Shards count.
	shards uint
	// Internal storage.
	storage storage
	// Sampler to equal should item be stored or not.
	sampler Sampler
	// Metrics writer component.
	mw MetricsWriter
}

// NewPool inits new pool with given size.
// Deprecated: use New instead.
func NewPool(size uint, options ...Option) Pool {
	return New(size, options...)
}

// New inits new pool with given size.
func New(size uint, options ...Option) Pool {
	p := &pool{size: size}
	for _, fn := range options {
		fn(p)
	}
	p.init()
	return p
}

func (p *pool) Get() any {
	x, s := p.storage.get()
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
	ok, s := p.storage.put(x)
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
	if p.sampler == nil {
		p.sampler = dummySampler{}
	}
	// Check size and init the storage.
	if p.size == 0 {
		p.size = defaultPoolSize
	}
	if p.shards > 0 {
		p.storage = newSharded(p.size, p.shards)
	} else {
		p.storage = newSingle(p.size)
	}

	if p.mw == nil {
		p.mw = dummyMetricsWriter{}
	}
}
