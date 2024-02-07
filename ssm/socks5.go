package main

import (
	"fmt"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"os"
	"os/exec"
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

func startPortForwarding(kubeconfigPath, namespace, svc string, servicePort int, localPort int) (*exec.Cmd, error) {
	if isPortOpen(localPort) {
		log.Printf("Local port %d is already open. Using the existing forwarding...\n", localPort)
		return nil, nil
	}

	log.Println("Starting port forwarding...")
	portForwardCmd := exec.Command("kubectl", "port-forward", "svc/"+svc, fmt.Sprintf("%d:%d", localPort, servicePort), "-n", namespace, "--kubeconfig", kubeconfigPath)
	portForwardCmd.Stderr = os.Stderr

	if err := portForwardCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start port-forwarding: %w", err)
	}
	log.Println("Port forwarding started.")

	if !waitForPortOpen(localPort, 10*time.Second) {
		portForwardCmd.Process.Kill()
		return nil, fmt.Errorf("timeout reached, port %d did not open", localPort)
	}

	return portForwardCmd, nil
}

func setupProxyDialer(localPort int, targetHost string) (net.Conn, error) {
	log.Println("Creating SOCKS5 dialer...")
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("localhost:%d", localPort), nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}
	log.Println("SOCKS5 dialer created.")

	log.Printf("Dialing SSH server %s via SOCKS5 proxy...\n", targetHost)
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", targetHost, 22))
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH server via SOCKS5 proxy: %w", err)
	}
	log.Println("SSH server dialed.")

	return conn, nil
}
