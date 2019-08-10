package interfaceutil_test

import (
	"github.com/iikira/speedtest/speedtestutil/interfaceutil"
	"testing"
)

func TestParseLocalAddr(t *testing.T) {
	addrs, err := interfaceutil.ParseLocalAddr("en0")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(addrs)
	if len(addrs) == 0 {
		t.Fatal(interfaceutil.ErrNoSuchAddr)
	}
	t.Logf("network: %s, address: %s\n", addrs[0].Network(), addrs[0].String())
}

func TestCheckLocalTCPAddrByString(t *testing.T) {
	pass, err := interfaceutil.CheckLocalTCPAddrByString("fe80::1488:fbe3:26f2:d174")
	if !pass {
		t.Logf("failed, %s\n", err)
	} else {
		t.Log("pass")
	}

	pass, err = interfaceutil.CheckLocalTCPAddrByString("192.168.50.179")
	if !pass {
		t.Logf("failed, %s\n", err)
	} else {
		t.Log("pass")
	}
}

func TestGetAvaliableLocalTCPAddr(t *testing.T) {
	tcpAddr, err := interfaceutil.GetAvaliableLocalTCPAddr("lo0")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(tcpAddr.IP)
}
