package types

import (
	"math"
	"testing"
	"time"
)

func TestPingResultUpdateStatistics(t *testing.T) {
	cases := []struct {
		name     string
		replies  []*PingReply
		sent     int
		received int
		loss     int
		lossRate float64
		minRTT   time.Duration
		maxRTT   time.Duration
		avgRTT   time.Duration
		stddev   time.Duration
	}{
		{
			name: "no replies",
		},
		{
			name: "mixed replies",
			replies: []*PingReply{
				{Status: StatusSuccess, RTT: 10 * time.Millisecond},
				{Status: StatusSuccess, RTT: 20 * time.Millisecond},
				{Status: StatusTimeout, RTT: 0},
			},
			sent:     3,
			received: 2,
			loss:     1,
			lossRate: 100.0 / 3.0,
			minRTT:   10 * time.Millisecond,
			maxRTT:   20 * time.Millisecond,
			avgRTT:   15 * time.Millisecond,
			stddev:   5 * time.Millisecond,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := &PingResult{
				Replies: tc.replies,
			}
			result.UpdateStatistics()
			if result.Statistics == nil {
				t.Fatalf("Statistics should not be nil")
			}
			stats := result.Statistics
			if stats.Sent != tc.sent {
				t.Fatalf("Sent: expected %d, got %d", tc.sent, stats.Sent)
			}
			if stats.Received != tc.received {
				t.Fatalf("Received: expected %d, got %d", tc.received, stats.Received)
			}
			if stats.Loss != tc.loss {
				t.Fatalf("Loss: expected %d, got %d", tc.loss, stats.Loss)
			}
			if math.Abs(stats.LossRate-tc.lossRate) > 0.000001 {
				t.Fatalf("LossRate: expected %.6f, got %.6f", tc.lossRate, stats.LossRate)
			}
			if stats.MinRTT != tc.minRTT {
				t.Fatalf("MinRTT: expected %v, got %v", tc.minRTT, stats.MinRTT)
			}
			if stats.MaxRTT != tc.maxRTT {
				t.Fatalf("MaxRTT: expected %v, got %v", tc.maxRTT, stats.MaxRTT)
			}
			if stats.AvgRTT != tc.avgRTT {
				t.Fatalf("AvgRTT: expected %v, got %v", tc.avgRTT, stats.AvgRTT)
			}
			if stats.StdDevRTT != tc.stddev {
				t.Fatalf("StdDevRTT: expected %v, got %v", tc.stddev, stats.StdDevRTT)
			}
		})
	}
}
