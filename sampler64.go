package lbpool

import "math"

const (
	baseShift = 17 // 2^17 = 131072
	base64    = 1 << baseShift
	wordShift = 6 // 2^6 = 64
	wordBits  = 1 << wordShift
	wordCount = base64 >> wordShift
)

type sampler64 struct {
	base   uint64
	lookup [wordCount]uint64 // ceil(base / 64)
	mask   [64]uint64
}

func newSampler64(base uint64, releaseFactor float64) sampler {
	if releaseFactor < minReleaseFactor || math.IsNaN(releaseFactor) || math.IsInf(releaseFactor, 0) {
		releaseFactor = 0 // all requests will pass
	}
	if releaseFactor > maxReleaseFactor {
		releaseFactor = 1 // all requests will drop
	}
	s := &sampler64{base: base}
	if releaseFactor == 0 {
		return s
	}
	threshold := uint64(releaseFactor*float64(s.base) + epsilon)
	if threshold > s.base {
		threshold = s.base
	}
	// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm implementation
	var e uint64
	threshold = uint64(releaseFactor*float64(s.base) + epsilon)
	if threshold > s.base {
		threshold = s.base
	}
	for i := uint64(0); i < s.base; i++ {
		e += threshold
		if e >= s.base {
			e -= s.base
			wordIdx := i >> wordShift          // i / 64
			bitIdx := uint(i & (wordBits - 1)) // i % 64
			s.lookup[wordIdx] |= 1 << bitIdx
		}
	}
	// calculate masks
	for i := uint64(0); i < 64; i++ {
		s.mask[i] = uint64(1) << i
	}
	return s
}

func (s *sampler64) shouldDrop(i uint64) bool {
	j := i % wordCount
	v := (s.lookup[j/64] & (1 << (j % 64))) >> (j % 64)
	return v != 0
}
