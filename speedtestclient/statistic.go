package speedtestclient

import (
	"github.com/iikira/iikira-go-utils/utils/expires"
	"sync/atomic"
	"time"
)

type (
	// Statistic 统计
	Statistic struct {
		totalSize       int64     // 总大小
		transferSize    int64     // 已传输的数据量
		speedPerSecond  int64     // 速度
		speedPerSeconds []int64   // 用来计算平均速度的
		startTime       time.Time // 启动时间
		deadline        time.Time // 截止时间
	}
)

func (s *Statistic) TotalSize() int64 {
	return atomic.LoadInt64(&s.totalSize)
}

func (s *Statistic) TransferSize() int64 {
	return s.transferSize
}

func (s *Statistic) SpeedPerSecond() int64 {
	return s.speedPerSecond
}

func (s *Statistic) AppendSpeedPerSecond(speed int64) {
	s.speedPerSeconds = append(s.speedPerSeconds, speed)
}

func (s *Statistic) Elapsed() (elapsed time.Duration) {
	elapsed = time.Now().Sub(s.startTime).Round(100 * time.Millisecond)
	return elapsed
}

func (s *Statistic) ElapsedAndLeft() (elapsed, left time.Duration) {
	nowTime := time.Now()
	elapsed = nowTime.Sub(s.startTime).Round(100 * time.Millisecond)
	left = s.deadline.Sub(nowTime).Round(100 * time.Millisecond)
	if left < 0 {
		left = 0
	}
	return elapsed, left
}

func (s *Statistic) StartTimer() {
	s.startTime = time.Now()
	expires.StripMono(&s.startTime)
}

func (s *Statistic) AddTransferSize(size int64) int64 {
	return atomic.AddInt64(&s.transferSize, size)
}
