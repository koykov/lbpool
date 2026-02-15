package victoria

import "github.com/koykov/vmchain"

type Writer interface {
	Hit()
	New()
	Store()
	Leak(reason string)
}

type writer struct {
	name string
}

func NewWriter(name string) Writer {
	w := &writer{name: name}
	return w
}

func (w *writer) Hit() {
	vmchain.Counter("lbpool_hit").WithLabel("pool", w.name).Inc()
	vmchain.Gauge("lbpool_size", nil).WithLabel("pool", w.name).Dec()
}

func (w *writer) New() {
	vmchain.Counter("lbpool_new").WithLabel("pool", w.name).Inc()
}

func (w *writer) Store() {
	vmchain.Counter("lbpool_store").WithLabel("pool", w.name).Inc()
	vmchain.Gauge("lbpool_size", nil).WithLabel("pool", w.name).Inc()
}

func (w *writer) Leak(reason string) {
	vmchain.Counter("lbpool_leak").WithLabel("pool", w.name).WithLabel("reason", reason).Inc()
}

var _ = NewWriter
