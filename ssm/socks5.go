package main

import (
	"fmt"
	"net"
	"time"
)

func isPortOpen(port int) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)), timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func waitForPortOpen(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isPortOpen(port) {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}
