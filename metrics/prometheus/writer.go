package prometheus

import "github.com/prometheus/client_golang/prometheus"

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
	promHit.WithLabelValues(w.name, shard).Inc()
	promSize.WithLabelValues(w.name, shard).Dec()
}

func (w *writer) New(shard string) {
	promNew.WithLabelValues(w.name, shard).Inc()
}

func (w *writer) Store(shard string) {
	promStore.WithLabelValues(w.name, shard).Inc()
	promSize.WithLabelValues(w.name, shard).Inc()
}

func (w *writer) Leak(shard, reason string) {
	promStore.WithLabelValues(w.name, shard, reason).Inc()
}

var (
	promSize                              *prometheus.GaugeVec
	promHit, promNew, promStore, promLeak *prometheus.CounterVec
)

func init() {
	promSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lbpool_size",
		Help: "Pool size.",
	}, []string{"pool", "shard"})
	promHit = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_hit",
		Help: "How many items found in the pool.",
	}, []string{"pool", "shard"})
	promHit = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_new",
		Help: "How many items makes due to not found in the pool.",
	}, []string{"pool", "shard"})
	promStore = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_store",
		Help: "How many items returned back in the pool.",
	}, []string{"pool", "shard"})
	promLeak = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_leak",
		Help: "How many items dropped to the floor.",
	}, []string{"pool", "shard", "reason"})
	prometheus.MustRegister(promSize, promHit, promNew, promStore, promLeak)
}
