package lbpool

type MetricsWriter interface {
	Hit()
	New()
	Store()
	Leak(reason string)
}

type dummyMetricsWriter struct{}

func (dummyMetricsWriter) Hit()          {}
func (dummyMetricsWriter) New()          {}
func (dummyMetricsWriter) Store()        {}
func (dummyMetricsWriter) Leak(_ string) {}
