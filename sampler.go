package lbpool

import "math"

const (
	base = 1e5

	minReleaseFactor = 0.00001
	maxReleaseFactor = 0.99999
	epsilon          = 1e-9
)

type sampler struct {
	base   uint64
	lookup [base]bool
}

func newSampler(base uint64, releaseFactor float64) *sampler {
	if releaseFactor < minReleaseFactor || math.IsNaN(releaseFactor) || math.IsInf(releaseFactor, 0) {
		releaseFactor = 0 // all requests will pass
	}
	if releaseFactor > maxReleaseFactor {
		releaseFactor = 1 // all requests will drop
	}
	s := &sampler{base: base}
	if releaseFactor == 0 {
		return s
	}
	threshold := uint64(releaseFactor*float64(s.base) + epsilon)
	if threshold > s.base {
		threshold = s.base
	}
	// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm implementation
	var e uint64
	for i := uint64(0); i < s.base; i++ {
		e += threshold
		if e >= s.base {
			e -= s.base
			s.lookup[i] = true
		}
	}
	return s
}

func (s *sampler) shouldDrop(i uint64) bool {
	return s.lookup[i%s.base]
}
