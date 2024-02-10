package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
)

func connectTarget(TargetHost TargetHost) {
	log.Printf("Configuring SSH client for %s...\n", TargetHost.Hostname)
	config := &ssh.ClientConfig{
		User: TargetHost.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(TargetHost.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Connecting to SSH server %s...\n", TargetHost.Hostname)
	client, err := ssh.Dial("tcp", TargetHost.Hostname+":22", config)
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

func connectJumpTarget(JumpHost JumpHost, TargetHost TargetHost) {
	log.Println("Configuring SSH client for jump host...")
	jumpConfig := &ssh.ClientConfig{
		User: JumpHost.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(JumpHost.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Connecting to jump host %s...\n", JumpHost.Hostname)
	jumpClient, err := ssh.Dial("tcp", JumpHost.Hostname+":22", jumpConfig)
	if err != nil {
		log.Fatalf("Failed to dial jump host: %s", err)
	}
	defer jumpClient.Close()
	log.Println("Connected to jump host.")

	log.Println("Configuring SSH client for target host...")
	targetConfig := &ssh.ClientConfig{
		User: TargetHost.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(TargetHost.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Establishing connection to target host %s through jump host...\n", TargetHost.Hostname)
	conn, err := jumpClient.Dial("tcp", TargetHost.Hostname+":22")
	if err != nil {
		log.Fatalf("Failed to dial target host from jump host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, TargetHost.Hostname, targetConfig)
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

func connectKubeTarget(KubeJumpHost KubeJumpHost, TargetHost TargetHost) {

	portForwardCmd, err := startPortForwarding(KubeJumpHost)
	if err != nil {
		log.Fatalf("Error starting port forwarding: %v", err)
	}
	if portForwardCmd != nil {
		defer portForwardCmd.Process.Kill()
	}

	conn, err := setupProxyDialer(KubeJumpHost.LocalPort, TargetHost.Hostname)
	if err != nil {
		log.Fatalf("Error setting up proxy dialer: %v", err)
	}

	log.Println("Setting up SSH connection...")
	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("%s:%d", TargetHost.Hostname, 22), &ssh.ClientConfig{
		User:            TargetHost.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(TargetHost.Password)},
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

func connectKubeJumpTarget(KubeJumpHost KubeJumpHost, JumpHost JumpHost, TargetHost TargetHost) {

	portForwardCmd, err := startPortForwarding(KubeJumpHost)
	if err != nil {
		log.Fatalf("Error starting port forwarding: %v", err)
	}
	if portForwardCmd != nil {
		defer portForwardCmd.Process.Kill()
	}

	conn, err := setupProxyDialer(KubeJumpHost.LocalPort, TargetHost.Hostname)
	if err != nil {
		log.Fatalf("Error setting up proxy dialer: %v", err)
	}

	jumpHostConfig := &ssh.ClientConfig{
		User: JumpHost.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(JumpHost.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Connecting to jump host %s...\n", JumpHost.Hostname)
	jumpHostSSHConn, chans, reqs, err := ssh.NewClientConn(conn, JumpHost.Hostname, jumpHostConfig)
	if err != nil {
		log.Fatalf("Failed to establish SSH connection to jump host: %s", err)
	}
	jumpHostClient := ssh.NewClient(jumpHostSSHConn, chans, reqs)
	defer jumpHostClient.Close()

	log.Printf("Connecting to target host %s via jump host...\n", TargetHost.Hostname)
	targetHostConfig := &ssh.ClientConfig{
		User: TargetHost.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(TargetHost.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	targetHostConn, err := jumpHostClient.Dial("tcp", fmt.Sprintf("%s:%d", TargetHost.Hostname, 22))
	if err != nil {
		log.Fatalf("Failed to dial target host via jump host: %s", err)
	}

	targetHostSSHConn, chans, reqs, err := ssh.NewClientConn(targetHostConn, TargetHost.Hostname, targetHostConfig)
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
