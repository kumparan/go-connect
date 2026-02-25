package middleware

import "testing"

func Test_isPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		// Private ranges
		{"10.0.133.121", true},   // 10.x.x.x
		{"192.168.1.50", true},   // 192.168.x.x
		{"172.16.5.10", true},    // 172.16.x.x
		{"172.31.255.255", true}, // upper bound of 172.16/12

		// Public ranges
		{"8.8.8.8", false},    // Google DNS
		{"1.1.1.1", false},    // Cloudflare DNS
		{"172.32.0.1", false}, // outside private 172.16–31
		{"11.0.0.1", false},   // not in 10.x.x.x

		// Edge cases
		{"127.0.0.1", false},   // loopback
		{"169.254.1.1", false}, // link-local
		{"invalid-ip", false},  // malformed input
	}

	for _, tt := range tests {
		result := isPrivateIP(tt.ip)
		if result != tt.expected {
			t.Errorf("isPrivateIP(%s) = %v; want %v", tt.ip, result, tt.expected)
		}
	}
}
