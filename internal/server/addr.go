package server

import (
	"net"
	"strings"
)

// LocalIPs returns IPv4 addresses of this host (excluding loopback).
// Used to show which URLs the server is reachable at on the LAN.
func LocalIPs() []string {
	var out []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			ip := ipNet.IP.To4()
			if ip != nil {
				out = append(out, ip.String())
			}
		}
	}
	return out
}

// PortFromAddr extracts the port from a listen address like ":8080" or "0.0.0.0:8080".
func PortFromAddr(addr string) string {
	if addr == "" {
		return "8080"
	}
	if idx := strings.LastIndex(addr, ":"); idx >= 0 && idx < len(addr)-1 {
		return addr[idx+1:]
	}
	return addr
}

// ServerURLs returns http://<ip>:<port> for each local IP using the given port.
func ServerURLs(port string) []string {
	if port == "" {
		port = "8080"
	}
	ips := LocalIPs()
	if len(ips) == 0 {
		return []string{}
	}
	urls := make([]string, 0, len(ips))
	for _, ip := range ips {
		urls = append(urls, "http://"+net.JoinHostPort(ip, port))
	}
	return urls
}
