package lbpool

import (
	"sync"
)

const (
	poolDefSize = 64

	stateNil  = 0
	stateInit = 1
)

type Releaser interface {
	Release()
}

type Pool struct {
	Size  uint
	ch    chan interface{}
	state int
	once  sync.Once
	New   func() interface{}
}

func NewPool(size uint) *Pool {
	p := Pool{Size: size}
	p.initPool()
	return &p
}

func (p *Pool) initPool() {
	if p.Size == 0 {
		p.Size = poolDefSize
	}
	p.ch = make(chan interface{}, p.Size)
	p.state = stateInit
}

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

func (p *Pool) Put(x Releaser) {
	select {
	case p.ch <- x:
		return
	default:
		x.Release()
	}
}
