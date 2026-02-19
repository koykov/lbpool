package lbpool

import (
	"fmt"
	"math"
	"testing"
)

func TestSampler(t *testing.T) {
	t.Run("count drops", func(t *testing.T) {
		type testcase struct {
			name          string
			releaseFactor float64
			expectedDrops int
		}
		tests := []testcase{
			{
				name:          "zero",
				releaseFactor: 0,
				expectedDrops: 0,
			},
			{
				name:          "one",
				releaseFactor: 1,
				expectedDrops: base,
			},
			{
				name:          "min release factor",
				releaseFactor: 0.00001,
				expectedDrops: 1,
			},
			{
				name:          "max release factor",
				releaseFactor: 0.99999,
				expectedDrops: base - 1,
			},
			{
				name:          "one percent",
				releaseFactor: 0.01,
				expectedDrops: 1000,
			},
			{
				name:          "one third",
				releaseFactor: 1.0 / 3.0,
				expectedDrops: int(math.Floor(1.0/3.0*base + epsilon)),
			},
			{
				name:          "half",
				releaseFactor: 0.5,
				expectedDrops: base / 2,
			},
			{
				name:          "two thirds",
				releaseFactor: 2.0 / 3.0,
				expectedDrops: int(math.Floor(2.0/3.0*base + epsilon)),
			},
			{
				name:          "clamping below min",
				releaseFactor: 0.000001,
				expectedDrops: 0,
			},
			{
				name:          "clamping above max",
				releaseFactor: 0.999999,
				expectedDrops: base,
			},
			{
				name:          "NaN",
				releaseFactor: math.NaN(),
				expectedDrops: 0,
			},
			{
				name:          "+Inf",
				releaseFactor: math.Inf(1),
				expectedDrops: 0,
			},
			{
				name:          "-Inf",
				releaseFactor: math.Inf(-1),
				expectedDrops: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := newSampler(base, tt.releaseFactor)

				dropCount := 0
				for i := 0; i < base; i++ {
					if s.lookup[i] {
						dropCount++
					}
				}

				if dropCount != tt.expectedDrops {
					t.Errorf("expected %d drops, got %d for releaseFactor=%v",
						tt.expectedDrops, dropCount, tt.releaseFactor)
				}

				if tt.releaseFactor == 0 {
					for i := 0; i < base; i++ {
						if s.lookup[i] {
							t.Errorf("expected no drops for releaseFactor=0, but found at index %d", i)
							break
						}
					}
				}

				if tt.releaseFactor == 1 {
					for i := 0; i < base; i++ {
						if !s.lookup[i] {
							t.Errorf("expected all drops for releaseFactor=1, but found false at index %d", i)
							break
						}
					}
				}
			})
		}
	})
	t.Run("distribution/uniform", func(t *testing.T) {
		type testcase struct {
			name          string
			releaseFactor float64
			iterations    uint64
		}
		tests := []testcase{
			{
				name:          "one percent uniform",
				releaseFactor: 0.01,
				iterations:    100,
			},
			{
				name:          "one third uniform",
				releaseFactor: 1.0 / 3.0,
				iterations:    100,
			},
			{
				name:          "half uniform",
				releaseFactor: 0.5,
				iterations:    100,
			},
			{
				name:          "two thirds uniform",
				releaseFactor: 2.0 / 3.0,
				iterations:    100,
			},
			{
				name:          "ninety nine percent",
				releaseFactor: 0.99,
				iterations:    100,
			},
			{
				name:          "min release factor",
				releaseFactor: 0.00001,
				iterations:    10000,
			},
		}

		const tolerance = 0.01
		testfn := func(t *testing.T, tt testcase, s sampler, base uint64) {
			totalRequests := tt.iterations * base
			totalDrops := 0

			for i := uint64(0); i < totalRequests; i++ {
				if s.shouldDrop(i) {
					totalDrops++
				}
			}

			expectedDrops := int(float64(totalRequests) * tt.releaseFactor)

			deviation := math.Abs(float64(totalDrops-expectedDrops)) / float64(expectedDrops)
			if deviation > tolerance {
				t.Errorf("distribution deviation too high: got %d drops, expected %d (deviation %.2f%%)",
					totalDrops, expectedDrops, deviation*100)
			}
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Run("bsampler", func(t *testing.T) {
					testfn(t, tt, newSampler(base, tt.releaseFactor), base)
				})
				t.Run("sampler8", func(t *testing.T) {
					testfn(t, tt, newSampler8(base, tt.releaseFactor), base)
				})
				t.Run("sampler64", func(t *testing.T) {
					testfn(t, tt, newSampler64(base, tt.releaseFactor), base)
				})
			})
		}
	})
	t.Run("distribution/Bresenham", func(t *testing.T) {
		thresholds := []int{1, 2, 3, 10, 100, 1000, 50000}

		for _, threshold := range thresholds {
			t.Run(fmt.Sprintf("threshold_%d", threshold), func(t *testing.T) {
				s := &bsampler{}
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
						t.Errorf("segment %d: expected ~%d drops, got %d", seg, expectedPerSegment, count)
					}
				}
			})
		}
	})
	t.Run("deterministic", func(t *testing.T) {
		testfn := func(t *testing.T, s1, s2 sampler, base uint64) {
			for i := uint64(0); i < base; i++ {
				if s1.shouldDrop(i) != s2.shouldDrop(i) {
					t.Errorf("samplers with same releaseFactor differ at index %d", i)
					break
				}
			}
		}
		t.Run("bsampler", func(t *testing.T) {
			testfn(t, newSampler(base, 0.33), newSampler(base, 0.33), base)
		})
		t.Run("sampler8", func(t *testing.T) {
			testfn(t, newSampler8(base, 0.33), newSampler8(base, 0.33), base)
		})
		t.Run("sampler64", func(t *testing.T) {
			testfn(t, newSampler64(base, 0.33), newSampler64(base, 0.33), base)
		})
	})
	t.Run("bias/no local", func(t *testing.T) {
		type testcase struct {
			name          string
			releaseFactor float64
			windowSize    uint64
		}
		tests := []testcase{
			{
				name:          "one third local",
				releaseFactor: 1.0 / 3.0,
				windowSize:    1000,
			},
			{
				name:          "half local",
				releaseFactor: 0.5,
				windowSize:    1000,
			},
			{
				name:          "one percent local",
				releaseFactor: 0.01,
				windowSize:    10000,
			},
		}

		const tolerance = 0.2
		testfn := func(t *testing.T, tt testcase, s sampler, base uint64) {
			for start := uint64(0); start < base-tt.windowSize; start += tt.windowSize {
				windowDrops := 0
				for i := start; i < start+tt.windowSize; i++ {
					if s.shouldDrop(i) {
						windowDrops++
					}
				}

				expectedWindowDrops := int(float64(tt.windowSize) * tt.releaseFactor)
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
				t.Run("bsampler", func(t *testing.T) {
					testfn(t, tt, newSampler(base, tt.releaseFactor), base)
				})
				t.Run("sampler8", func(t *testing.T) {
					testfn(t, tt, newSampler8(base, tt.releaseFactor), base)
				})
				t.Run("sampler64", func(t *testing.T) {
					testfn(t, tt, newSampler64(base, tt.releaseFactor), base)
				})
			})
		}
	})
}

func BenchmarkSampler(b *testing.B) {
	benchfn := func(b *testing.B, s sampler) {
		b.ReportAllocs()
		b.ResetTimer()
		var counter uint64
		for i := 0; i < b.N; i++ {
			_ = s.shouldDrop(counter)
			counter++
		}
	}
	b.Run("bsampler", func(b *testing.B) {
		benchfn(b, newSampler(base, 0.33))
	})
	b.Run("sampler8", func(b *testing.B) {
		benchfn(b, newSampler8(base, 0.33))
	})
	b.Run("sampler64", func(b *testing.B) {
		benchfn(b, newSampler64(base, 0.33))
	})
}
