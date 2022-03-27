package speedtestclient

import (
	"fmt"
	"github.com/iikira/iikira-go-utils/requester"
	"net/url"
)

const (
	SpeedtestHost = "www.speedtest.net"
)

type (
	SpeedtestClient struct {
		hc *requester.HTTPClient
	}
)

func NewSpeedtestClient() *SpeedtestClient {
	return &SpeedtestClient{}
}

func (sc *SpeedtestClient) lazyInit() {
	if sc.hc == nil {
		sc.hc = requester.NewHTTPClient()
	}
}

// SetProxy 设置代理
func (sc *SpeedtestClient) SetProxy(proxyAddr string) {
	sc.lazyInit()
	sc.hc.SetProxy(proxyAddr)
}

func (sc *SpeedtestClient) genURL(path string, param map[string]interface{}) *url.URL {
	u := url.URL{
		Scheme: "https",
		Host:   SpeedtestHost,
		Path:   path,
	}
	if param == nil {
		return &u
	}

	uv := u.Query()
	for k, v := range param {
		uv.Set(k, fmt.Sprint(v))
	}

	u.RawQuery = uv.Encode()
	return &u
}
