package socks

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const path = "/tmp/proxy.sock"

type Config struct {
	// Environment variables provided by openconnect
	IP4InternalAddress string
	IP4InternalMTU     string
	VPNFD              uintptr
	IP4InternalDNS     []string
	Reason             string
}

func GatherEnv() *Config {
	cfg := &Config{}
	cfg.Reason = os.Getenv("reason")
	cfg.IP4InternalAddress = os.Getenv("INTERNAL_IP4_ADDRESS")
	cfg.IP4InternalMTU = os.Getenv("INTERNAL_IP4_MTU")
	dnses := os.Getenv("INTERNAL_IP4_DNS") // space sparated values
	if len(dnses) > 0 {
		cfg.IP4InternalDNS = strings.Split(dnses, " ")
	}

	svpnfd := os.Getenv("VPNFD")
	if len(svpnfd) > 0 {
		ivpnfd, err := strconv.Atoi(svpnfd)
		if err != nil {
			log.Fatal("invalid fd", err)
		}
		cfg.VPNFD = uintptr(ivpnfd)
	}
	return cfg
}
