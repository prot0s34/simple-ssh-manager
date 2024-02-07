package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
	"log"
	"os"
	"os/exec"
	"time"
)

func executeSSHCommand(targetUsername, targetPassword, targetHost string) {
	log.Printf("Configuring SSH client for %s...\n", targetHost)

	config := &ssh.ClientConfig{
		User: targetUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(targetPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Connecting to SSH server %s...\n", targetHost)

	client, err := ssh.Dial("tcp", targetHost+":22", config)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}
	defer client.Close()
	log.Println("SSH server connection established.")

	log.Println("Creating new SSH session...")

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
	}
	defer session.Close()
	log.Println("SSH session created.")

	err = handleShell(session)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func executeSSHJumpCommand(jumpUsername, jumpPassword, jumpHost, targetUsername, targetPassword, targetHost string) {
	log.Println("Configuring SSH client for jump host...")
	jumpConfig := &ssh.ClientConfig{
		User: jumpUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(jumpPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Connecting to jump host %s...\n", jumpHost)
	jumpClient, err := ssh.Dial("tcp", jumpHost+":22", jumpConfig)
	if err != nil {
		log.Fatalf("Failed to dial jump host: %s", err)
	}
	defer jumpClient.Close()
	log.Println("Connected to jump host.")

	log.Println("Configuring SSH client for target host...")
	targetConfig := &ssh.ClientConfig{
		User: targetUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(targetPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Establishing connection to target host %s through jump host...\n", targetHost)
	conn, err := jumpClient.Dial("tcp", targetHost+":22")
	if err != nil {
		log.Fatalf("Failed to dial target host from jump host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetHost, targetConfig)
	if err != nil {
		log.Fatalf("Failed to create new SSH client connection to target host: %s", err)
	}

	targetClient := ssh.NewClient(ncc, chans, reqs)
	session, err := targetClient.NewSession()

	if err != nil {
		log.Fatalf("Failed to create session on target host: %s", err)
	}
	defer session.Close()

	err = handleShell(session)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func executeSSHKubeCommand(kubeconfigPath, namespace, podName, targetUsername, targetPassword, targetHost string) {
	localPort := 49152
	targetPort := 1080

	if isPortOpen(localPort) {
		log.Printf("Local port %d is already open. Attempting to use the existing forwarding...\n", localPort)
	} else {
		log.Println("Starting port forwarding...")
		portForwardCmd := exec.Command("kubectl", "port-forward", "svc/dante", fmt.Sprintf("%d:%d", localPort, targetPort), "-n", namespace, "--kubeconfig", kubeconfigPath)
		portForwardCmd.Stderr = os.Stderr

		if err := portForwardCmd.Start(); err != nil {
			log.Fatalf("Failed to start port-forwarding: %s", err)
		}
		log.Println("Port forwarding started.")

		defer func() {
			log.Println("Terminating port forwarding...")
			if err := portForwardCmd.Process.Kill(); err != nil {
				log.Printf("Failed to kill port-forwarding process: %s", err)
			}
			log.Println("Port forwarding terminated.")
		}()
	}

	time.Sleep(2 * time.Second)

	log.Println("Creating SOCKS5 dialer...")
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("localhost:%d", localPort), nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to create SOCKS5 dialer: %s", err)
	}
	log.Println("SOCKS5 dialer created.")

	log.Printf("Dialing SSH server %s via SOCKS5 proxy...\n", targetHost)
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", targetHost, 22))
	if err != nil {
		log.Fatalf("Failed to dial SSH server via SOCKS5 proxy: %s", err)
	}
	log.Println("SSH server dialed.")

	log.Println("Setting up SSH connection...")
	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("%s:%d", targetHost, 22), &ssh.ClientConfig{
		User:            targetUsername,
		Auth:            []ssh.AuthMethod{ssh.Password(targetPassword)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		log.Fatalf("Failed to create SSH client connection: %s", err)
	}
	client := ssh.NewClient(ncc, chans, reqs)
	log.Println("SSH connection established.")
	defer client.Close()

	log.Println("Starting SSH session...")
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
	}
	defer session.Close()

	err = handleShell(session)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func executeSSHKubeJumpCommand(kubeconfigPath, namespace, podName, jumpHost, jumpUsername, jumpPassword, targetUsername, targetPassword, targetHost string) {
	localPort := 49152
	targetPort := 1080

	if isPortOpen(localPort) {
		log.Printf("Local port %d is already open. Attempting to use the existing forwarding...\n", localPort)
	} else {
		log.Println("Starting port forwarding...")
		portForwardCmd := exec.Command("kubectl", "port-forward", "svc/dante", fmt.Sprintf("%d:%d", localPort, targetPort), "-n", namespace, "--kubeconfig", kubeconfigPath)
		portForwardCmd.Stderr = os.Stderr

		if err := portForwardCmd.Start(); err != nil {
			log.Fatalf("Failed to start port-forwarding: %s", err)
		}
		log.Println("Port forwarding started.")

		defer func() {
			log.Println("Terminating port forwarding...")
			if err := portForwardCmd.Process.Kill(); err != nil {
				log.Printf("Failed to kill port-forwarding process: %s", err)
			}
			log.Println("Port forwarding terminated.")
		}()
	}

	time.Sleep(2 * time.Second)

	log.Println("Creating SOCKS5 dialer...")
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("localhost:%d", localPort), nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to create SOCKS5 dialer: %s", err)
	}
	log.Println("SOCKS5 dialer created.")

	log.Printf("Connecting to jump host %s...\n", jumpHost)
	jumpHostConfig := &ssh.ClientConfig{
		User: jumpUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(jumpPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	jumpHostConn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", jumpHost, 22))
	if err != nil {
		log.Fatalf("Failed to dial jump host: %s", err)
	}

	jumpHostSSHConn, chans, reqs, err := ssh.NewClientConn(jumpHostConn, jumpHost, jumpHostConfig)
	if err != nil {
		log.Fatalf("Failed to establish SSH connection to jump host: %s", err)
	}
	jumpHostClient := ssh.NewClient(jumpHostSSHConn, chans, reqs)
	defer jumpHostClient.Close()

	log.Printf("Connecting to target host %s via jump host...\n", targetHost)
	targetHostConfig := &ssh.ClientConfig{
		User: targetUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(targetPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	targetHostConn, err := jumpHostClient.Dial("tcp", fmt.Sprintf("%s:%d", targetHost, 22))
	if err != nil {
		log.Fatalf("Failed to dial target host via jump host: %s", err)
	}

	targetHostSSHConn, chans, reqs, err := ssh.NewClientConn(targetHostConn, targetHost, targetHostConfig)
	if err != nil {
		log.Fatalf("Failed to establish SSH connection to target host: %s", err)
	}
	targetHostClient := ssh.NewClient(targetHostSSHConn, chans, reqs)
	defer targetHostClient.Close()

	log.Println("Starting SSH session on target host...")
	session, err := targetHostClient.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session on target host: %s", err)
	}
	defer session.Close()

	err = handleShell(session)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
