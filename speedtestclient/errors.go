package speedtestclient

import (
	"errors"
	"net"
)

var (
	ErrHiResponse     = errors.New("unexpected HI response")
	ErrPingResponse   = errors.New("unexpected PING response")
	ErrNotSocks5Proxy = errors.New("not socks5 proxy")
)

func IsTimeout(err error) bool {
	netError, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	return netError.Err.Error() == "i/o timeout"
}
