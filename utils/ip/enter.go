package ip

import (
	"fmt"
	"net"
)

func IsPrivateIP(ip string) (bool, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip)
	}

	// 回环地址（IPv4 和 IPv6）
	if parsed.IsLoopback() {
		return true, nil
	}

	// 链路本地（IPv4: 169.254.0.0/16, IPv6: fe80::/10）
	if parsed.IsLinkLocalUnicast() {
		return true, nil
	}

	// 私有地址范围（10/8, 172.16/12, 192.168/16）
	if parsed.IsPrivate() {
		return true, nil
	}

	// 0.0.0.0
	if parsed.IsUnspecified() {
		return true, nil
	}

	return false, nil
}
