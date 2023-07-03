package speedtestclient

import (
	"bytes"
	"context"
	"github.com/iikira/iikira-go-utils/requester/rio/speeds"
	"github.com/iikira/iikira-go-utils/utils/cachepool"
	"github.com/iikira/iikira-go-utils/utils/converter"
	"github.com/iikira/speedtest/speedtestutil/bytemessage"
	"golang.org/x/net/proxy"
	"net"
	"net/url"
	"time"
)

const (
	PingTimeout    = 10 * time.Second
	HiTimeout      = 1 * time.Minute
	UpDownloadSize = 128 * converter.PB
)

type (
	SpeedtestClientWithHost struct {
		SpeedtestClient
		Host      string
		socks5URL *url.URL
		localAddr *net.TCPAddr
	}

	UpDownloadOption struct {
		Timeout          time.Duration
		Parallel         int
		CallbackInterval time.Duration // 回调函数调用的时间间隔
	}

	upDownloadHandleFunc func(commonBuf []byte, errChan chan<- error, statistic *Statistic, speedStat *speeds.Speeds)
)

func (sc *SpeedtestClient) WithHost(host string) *SpeedtestClientWithHost {
	return &SpeedtestClientWithHost{
		SpeedtestClient: *sc,
		Host:            host,
	}
}

func (sch *SpeedtestClientWithHost) SetSocks5Proxy(proxyAddr string) (err error) {
	u, err := url.Parse(proxyAddr)
	if err != nil {
		return
	}
	if u.Scheme != "socks5" {
		err = ErrNotSocks5Proxy
		return
	}
	sch.socks5URL = u
	return
}

func (sch *SpeedtestClientWithHost) SetLocalAddr(localAddr *net.TCPAddr) {
	sch.localAddr = localAddr
}

func (sch *SpeedtestClientWithHost) dialHost() (tcpConn *net.TCPConn, err error) {
	var dialer proxy.Dialer
	dialer = &net.Dialer{
		LocalAddr: sch.localAddr,
	}
	if sch.socks5URL != nil {
		dialer, err = proxy.FromURL(sch.socks5URL, dialer)
		if err != nil {
			return nil, err
		}
	}

	conn, err := dialer.Dial("tcp", sch.Host)
	if err != nil {
		return
	}

	tcpConn = conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	return
}

func (sch *SpeedtestClientWithHost) HI() (res *HIRes, err error) {
	conn, err := sch.dialHost()
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(HiTimeout))
	nowTime := time.Now()
	_, err = conn.Write(bytemessage.Smessagef("HI\n"))
	if err != nil {
		return
	}

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	if bytes.Index(buf[:n], []byte("HELLO")) != 0 {
		err = ErrHiResponse
		return
	}

	res = &HIRes{
		Message: converter.ToString(bytes.TrimSuffix(buf[:n], []byte{'\n'})),
		Latency: time.Since(nowTime),
	}
	return
}

func (sch *SpeedtestClientWithHost) Ping(times int, sleep time.Duration, callback PingCallback) (res *PingRes, err error) {
	if times < 1 {
		res = NewPingRes(nil)
		return
	}

	conn, err := sch.dialHost()
	if err != nil {
		return
	}
	defer conn.Close()

	var (
		buf       = make([]byte, 256)
		latencies = make([]time.Duration, 0, times)
		n         int
	)
	for i := 0; i < times; i++ {
		conn.SetDeadline(time.Now().Add(PingTimeout))
		nowTime := time.Now()
		_, err = conn.Write(bytemessage.Smessagef("PING %d\n", nowTime.UnixNano()/1e6))
		if err != nil {
			if IsTimeout(err) {
				latencies = append(latencies, -1)
				continue
			}
			return
		}

		// 读取响应
		n, err = conn.Read(buf)
		if err != nil {
			if IsTimeout(err) {
				latencies = append(latencies, -1)
				continue
			}
			return
		}

		// 计算延时
		latency := time.Since(nowTime)

		fields := bytes.Fields(bytes.TrimSuffix(buf[:n], []byte{'\n'}))
		if len(fields) != 2 {
			err = ErrPingResponse
			return
		}
		if bytes.Compare(fields[0], []byte("PONG")) != 0 {
			err = ErrPingResponse
			return
		}

		latencies = append(latencies, latency)
		if callback != nil {
			callback(i, latency)
		}
		time.Sleep(sleep)
	}

	err = nil
	res = NewPingRes(latencies)
	return
}

func (sch *SpeedtestClientWithHost) upDownload(opt *UpDownloadOption, callback UpDownloadCallback, gofn upDownloadHandleFunc) (res *UpDownloadRes, err error) {
	if opt == nil {
		opt = &UpDownloadOption{
			Timeout:          15 * time.Second,
			Parallel:         1,
			CallbackInterval: 500 * time.Millisecond,
		}
	} else if opt.Parallel < 1 {
		opt.Parallel = 1
	}

	var (
		// 统计
		statistic = Statistic{
			totalSize:       UpDownloadSize,
			speedPerSeconds: make([]int64, 0, 32),
			deadline:        time.Now().Add(opt.Timeout),
		}
		speedStat   = speeds.Speeds{} // 计算速度
		ticker      = time.NewTicker(opt.CallbackInterval)
		ctx, cancel = context.WithDeadline(context.Background(), statistic.deadline)
		commonBuf   = cachepool.RawMallocByteSlice(2048)
		errChan     = make(chan error, opt.Parallel)
	)
	defer cancel()

	statistic.StartTimer() // 开始计时
	for i := 0; i < opt.Parallel; i++ {
		go gofn(commonBuf, errChan, &statistic, &speedStat)
	}

	// 监控 直到达到时间
	// TODO: 这样会频繁创建goroutine, 可优化
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err = <-errChan:
				if err != nil {
					cancel()
					return
				}
				// 下一轮
				go gofn(commonBuf, errChan, &statistic, &speedStat)
			}
		}
	}()

	go func() { // start callback
		for {
			select {
			case <-ticker.C:
				speed := speedStat.GetSpeeds()
				statistic.AppendSpeedPerSecond(speed)

				// 更新统计
				statistic.speedPerSecond = speed
				if callback != nil {
					callback(&statistic)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	ticker.Stop()

	elapsed := statistic.Elapsed()
	res = NewUpDownloadRes(elapsed, &statistic)
	return
}

func (sch *SpeedtestClientWithHost) Download(opt *UpDownloadOption, callback UpDownloadCallback) (res *UpDownloadRes, err error) {
	return sch.upDownload(opt, callback, func(commonBuf []byte, errChan chan<- error, statistic *Statistic, speedStat *speeds.Speeds) {
		conn, err := sch.dialHost()
		if err != nil {
			errChan <- err
			return
		}
		defer conn.Close()

		// 1分钟
		// 超时会产生错误：io.EOF
		conn.SetDeadline(time.Now().Add(1 * time.Minute))
		_, err = conn.Write(bytemessage.Smessagef("DOWNLOAD %d\n", UpDownloadSize))
		if err != nil { // 暂不处理
			errChan <- err
			return
		}

		var n int
		for {
			n, err = conn.Read(commonBuf)
			speedStat.Add(int64(n))
			statistic.AddTransferSize(int64(n)) // 增加
			if err != nil {
				if IsTimeout(err) { // 达到deadline
					err = nil
					break
				}
				break
			}
		}
		errChan <- err
		return
	})
}

func (sch *SpeedtestClientWithHost) Upload(opt *UpDownloadOption, callback UpDownloadCallback) (res *UpDownloadRes, err error) {
	return sch.upDownload(opt, callback, func(commonBuf []byte, errChan chan<- error, statistic *Statistic, speedStat *speeds.Speeds) {
		conn, err := sch.dialHost()
		if err != nil {
			errChan <- err
			return
		}
		defer conn.Close()

		// 1分钟
		// 超时会产生错误：broken pipe
		conn.SetDeadline(time.Now().Add(1 * time.Minute))
		_, err = conn.Write(bytemessage.Smessagef("UPLOAD %d\n", UpDownloadSize))
		if err != nil { // 暂不处理
			errChan <- err
			return
		}

		var n int
		for {
			n, err = conn.Write(commonBuf)
			speedStat.Add(int64(n))
			statistic.AddTransferSize(int64(n)) // 增加
			if err != nil {
				if IsTimeout(err) { // 达到deadline
					err = nil
					break
				}
				break
			}
		}
		errChan <- err
		return
	})
}
