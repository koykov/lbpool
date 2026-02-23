package sampler

type RNG interface {
	Seed(int64)
	Float64() float64
}

type Random struct {
	rate float64
	rng  RNG
}

func (s *Random) Sample() bool {
	return s.rng.Float64() < s.rate
}
