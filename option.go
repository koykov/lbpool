package lbpool

type Option func(p *pool)

func WithNewFn(fn func() any) Option {
	return func(p *pool) {
		p.newfn = fn
	}
}

func WithShards(shardsCount uint) Option {
	return func(p *pool) {
		p.shards = shardsCount
	}
}

func WithSampler(sampler Sampler) Option {
	return func(p *pool) {
		p.sampler = sampler
	}
}

func WithMetricsWriter(mw MetricsWriter) Option {
	return func(p *pool) {
		p.mw = mw
	}
}
