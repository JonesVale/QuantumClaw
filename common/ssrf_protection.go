package common

import (
	"net"
	"strings"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

var blockedIPRanges = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
}

func IsSSRFBlocked(host string) bool {
	host = strings.TrimSpace(host)
	host = strings.Split(host, ":")[0]
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, cidr := range blockedIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			logger.SysLog("SSRF blocked: " + host)
			return true
		}
	}
	return false
}
