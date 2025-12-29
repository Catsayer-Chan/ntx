package stats

import (
	"math"
	"testing"
	"time"
)

func TestComputeRTTStats(t *testing.T) {
	cases := []struct {
		name   string
		rtts   []time.Duration
		min    time.Duration
		max    time.Duration
		avg    time.Duration
		stddev time.Duration
	}{
		{
			name: "empty slice",
		},
		{
			name:   "single rtt",
			rtts:   []time.Duration{5 * time.Millisecond},
			min:    5 * time.Millisecond,
			max:    5 * time.Millisecond,
			avg:    5 * time.Millisecond,
			stddev: 0,
		},
		{
			name:   "multiple rtts",
			rtts:   []time.Duration{time.Millisecond, 3 * time.Millisecond, 5 * time.Millisecond},
			min:    time.Millisecond,
			max:    5 * time.Millisecond,
			avg:    3 * time.Millisecond,
			stddev: time.Duration(float64(time.Millisecond) * math.Sqrt(8.0/3.0)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			min, max, avg, stddev := ComputeRTTStats(tc.rtts)
			if min != tc.min {
				t.Fatalf("min: expected %v, got %v", tc.min, min)
			}
			if max != tc.max {
				t.Fatalf("max: expected %v, got %v", tc.max, max)
			}
			if avg != tc.avg {
				t.Fatalf("avg: expected %v, got %v", tc.avg, avg)
			}
			if stddev != tc.stddev {
				t.Fatalf("stddev: expected %v, got %v", tc.stddev, stddev)
			}
		})
	}
}
