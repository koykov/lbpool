package victoria

import "github.com/koykov/vmchain"

type Writer interface {
	Hit(shard string)
	New(shard string)
	Store(shard string)
	Leak(shard, reason string)
}

type writer struct {
	name string
}

func NewWriter(name string) Writer {
	w := &writer{name: name}
	return w
}

func (w *writer) Hit(shard string) {
	vmchain.Counter("lbpool_hit").WithLabel("pool", w.name).WithLabel("shard", shard).Inc()
	vmchain.Gauge("lbpool_size", nil).WithLabel("pool", w.name).WithLabel("shard", shard).Dec()
}

func (w *writer) New(shard string) {
	vmchain.Counter("lbpool_new").WithLabel("pool", w.name).WithLabel("shard", shard).Inc()
}

func (w *writer) Store(shard string) {
	vmchain.Counter("lbpool_store").WithLabel("pool", w.name).WithLabel("shard", shard).Inc()
	vmchain.Gauge("lbpool_size", nil).WithLabel("pool", w.name).WithLabel("shard", shard).Inc()
}

func (w *writer) Leak(shard, reason string) {
	vmchain.Counter("lbpool_leak").WithLabel("pool", w.name).WithLabel("shard", shard).WithLabel("reason", reason).Inc()
}

var _ = NewWriter
