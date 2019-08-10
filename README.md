# speedtest

Command line client for speedtest.net

# Install

```
go get -u -v github.com/iikira/speedtest
```

# Usage
```
Usage of ./speedtest:
  -disable_down
        Disable DOWNLOAD
  -disable_hi
        Disable HI
  -disable_ping
        Disable PING
  -disable_up
        Disable UPLOAD
  -down_parallel int
        Max download parallel (default 2)
  -down_time string
        Download time (default "15s")
  -list_all
        list all Speedtest.net server, priority 0
  -list_nearby
        list nearby Speedtest.net server, priority 1
  -local_info
        get local info, e.g. ISP
  -ping_times int
        Times of PING (default 3)
  -proxy string
        http or socks proxy address
  -refresh_interval string
        Upload or Download refresh interval (default "1s")
  -server_host string
        Speedtest.net server host, priority 3
  -server_id int
        Speedtest.net server id, priority 2
  -source_addr string
        Local source address, priority 0
  -source_interface string
        Local source interface, priority 1
  -up_parallel int
        Max upload parallel (default 2)
  -up_time string
        Upload time (default "15s")
```
