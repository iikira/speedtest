package interfaceutil

import (
	"errors"
	"net"
	"os"
)

var (
	ErrNoSuchAddr = errors.New("no such local addr")
)

func ParseLocalAddr(interfaceName string) (ipnets []*net.IPNet, err error) {
	i, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return
	}

	addrs, err := i.Addrs()
	if err != nil {
		return nil, err
	}

	ipnets = make([]*net.IPNet, 0, len(addrs))
	for _, addr := range addrs {
		ipnets = append(ipnets, addr.(*net.IPNet))
	}

	return ipnets, nil
}

func GetAvaliableLocalTCPAddr(interfaceName string) (tcpAddr *net.TCPAddr, err error) {
	ipnets, err := ParseLocalAddr(interfaceName)
	if err != nil {
		return
	}
	for _, ipnet := range ipnets {
		pass, _ := CheckLocalTCPAddr(&net.TCPAddr{IP: ipnet.IP})
		if pass {
			tcpAddr = &net.TCPAddr{
				IP: ipnet.IP,
			}
			return
		}
	}
	return nil, ErrNoSuchAddr
}

func CheckLocalTCPAddrByString(addr string) (pass bool, err error) {
	if addr == "" {
		return
	}

	localAddr := &net.TCPAddr{
		IP: net.ParseIP(addr),
	}

	return CheckLocalTCPAddr(localAddr)
}

func CheckLocalTCPAddr(tcpAddr *net.TCPAddr) (pass bool, err error) {
	dialer := &net.Dialer{
		LocalAddr: tcpAddr,
	}

	var testAddr string
	to4 := tcpAddr.IP.To4()
	if to4 != nil { //ipv4
		testAddr = "127.0.0.1"
	} else { // ipv6
		testAddr = "[::1]"
	}
	conn, err := dialer.Dial("tcp", testAddr+":1")
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		netError, ok := err.(*net.OpError)
		if !ok {
			return
		}
		syscallError, ok := netError.Err.(*os.SyscallError)
		if !ok {
			return
		}
		if syscallError.Err.Error() == "connection refused" {
			return true, nil
		}
		return
	}
	return true, nil
}
