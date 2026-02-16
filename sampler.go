package lbpool

import "math"

const (
	base             = 100000
	minReleaseFactor = 0.00001
	maxReleaseFactor = 0.99999
	epsilon          = 1e-9
)

type sampler struct {
	lookup [base]bool
}

func newSampler(releaseFactor float64) *sampler {
	if releaseFactor < minReleaseFactor || math.IsNaN(releaseFactor) || math.IsInf(releaseFactor, 0) {
		releaseFactor = 0 // all requests will pass
	}
	if releaseFactor > maxReleaseFactor {
		releaseFactor = 1 // all requests will drop
	}
	s := &sampler{}
	if releaseFactor == 0 {
		return s
	}
	threshold := int(releaseFactor*base + epsilon)
	if threshold > base {
		threshold = base
	}
	// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm implementation
	var e int
	for i := 0; i < base; i++ {
		e += threshold
		if e >= base {
			e -= base
			s.lookup[i] = true
		}
	}
	return s
}

func (s *sampler) shouldDrop(i uint64) bool {
	return s.lookup[i%base]
}
