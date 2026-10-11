package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	stats := Stats{}
	numsLen := len(nums)

	if numsLen < 2 {
		return stats
	}

	initialMinimum := nums[1] - nums[0]
	initialMaximum := initialMinimum
	stats.Min = initialMinimum
	stats.Max = initialMaximum

	for i := 1; i < numsLen; i++ {
		neighborsDiff := nums[i] - nums[i-1]
		stats.Count++
		stats.Sum += neighborsDiff
		stats.Min = min(stats.Min, neighborsDiff)
		stats.Max = max(stats.Max, neighborsDiff)
	}

	return stats
}
