package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	numsLen := len(nums)

	if numsLen < 2 {
		return Stats{}
	}

	firstDiff := nums[1] - nums[0]
	stats := Stats{Min: firstDiff, Max: firstDiff}

	for i := 1; i < numsLen; i++ {
		neighborsDiff := nums[i] - nums[i-1]
		stats.Count++
		stats.Sum += neighborsDiff
		stats.Min = min(stats.Min, neighborsDiff)
		stats.Max = max(stats.Max, neighborsDiff)
	}

	return stats
}
