package ltpool

import "sync"

const (
	poolDefSize = 64

	stateNil  = 0
	stateInit = 1
)

type PoolerItem interface {
	GetPoolId() int
	SetPoolId(id int)
}

type Pool struct {
	mux   sync.Mutex
	items []interface{}
	free  []int
	state uint8

	statGet int
	statPut int

	New func() PoolerItem
}

func NewPool() *Pool {
	p := &Pool{}
	p.initPool()
	return p
}

func (p *Pool) initPool() {
	p.items = make([]interface{}, 0, poolDefSize)
	p.free = make([]int, 0, poolDefSize)
	p.state = stateInit
}

func (p *Pool) Get() interface{} {
	p.mux.Lock()

	p.statGet++

	if p.state == stateNil {
		p.initPool()
	}
	if len(p.free) == 0 {
		if p.New != nil {
			x := p.New()
			x.SetPoolId(len(p.items))
			p.items = append(p.items, x)

			p.mux.Unlock()
			return &x
		}
	} else {
		fl := len(p.free)-1
		x := p.items[p.free[fl]]
		p.free = p.free[:fl]

		p.mux.Unlock()
		return x
	}

	p.mux.Unlock()
	return nil
}

func (p *Pool) Put(x PoolerItem) {
	p.mux.Lock()

	p.statPut++

	if len(p.items) == 0 {
		p.items = append(p.items, x)
		p.free = append(p.free, x.GetPoolId())
		p.mux.Unlock()
		return
	}
	if x.GetPoolId() == 0 {
		x.SetPoolId(len(p.items))
		p.items = append(p.items, x)
	}
	p.free = append(p.free, x.GetPoolId())

	p.mux.Unlock()
}
