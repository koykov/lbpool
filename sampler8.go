package lbpool

import "math"

type sampler8 struct {
	base   uint64
	lookup [12500]uint8 // ceil(base / 8)
}

func newSampler8(base uint64, releaseFactor float64) *sampler8 {
	if releaseFactor < minReleaseFactor || math.IsNaN(releaseFactor) || math.IsInf(releaseFactor, 0) {
		releaseFactor = 0 // all requests will pass
	}
	if releaseFactor > maxReleaseFactor {
		releaseFactor = 1 // all requests will drop
	}
	s := &sampler8{base: base}
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
			s.lookup[i/8] |= 1 << uint8(i%8)
		}
	}
	return s
}

func (s *sampler8) shouldDrop(i uint64) bool {
	j := i % s.base
	v := (s.lookup[j/8] & (1 << (j % 8))) >> (j % 8)
	return v != 0
}
