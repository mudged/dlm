package routineengine

import (
	"net"
	"strings"
)

// APIBaseURL builds the base URL python3 uses for loopback §3.15 calls (e.g. http://127.0.0.1:8080).
func APIBaseURL(listenAddr string) string {
	if listenAddr == "" {
		return "http://127.0.0.1:8080"
	}
	// ":8080" form
	if strings.HasPrefix(listenAddr, ":") {
		return "http://127.0.0.1" + listenAddr
	}
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return "http://127.0.0.1:" + listenAddr
	}
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}
