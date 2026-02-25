package internal

import "net"

var privateBlocks = []*net.IPNet{mustCIDR("10.0.0.0/8"), mustCIDR("172.16.0.0/12"), mustCIDR("192.168.0.0/16")}

// IsPrivateIP checks if the given IP is in defined privateBlocks
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func mustCIDR(cidr string) *net.IPNet {
	_, block, _ := net.ParseCIDR(cidr)
	return block
}
