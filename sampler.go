package lbpool

type Sampler interface {
	Sample() bool
}

type dummySampler struct{}

func (dummySampler) Sample() bool { return true }
