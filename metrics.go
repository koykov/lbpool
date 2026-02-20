package lbpool

type MetricsWriter interface {
	Hit(shard string)
	New(shard string)
	Store(shard string)
	Leak(shard, reason string)
}

type dummyMetricsWriter struct{}

func (dummyMetricsWriter) Hit(string)       {}
func (dummyMetricsWriter) New(string)       {}
func (dummyMetricsWriter) Store(string)     {}
func (dummyMetricsWriter) Leak(_, _ string) {}
