package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
	"golang.org/x/term"
	"log"
	"os"
	"os/exec"
	"time"
)

func updateHostList(app *tview.Application, list *tview.List, hosts []Host, inventoryName string) {
	for _, host := range hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
		list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)
		list.Box.SetBorderAttributes(tcell.RuneBoard)
	}
	list.AddItem("", "'Q' to Quit, hjkl or Arrow Keys to navigate, 'Enter' to connect", 'q', func() {
		app.Stop()
	})
}

func createHostList(app *tview.Application, hosts []Host, inventoryName string) *tview.List {
	list := tview.NewList()

	for _, host := range hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
		list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)
		list.Box.SetBorderAttributes(tcell.RuneBoard)
	}

	list.AddItem("", "'Q' to Quit, hjkl or Arrow Keys to navigate, 'Enter' to connect", 'q', func() {
		app.Stop()
	})

	return list
}

func switchHostList(app *tview.Application, inventoryIndex *int, inventoryGroups []InventoryGroup, list *tview.List) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if inModalDialog {
			return event
		}

		if event.Key() == tcell.KeyLeft || event.Rune() == 'h' {
			*inventoryIndex = (*inventoryIndex - 1 + len(inventoryGroups)) % len(inventoryGroups)
			list.Clear()
			updateHostList(app, list, inventoryGroups[*inventoryIndex].Hosts, inventoryGroups[*inventoryIndex].Name)
		} else if event.Key() == tcell.KeyRight || event.Rune() == 'l' {
			*inventoryIndex = (*inventoryIndex + 1) % len(inventoryGroups)
			list.Clear()
			updateHostList(app, list, inventoryGroups[*inventoryIndex].Hosts, inventoryGroups[*inventoryIndex].Name)
		}
		if event.Rune() == 'j' && list.GetCurrentItem() < list.GetItemCount()-1 {
			list.SetCurrentItem(list.GetCurrentItem() + 1)
		} else if event.Rune() == 'k' && list.GetCurrentItem() > 0 {
			list.SetCurrentItem(list.GetCurrentItem() - 1)
		}
		return event
	})
}

func setHostSelected(list *tview.List, hosts []Host, app *tview.Application, inventoryGroups []InventoryGroup) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventoryGroups[inventoryIndex].Hosts[index]

		if host.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		inModalDialog = true
		dialog := tview.NewModal().
			SetText("Choose a jumphost option for host: " + host.Name).
			AddButtons([]string{"None", "Kube❯Jump", "Kube", "Jump", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				inModalDialog = false
				switch buttonIndex {
				case 0: // None
					app.Stop()
					executeSSHCommand(host.Username, host.Password, host.Hostname)
				case 1: // Kube + Jump
					app.Stop()
					podName, err := findJumpPod(inventoryGroups, inventoryIndex)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					jumpHost := inventoryGroups[inventoryIndex].JumpHostConfig.Hostname
					jumpHostUsername := inventoryGroups[inventoryIndex].JumpHostConfig.Username
					jumpHostPassword := inventoryGroups[inventoryIndex].JumpHostConfig.Password
					kubeconfig := inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath
					namespace := inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace
					executeSSHKubeJumpCommand(kubeconfig, namespace, podName, jumpHost, jumpHostUsername, jumpHostPassword, host.Username, host.Password, host.Hostname)
				case 2: // Kube
					app.Stop()
					podName, err := findJumpPod(inventoryGroups, inventoryIndex)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					kubeconfig := inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath
					namespace := inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace
					executeSSHKubeCommand(kubeconfig, namespace, podName, host.Username, host.Password, host.Hostname)
				case 3: // Jump
					app.Stop()
					jumpHost := inventoryGroups[inventoryIndex].JumpHostConfig.Hostname
					executeSSHCommand(host.Username, inventoryGroups[inventoryIndex].JumpHostConfig.Password, jumpHost)
				case 4: // Cancel
					inModalDialog = false
					app.SetRoot(list, true)
				}
			})

		app.SetRoot(dialog, true)
		switchHostList(app, &inventoryIndex, inventoryGroups, list)
	})
}

func findJumpPod(inventoryGroups []InventoryGroup, inventoryIndex int) (string, error) {
	if inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath == "" {
		return "", fmt.Errorf("error[initializeKubeJumpHostConfig]: KubeconfigPath is missing in the inventory")
	}

	clientset, err := initKubernetesClient(inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("error[initializeKubeJumpHostConfig]: initializing Kubernetes client: %v", err)
	}

	podName := inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName
	if podName == "" {
		podName, err = findPodByKeyword(clientset, inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace, inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodNameTemplate)
		if err != nil {
			return "", fmt.Errorf("error[initializeKubeJumpHostConfig]: %v", err)
		}
		inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName = podName
	}

	return podName, nil
}

func executeSSHCommand(username, password, hostname string) {
	log.Printf("Configuring SSH client for %s...\n", hostname)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Connecting to SSH server %s...\n", hostname)

	client, err := ssh.Dial("tcp", hostname+":22", config)
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

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to set terminal to raw mode: %s", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatalf("Failed to restore terminal: %s", err)
		}
	}()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	log.Println("Requesting pseudo terminal...")
	if err := session.RequestPty("xterm", 40, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		log.Fatalf("Failed to request pseudo terminal: %s", err)
	}

	log.Println("Starting shell...")

	if err := session.Shell(); err != nil {
		log.Fatalf("Failed to start shell: %s", err)
	}

	log.Println("Waiting for SSH session to finish...")
	if err := session.Wait(); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			if exitErr.ExitStatus() == 0 {
				log.Println("SSH session finished successfully.")
			} else {
				log.Printf("SSH session exited with non-zero status: %d\n", exitErr.ExitStatus())
			}
		} else {
			log.Fatalf("Failed to wait for session completion: %s", err)
		}
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
	log.Println("SSH session to target host established.")
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to set terminal to raw mode: %s", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatalf("Failed to restore terminal: %s", err)
		}
	}()
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	log.Println("Requesting pseudo terminal on target host...")
	if err := session.RequestPty("xterm", 40, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		log.Fatalf("Failed to request pseudo terminal on target host: %s", err)
	}

	log.Println("Starting shell on target host...")
	if err := session.Shell(); err != nil {
		log.Fatalf("Failed to start shell on target host: %s", err)
	}

	log.Println("Waiting for the session on target host to finish...")
	if err := session.Wait(); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			if exitErr.ExitStatus() == 0 {
				log.Println("Session on target host finished successfully.")
			} else {
				log.Printf("SSH session on target host exited with non-zero status: %d\n", exitErr.ExitStatus())
			}
		} else {
			log.Fatalf("Failed to wait for session completion on target host: %s", err)
		}
	}
}

func executeSSHKubeCommand(kubeconfigPath, namespace, podName, targetUsername, targetPassword, targetHost string) {
	localPort := 49152
	targetPort := 1080

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
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to set terminal to raw mode: %s", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatalf("Failed to restore terminal: %s", err)
		}
	}()
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.RequestPty("xterm", 40, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		log.Fatalf("Failed to request pseudo terminal: %s", err)
	}

	if err := session.Shell(); err != nil {
		log.Fatalf("Failed to start shell: %s", err)
	}
	log.Println("Shell started.")

	if err := session.Wait(); err != nil {
		log.Printf("SSH session finished with error: %s", err)
	}
}

func executeSSHKubeJumpCommand(kubeconfigPath, namespace, podName, jumpHost, jumpHostUser, jumpHostPass, targetUsername, targetPassword, targetHost string) {
	localPort := 49152
	targetPort := 1080

	log.Println("Starting port forwarding...")
	portForwardCmd := exec.Command("kubectl", "port-forward", "svc/dante", fmt.Sprintf("%d:%d", localPort, targetPort), "-n", namespace, "--kubeconfig", kubeconfigPath)
	portForwardCmd.Stderr = os.Stderr

	if err := portForwardCmd.Start(); err != nil {
		log.Fatalf("Failed to start port-forwarding: %s", err)
	}
	log.Println("Port forwarding started.")

	defer func() {
		fmt.Print("\033[0m")
		log.Println("Terminating port forwarding...")
		if err := portForwardCmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill port-forwarding process: %s", err)
		}
		fmt.Print("\033[0m")
		log.Println("Port forwarding terminated.")
	}()

	time.Sleep(2 * time.Second)

	log.Println("Creating SOCKS5 dialer...")
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("localhost:%d", localPort), nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to create SOCKS5 dialer: %s", err)
	}
	log.Println("SOCKS5 dialer created.")

	log.Printf("Connecting to jump host %s...\n", jumpHost)
	jumpHostConfig := &ssh.ClientConfig{
		User: jumpHostUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(jumpHostPass),
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
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to set terminal to raw mode: %s", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatalf("Failed to restore terminal: %s", err)
		}
	}()
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.RequestPty("xterm", 40, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		log.Fatalf("Failed to request pseudo terminal on target host: %s", err)
	}

	if err := session.Shell(); err != nil {
		log.Fatalf("Failed to start shell on target host: %s", err)
	}
	log.Println("Shell started on target host.")

	if err := session.Wait(); err != nil {
		log.Printf("SSH session on target host finished with error: %s", err)
	}
}
