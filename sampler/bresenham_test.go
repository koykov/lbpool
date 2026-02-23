package sampler

import (
	"fmt"
	"math"
	"testing"
)

func TestBresenham(t *testing.T) {
	t.Run("count drops", func(t *testing.T) {
		type testcase struct {
			name            string
			sampleRate      float64
			expectedSamples int
		}
		tests := []testcase{
			{
				name:            "zero",
				sampleRate:      0,
				expectedSamples: 0,
			},
			{
				name:            "one",
				sampleRate:      1,
				expectedSamples: base,
			},
			{
				name:            "min sample rate",
				sampleRate:      0.00001,
				expectedSamples: 1,
			},
			{
				name:            "max sample rate",
				sampleRate:      0.99999,
				expectedSamples: base - 1,
			},
			{
				name:            "one percent",
				sampleRate:      0.01,
				expectedSamples: 1000,
			},
			{
				name:            "one third",
				sampleRate:      1.0 / 3.0,
				expectedSamples: int(math.Floor(1.0/3.0*base + epsilon)),
			},
			{
				name:            "half",
				sampleRate:      0.5,
				expectedSamples: base / 2,
			},
			{
				name:            "two thirds",
				sampleRate:      2.0 / 3.0,
				expectedSamples: int(math.Floor(2.0/3.0*base + epsilon)),
			},
			{
				name:            "clamping below min",
				sampleRate:      0.000001,
				expectedSamples: 0,
			},
			{
				name:            "clamping above max",
				sampleRate:      0.999999,
				expectedSamples: base,
			},
			{
				name:            "NaN",
				sampleRate:      math.NaN(),
				expectedSamples: 0,
			},
			{
				name:            "+Inf",
				sampleRate:      math.Inf(1),
				expectedSamples: 0,
			},
			{
				name:            "-Inf",
				sampleRate:      math.Inf(-1),
				expectedSamples: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := newBresenham(tt.sampleRate)

				var samplesCount int
				for i := 0; i < base; i++ {
					if s.lookup[i] {
						samplesCount++
					}
				}

				if samplesCount != tt.expectedSamples {
					t.Errorf("expected %d samples, got %d for sampleRate=%v",
						tt.expectedSamples, samplesCount, tt.sampleRate)
				}

				if tt.sampleRate == 0 {
					for i := 0; i < base; i++ {
						if s.lookup[i] {
							t.Errorf("expected no samples for sampleRate=0, but found at index %d", i)
							break
						}
					}
				}

				if tt.sampleRate == 1 {
					for i := 0; i < base; i++ {
						if !s.lookup[i] {
							t.Errorf("expected all samples for sampleRate=1, but found false at index %d", i)
							break
						}
					}
				}
			})
		}
	})
	t.Run("distribution/uniform", func(t *testing.T) {
		type testcase struct {
			name       string
			sampleRate float64
			iterations uint64
		}
		tests := []testcase{
			{
				name:       "one percent uniform",
				sampleRate: 0.01,
				iterations: 100,
			},
			{
				name:       "one third uniform",
				sampleRate: 1.0 / 3.0,
				iterations: 100,
			},
			{
				name:       "half uniform",
				sampleRate: 0.5,
				iterations: 100,
			},
			{
				name:       "two thirds uniform",
				sampleRate: 2.0 / 3.0,
				iterations: 100,
			},
			{
				name:       "ninety nine percent",
				sampleRate: 0.99,
				iterations: 100,
			},
			{
				name:       "min sample rate",
				sampleRate: 0.00001,
				iterations: 10000,
			},
		}

		const tolerance = 0.01
		testfn := func(t *testing.T, tt testcase, s *bresenham) {
			totalRequests := tt.iterations * base
			totalDrops := 0
			for i := uint64(0); i < totalRequests; i++ {
				if !s.Sample() {
					totalDrops++
				}
			}

			expectedDrops := int(float64(totalRequests) * (1 - tt.sampleRate))

			deviation := math.Abs(float64(totalDrops-expectedDrops)) / float64(expectedDrops)
			if deviation > tolerance {
				t.Errorf("distribution deviation too high: got %d drops, expected %d (deviation %.2f%%)",
					totalDrops, expectedDrops, deviation*100)
			}
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testfn(t, tt, newBresenham(tt.sampleRate))
			})
		}
	})
	t.Run("distribution/Bresenham", func(t *testing.T) {
		thresholds := []int{1, 2, 3, 10, 100, 1000, 50000}

		for _, threshold := range thresholds {
			t.Run(fmt.Sprintf("threshold_%d", threshold), func(t *testing.T) {
				s := &bresenham{}
				var e int
				for i := 0; i < base; i++ {
					e += threshold
					if e >= base {
						e -= base
						s.lookup[i] = true
					}
				}

				trueCount := 0
				for i := 0; i < base; i++ {
					if s.lookup[i] {
						trueCount++
					}
				}

				if trueCount != threshold {
					t.Errorf("expected %d true, got %d", threshold, trueCount)
				}

				segments := 10
				segmentSize := base / segments
				expectedPerSegment := threshold / segments

				for seg := 0; seg < segments; seg++ {
					start := seg * segmentSize
					end := start + segmentSize
					count := 0

					for i := start; i < end; i++ {
						if s.lookup[i] {
							count++
						}
					}

					minExpected := int(float64(expectedPerSegment) * 0.8)
					maxExpected := int(float64(expectedPerSegment) * 1.2)

					if threshold > segments && (count < minExpected || count > maxExpected) {
						t.Errorf("segment %d: expected ~%d samples, got %d", seg, expectedPerSegment, count)
					}
				}
			})
		}
	})
	t.Run("deterministic", func(t *testing.T) {
		testfn := func(t *testing.T, s1, s2 *bresenham) {
			for i := 0; i < base; i++ {
				if s1.Sample() != s2.Sample() {
					t.Errorf("samplers with same sampleRate differ at index %d", i)
					break
				}
			}
		}
		testfn(t, newBresenham(0.33), newBresenham(0.33))
	})
	t.Run("bias/no local", func(t *testing.T) {
		type testcase struct {
			name       string
			sampleRate float64
			windowSize uint64
		}
		tests := []testcase{
			{
				name:       "one third local",
				sampleRate: 1.0 / 3.0,
				windowSize: 1000,
			},
			{
				name:       "half local",
				sampleRate: 0.5,
				windowSize: 1000,
			},
			{
				name:       "one percent local",
				sampleRate: 0.01,
				windowSize: 10000,
			},
		}

		const tolerance = 0.2
		testfn := func(t *testing.T, tt testcase, s *bresenham, base uint64) {
			for start := uint64(0); start < base-tt.windowSize; start += tt.windowSize {
				var windowDrops int
				for i := start; i < start+tt.windowSize; i++ {
					if !s.Sample() {
						windowDrops++
					}
				}

				expectedWindowDrops := int(float64(tt.windowSize) * (1 - tt.sampleRate))
				if expectedWindowDrops == 0 {
					expectedWindowDrops = 1
				}

				deviation := math.Abs(float64(windowDrops-expectedWindowDrops)) / float64(expectedWindowDrops)
				if deviation > tolerance {
					t.Errorf("local bias detected at window [%d, %d]: got %d drops, expected ~%d (deviation %.2f%%)",
						start, start+tt.windowSize, windowDrops, expectedWindowDrops, deviation*100)
				}
			}
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testfn(t, tt, newBresenham(tt.sampleRate), base)
			})
		}
	})
}

func BenchmarkBresenham(b *testing.B) {
	s := newBresenham(0.33)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Sample()
	}
}
