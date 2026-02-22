package sampler

import (
	"math"
	"sync/atomic"
)

const (
	base = 1e5

	minSampleRate = 0.00001
	maxSampleRate = 0.99999
	epsilon       = 1e-9
)

// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm implementation
type bresenham struct {
	lookup [base]bool
	c      uint64
}

func NewBresenham(sampleRate float64) Sampler {
	if sampleRate < minSampleRate || math.IsNaN(sampleRate) || math.IsInf(sampleRate, 0) {
		sampleRate = 0 // all requests will drop
	}
	if sampleRate > maxSampleRate {
		sampleRate = 1 // all requests will pass
	}
	s := &bresenham{c: math.MaxUint64}
	if sampleRate == 0 {
		return s
	}
	threshold := uint64(sampleRate*base + epsilon)
	if threshold > base {
		threshold = base
	}
	var e uint64
	for i := uint64(0); i < base; i++ {
		e += threshold
		if e >= base {
			e -= base
			s.lookup[i] = true
		}
	}
	return s
}

func (s *bresenham) Sample() bool {
	return !s.lookup[atomic.AddUint64(&s.c, 1)%base]
}
