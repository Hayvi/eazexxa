package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request, trustProxyHops int) string {
	if trustProxyHops > 0 {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				idx := len(ips) - trustProxyHops
				if idx < 0 {
					idx = 0
				}
				ip := strings.TrimSpace(ips[idx])
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}

		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			if net.ParseIP(xri) != nil {
				return xri
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
