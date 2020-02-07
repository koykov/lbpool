package lbpool

const (
	poolDefSize = 64

	stateNil  = 0
	stateInit = 1
)

type Releaser interface {
	Release()
}

type Pool struct {
	c   chan interface{}
	l   int
	s   int
	New func() interface{}
}

func NewPool(limit int) *Pool {
	p := Pool{l: limit}
	p.initPool()
	return &p
}

func (p *Pool) initPool() {
	if p.l == 0 {
		p.l = poolDefSize
	}
	p.c = make(chan interface{}, p.l)
	p.s = stateInit
}

func (p *Pool) Get() interface{} {
	if p.s == stateNil {
		p.initPool()
	}

	var x interface{}
	select {
	case x = <-p.c:
		return x
	default:
		if p.New != nil {
			x = p.New()
			return x
		}
	}
	return nil
}

func (p *Pool) Put(x Releaser) {
	select {
	case p.c <- x:
		return
	default:
		x.Release()
	}
}
