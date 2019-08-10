package speedtestclient

import (
	"sort"
	"time"
	"unsafe"
)

type (
	HIRes struct {
		Message string
		Latency time.Duration
	}

	PingRes struct {
		Latencies []time.Duration
		Average   time.Duration
		Min       time.Duration
		Max       time.Duration
		Median    time.Duration
		Total     time.Duration
	}

	UpDownloadRes struct {
		TimeElapsed       time.Duration
		SpeedsPerSecond   []int64
		MaxSpeedPerSecond int64
		MinSpeedPerSecond int64
		AverageSpeed      int64
		MedianSpeed       int64
	}

	PingCallback       func(seq int, latency time.Duration)
	UpDownloadCallback func(speedPerSecond int64, elapsed, left time.Duration)

	TimeDurationSlice []time.Duration
)

func NewPingRes(latencies []time.Duration) *PingRes {
	latenciesLen := len(latencies)
	res := PingRes{
		Latencies: make([]time.Duration, 0, latenciesLen),
	}

	if latenciesLen == 0 {
		return &res
	}

	for _, latency := range latencies {
		res.Latencies = append(res.Latencies, latency)
		if latency == -1 { // -1为超时
			continue
		}
		if latency > res.Max {
			res.Max = latency
		}
		if latency < res.Min || res.Min == 0 {
			res.Min = latency
		}
		res.Total += latency
	}
	res.Average = res.Total / time.Duration(latenciesLen)
	sort.Sort(TimeDurationSlice(latencies))
	res.Median = latencies[latenciesLen/2]
	return &res
}

func NewUpDownloadRes(timeElapsed time.Duration, speeds []int64) *UpDownloadRes {
	speedsLen := len(speeds)
	res := UpDownloadRes{
		TimeElapsed:     timeElapsed,
		SpeedsPerSecond: make([]int64, 0, speedsLen),
	}

	if speedsLen == 0 {
		return &res
	}

	var total int64
	for _, speed := range speeds {
		res.SpeedsPerSecond = append(res.SpeedsPerSecond, speed)
		if speed > res.MaxSpeedPerSecond {
			res.MaxSpeedPerSecond = speed
			if res.MinSpeedPerSecond == 0 {
				res.MinSpeedPerSecond = speed
			}
		}
		if speed < res.MinSpeedPerSecond {
			res.MinSpeedPerSecond = speed
		}
		total += speed
	}

	res.AverageSpeed = total / int64(speedsLen)
	sort.Sort(TimeDurationSlice(*(*[]time.Duration)(unsafe.Pointer(&speeds))))
	res.MedianSpeed = speeds[speedsLen/2]
	return &res
}

func (p TimeDurationSlice) Len() int           { return len(p) }
func (p TimeDurationSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p TimeDurationSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
