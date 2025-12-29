package stats

import (
	"math"
	"time"
)

// ComputeRTTStats 计算 RTT 样本的最小/最大/平均/标准差
func ComputeRTTStats(rtts []time.Duration) (min, max, avg, stddev time.Duration) {
	if len(rtts) == 0 {
		return
	}

	min = rtts[0]
	max = rtts[0]
	var sum time.Duration

	for _, rtt := range rtts {
		if rtt < min {
			min = rtt
		}
		if rtt > max {
			max = rtt
		}
		sum += rtt
	}

	avg = sum / time.Duration(len(rtts))

	var variance float64
	avgFloat := float64(avg.Nanoseconds())
	for _, rtt := range rtts {
		diff := float64(rtt.Nanoseconds()) - avgFloat
		variance += diff * diff
	}
	variance /= float64(len(rtts))
	stddev = time.Duration(math.Sqrt(variance))

	return
}
