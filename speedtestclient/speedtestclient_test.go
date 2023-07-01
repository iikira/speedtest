package speedtestclient_test

import (
	"fmt"
	"github.com/iikira/iikira-go-utils/utils/converter"
	"github.com/iikira/speedtest/speedtestclient"
	"testing"
	"time"
)

var (
	Client   = speedtestclient.NewSpeedtestClient()
	WithHost = Client.WithHost("speedtest.zjmobile.com:8080")
)

func TestGetLocalInfoAndServerList(t *testing.T) {
	li, servList, err := Client.GetLocalInfoAndServerList()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", li)
	t.Logf("server list: \n")
	for k, serv := range servList {
		t.Logf("[%d] %#v\n", k, serv)
	}
}

func TestGetAllServerList(t *testing.T) {
	servList, err := Client.GetAllServerList()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("server list: \n")
	for k, serv := range servList {
		t.Logf("[%d] %#v\n", k, serv)
	}
}

func TestHi(t *testing.T) {
	res, err := WithHost.HI()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", res)
	t.Logf("Latency: %s\n", res.Latency)
}

func TestPing(t *testing.T) {
	res, err := WithHost.Ping(10, 1*time.Second, func(seq int, latency time.Duration) {
		t.Logf("[%d] PING Latency: %s\n", seq, latency)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", res)
}

func TestDownload(t *testing.T) {
	res, err := WithHost.Download(&speedtestclient.UpDownloadOption{
		Timeout:          15 * time.Second,
		Parallel:         2,
		CallbackInterval: 350 * time.Millisecond,
	}, func(statistic *speedtestclient.Statistic) {
		elapsed, left := statistic.ElapsedAndLeft()
		fmt.Printf("↓ %s/s in %s, left %s ... \n", converter.ConvertFileSize(statistic.SpeedPerSecond(), 2), elapsed/1e7*1e7, left/1e7*1e7)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", res)
}

func TestUpload(t *testing.T) {
	res, err := WithHost.Upload(&speedtestclient.UpDownloadOption{
		Timeout:          15 * time.Second,
		Parallel:         2,
		CallbackInterval: 350 * time.Millisecond,
	}, func(statistic *speedtestclient.Statistic) {
		elapsed, left := statistic.ElapsedAndLeft()
		fmt.Printf("↑ %s/s in %s, left %s ... \n", converter.ConvertFileSize(statistic.SpeedPerSecond(), 2), elapsed/1e7*1e7, left/1e7*1e7)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", res)
}
