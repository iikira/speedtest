package main

import (
	"flag"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/speedtest/speedtestclient"
	"github.com/iikira/speedtest/speedtestutil/interfaceutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	isListAll           bool
	isListNearby        bool
	isGetLocalInfo      bool
	speedtestServerID   int
	speedtestServerHost string
	uploadParallel      int
	downloadParallel    int
	uploadTime          string
	downloadTime        string
	pingTimes           int
	disableUpload       bool
	disableDownload     bool
	disableHi           bool
	disablePing         bool
	sourceAddr          string
	sourceInterface     string
	proxy               string

	refreshInterval string

	client    *speedtestclient.SpeedtestClient
	localAddr *net.TCPAddr
)

func init() {
	flag.BoolVar(&isListAll, "list_all", false, "list all Speedtest.net server, priority 0")
	flag.BoolVar(&isListNearby, "list_nearby", false, "list nearby Speedtest.net server, priority 1")
	flag.BoolVar(&isGetLocalInfo, "local_info", false, "get local info, e.g. ISP")
	flag.IntVar(&speedtestServerID, "server_id", 0, "Speedtest.net server id, priority 2")
	flag.StringVar(&speedtestServerHost, "server_host", "", "Speedtest.net server host, priority 3")
	flag.IntVar(&uploadParallel, "up_parallel", 2, "Max upload parallel")
	flag.IntVar(&downloadParallel, "down_parallel", 2, "Max download parallel")
	flag.StringVar(&uploadTime, "up_time", "15s", "Upload time")
	flag.StringVar(&downloadTime, "down_time", "15s", "Download time")
	flag.IntVar(&pingTimes, "ping_times", 3, "Times of PING")
	flag.BoolVar(&disableUpload, "disable_up", false, "Disable UPLOAD")
	flag.BoolVar(&disableDownload, "disable_down", false, "Disable DOWNLOAD")
	flag.BoolVar(&disableHi, "disable_hi", false, "Disable HI")
	flag.BoolVar(&disablePing, "disable_ping", false, "Disable PING")
	flag.StringVar(&sourceAddr, "source_addr", "", "Local source address, priority 0")
	flag.StringVar(&sourceInterface, "source_interface", "", "Local source interface, priority 1")
	flag.StringVar(&proxy, "proxy", "", "http or socks proxy address")
	flag.StringVar(&refreshInterval, "refresh_interval", "1s", "Upload or Download refresh interval")
	flag.Parse()

	client = speedtestclient.NewSpeedtestClient()
	client.SetProxy(proxy)
}

func main() {
	var err error

	// set local source addr
	if sourceAddr != "" {
		localAddr = &net.TCPAddr{
			IP: net.ParseIP(sourceAddr),
		}
		requester.SetLocalTCPAddrList(sourceAddr)
	} else if sourceInterface != "" {
		localAddr, err = interfaceutil.GetAvaliableLocalTCPAddr(sourceInterface)
		if err != nil {
			log.Fatalf("get avaliable interface source addr error: %s, please specify source_addr\n", err)
		}
		requester.SetLocalTCPAddrList(localAddr.IP.String())
	}

	if isListAll {
		servList, err := client.GetAllServerList()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("All Server List: ")
		servList.PrintTo(os.Stdout)
		return
	}

	// list server or local info
	if isListNearby || isGetLocalInfo {
		li, servList, err := client.GetLocalInfoAndServerList()
		if err != nil {
			log.Fatalln(err)
		}
		if isListNearby {
			fmt.Println("Nearby Server List: ")
			servList.PrintTo(os.Stdout)
		}
		if isGetLocalInfo {
			fmt.Println("Local Info: ")
			li.PrintTo(os.Stdout)
		}
		return
	}

	// query server host by id
	if speedtestServerHost == "" {
		if speedtestServerID != 0 {
			servList, err := client.GetAllServerList()
			if err != nil {
				log.Fatalln(err)
			}
			server := servList.FindByID(speedtestServerID)
			if server == nil {
				log.Fatalf("server host not found, id: %d\n", speedtestServerID)
			}

			speedtestServerHost = server.Host
			log.Printf("server found, %s\n", server)
		} else if speedtestServerHost == "" {
			// default, find host
			_, servList, err := client.GetLocalInfoAndServerList()
			if err != nil {
				log.Fatalln(err)
			}
			if len(servList) == 0 {
				log.Fatalf("server not found\n")
			}

			speedtestServerHost = servList[0].Host
			log.Printf("server found, %s\n", servList[0])
		}
	}

	withHost := client.WithHost(speedtestServerHost)

	// set proxy
	if proxy != "" {
		err = withHost.SetSocks5Proxy(proxy)
		if err != nil {
			log.Fatalf("set proxy error: %s\n", err)
		}
	}

	// set local addr
	withHost.SetLocalAddr(localAddr)

	// hi
	if !disableHi {
		hiRes, err := withHost.HI()
		if err != nil {
			log.Fatalf("HI errro: %s\n", err)
		}
		fmt.Printf("HI success, latency: %s\n", hiRes.Latency)
	}

	// ping
	if !disablePing {
		pingRes, err := withHost.Ping(pingTimes, 1*time.Second, func(seq int, latency time.Duration) {
			log.Printf("[%d] PING %s\n", seq, latency)
		})
		if err != nil {
			log.Fatalf("PING errro: %s\n", err)
		}

		fmt.Printf("PING RES: min/avg/max/median = %s/%s/%s/%s\n", pingRes.Min, pingRes.Average, pingRes.Max, pingRes.Median)
	}

	opt := speedtestclient.UpDownloadOption{}
	opt.CallbackInterval, err = time.ParseDuration(strings.ToLower(refreshInterval))
	if err != nil {
		log.Fatalf("parse refresh_interval error: %s\n", err)
	}

	if !disableDownload {
		opt.Timeout, err = time.ParseDuration(strings.ToLower(downloadTime))
		if err != nil {
			log.Fatalf("DOWNLOAD: parse down_time error: %s\n", err)
		}

		opt.Parallel = downloadParallel
		downRes, err := withHost.Download(&opt, upDownCallback("↓"))
		if err != nil {
			log.Fatalf("DOWNLOAD error: %s\n", err)
		}

		printRes("DOWNLOAD", downRes)
	}

	if !disableUpload {
		opt.Timeout, err = time.ParseDuration(strings.ToLower(uploadTime))
		if err != nil {
			log.Fatalf("UPLOAD: parse up_time error: %s\n", err)
		}

		opt.Parallel = uploadParallel
		upRes, err := withHost.Upload(&opt, upDownCallback("↑"))
		if err != nil {
			log.Fatalf("UPLOAD error: %s\n", err)
		}

		printRes("UPLOAD", upRes)
	}

}

func printRes(op string, res *speedtestclient.UpDownloadRes) {
	fmt.Printf(op+" RES: min/avg/max/median = %s/%s/%s/%s per second\n", converter.ConvertFileSize(res.MinSpeedPerSecond, 2), converter.ConvertFileSize(res.AverageSpeed, 2), converter.ConvertFileSize(res.MaxSpeedPerSecond, 2), converter.ConvertFileSize(res.MedianSpeed, 2))
}

func upDownCallback(character string) speedtestclient.UpDownloadCallback {
	return func(statistic *speedtestclient.Statistic) {
		elapsed, left := statistic.ElapsedAndLeft()
		log.Printf(
			character+" %s %s/s in %s, left %s\n",
			converter.ConvertFileSize(statistic.TransferSize(), 2),
			converter.ConvertFileSize(statistic.SpeedPerSecond(), 2),
			elapsed/1e7*1e7,
			left/1e7*1e7,
		)
	}
}
