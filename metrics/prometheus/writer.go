package prometheus

import "github.com/prometheus/client_golang/prometheus"

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
	promHit.WithLabelValues(w.name).Inc()
	promSize.WithLabelValues(w.name).Dec()
}

func (w *writer) New() {
	promNew.WithLabelValues(w.name).Inc()
}

func (w *writer) Store() {
	promStore.WithLabelValues(w.name).Inc()
	promSize.WithLabelValues(w.name).Inc()
}

func (w *writer) Leak(reason string) {
	promStore.WithLabelValues(w.name, reason).Inc()
}

var (
	promSize                              *prometheus.GaugeVec
	promHit, promNew, promStore, promLeak *prometheus.CounterVec
)

func init() {
	promSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lbpool_size",
		Help: "Pool size.",
	}, []string{"pool"})
	promHit = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_hit",
		Help: "How many items found in the pool.",
	}, []string{"pool"})
	promHit = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_new",
		Help: "How many items makes due to not found in the pool.",
	}, []string{"pool"})
	promStore = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_store",
		Help: "How many items returned back in the pool.",
	}, []string{"pool"})
	promLeak = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lbpool_leak",
		Help: "How many items dropped to the floor.",
	}, []string{"pool", "reason"})
	prometheus.MustRegister(promSize, promHit, promNew, promStore, promLeak)
}
