package lbpool

type Option func(p *pool)

func WithNewFn(fn func() any) Option {
	return func(p *pool) {
		p.newfn = fn
	}
}
