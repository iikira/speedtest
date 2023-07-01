package speedtestclient

import (
	"sort"
	"time"
	"unsafe"
)

type (
	// HIRes HI 结果
	HIRes struct {
		Message string
		Latency time.Duration
	}

	// PingRes PING 结果
	PingRes struct {
		Latencies []time.Duration
		Average   time.Duration
		Min       time.Duration
		Max       time.Duration
		Median    time.Duration
	}

	// UpDownloadRes 下载或上传的结果
	UpDownloadRes struct {
		TimeElapsed       time.Duration
		SpeedsPerSecond   []int64
		MaxSpeedPerSecond int64
		MinSpeedPerSecond int64
		AverageSpeed      int64
		MedianSpeed       int64
	}

	PingCallback func(seq int, latency time.Duration)

	//UpDownloadCallback 上传或下载的回调
	UpDownloadCallback func(statistic *Statistic)

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

	for i, latency := range latencies {
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

		// An = [(n-1)An-1 + an]/n
		n := time.Duration(i + 1)
		res.Average = (n-1)*res.Average/n + latency/n
	}
	sort.Sort(TimeDurationSlice(latencies))
	res.Median = latencies[latenciesLen/2]
	return &res
}

func NewUpDownloadRes(timeElapsed time.Duration, statistic *Statistic) *UpDownloadRes {
	speedsLen := len(statistic.speedPerSeconds)
	res := UpDownloadRes{
		TimeElapsed:     timeElapsed,
		SpeedsPerSecond: make([]int64, 0, speedsLen),
	}

	if speedsLen == 0 {
		return &res
	}

	for i, speed := range statistic.speedPerSeconds {
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

		// An = [(n-1)An-1 + an]/n
		n := int64(i + 1)
		res.AverageSpeed = (n-1)*res.AverageSpeed/n + speed/n
	}

	sort.Sort(TimeDurationSlice(*(*[]time.Duration)(unsafe.Pointer(&statistic.speedPerSeconds))))
	res.MedianSpeed = statistic.speedPerSeconds[speedsLen/2]
	return &res
}

func (p TimeDurationSlice) Len() int           { return len(p) }
func (p TimeDurationSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p TimeDurationSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
