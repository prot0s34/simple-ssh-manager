package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func updateHostList(app *tview.Application, list *tview.List, hosts []Host, inventoryName string) {
	for _, host := range hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
		list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)
		list.Box.SetBorderAttributes(tcell.RuneBoard)
	}
	list.AddItem("", "'Q' to Quit, < or > to change host list, 'Enter' to connect", 'q', func() {
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

	list.AddItem("", "'Q' to Quit, < or > to change host list, 'Enter' to connect", 'q', func() {
		app.Stop()
	})

	return list
}

func switchHostList(app *tview.Application, inventoryIndex *int, inventoryGroups []InventoryGroup, listHostsGroup *tview.List) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if inModalDialog {
			return event
		}

		if event.Key() == tcell.KeyLeft {
			*inventoryIndex = (*inventoryIndex - 1 + len(inventoryGroups)) % len(inventoryGroups)
			listHostsGroup.Clear()
			updateHostList(app, listHostsGroup, inventoryGroups[*inventoryIndex].Hosts, inventoryGroups[*inventoryIndex].Name)
		} else if event.Key() == tcell.KeyRight {
			*inventoryIndex = (*inventoryIndex + 1) % len(inventoryGroups)
			listHostsGroup.Clear()
			updateHostList(app, listHostsGroup, inventoryGroups[*inventoryIndex].Hosts, inventoryGroups[*inventoryIndex].Name)
		}

		return event
	})
}

func setHostSelected(list *tview.List, hosts []Host, app *tview.Application, inventoryGroups []InventoryGroup, listHostsGroup *tview.List) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventoryGroups[inventoryIndex].Hosts[index]

		if host.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		inModalDialog = true
		kubectlArgs := []string{
			"kubectl",
			"--kubeconfig",
			inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath,
			"-n",
			inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace,
			"exec",
			"-it",
		}

		dialog := tview.NewModal().
			SetText("Choose a jumphost option for host: " + host.Name).
			AddButtons([]string{"None", "Kube❯Jump", "Kube", "Jump", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				inModalDialog = false
				switch buttonIndex {
				case 0: // None
					app.Stop()
					infoArgs := []string{"echo", "-e", "\\033[31mConnection: ->\\033[0m", host.Hostname}
					connectionArgs := []string{"sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", "-q", host.Username + "@" + host.Hostname}
					executeCommand(infoArgs, connectionArgs)
				case 1: // Kube + Jump
					app.Stop()
					podName, err := initializeKubeJumpHostConfig(inventoryGroups, inventoryIndex)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					infoArgs := []string{"echo", "-e", "\\033[31mConnection: ->\\033[0m", podName, "\\033[31m->\\033[0m", inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "\\033[31m->\\033[0m", host.Hostname}
					connectionArgs := append(kubectlArgs, podName, "--", "sshpass", "-p", inventoryGroups[inventoryIndex].JumpHostConfig.Password, "ssh", "-o", "StrictHostKeyChecking no", "-q", "-t", inventoryGroups[inventoryIndex].JumpHostConfig.Username+"@"+inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "sshpass", "-p", "'"+host.Password+"'", "ssh", "-o", "'StrictHostKeyChecking no'", "-q", host.Username+"@"+host.Hostname)
					executeCommand(infoArgs, connectionArgs)
				case 2: // Kube
					app.Stop()
					podName, err := initializeKubeJumpHostConfig(inventoryGroups, inventoryIndex)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					infoArgs := []string{"echo", "-e", "\\033[31mConnection: ->\\033[0m", podName, "\\033[31m->\\033[0m", host.Hostname}
					connectionArgs := append(kubectlArgs, podName, "--", "sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", "-q", host.Username+"@"+host.Hostname)
					executeCommand(infoArgs, connectionArgs)
				case 3: // Jump
					app.Stop()
					infoArgs := []string{"echo", "-e", "\\033[31mConnection: ->\\033[0m", inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "\\033[31m->\\033[0m", host.Hostname}
					connectionArgs := []string{"sshpass", "-p", inventoryGroups[inventoryIndex].JumpHostConfig.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", "-q", inventoryGroups[inventoryIndex].JumpHostConfig.Username + "@" + inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "sshpass", "-p", host.Password, "ssh", "-o", "'StrictHostKeyChecking no'", "-t", "-q", host.Username + "@" + host.Hostname}
					executeCommand(infoArgs, connectionArgs)
				case 4: // Cancel
					inModalDialog = false
					app.SetRoot(listHostsGroup, true)
				}
			})

		app.SetRoot(dialog, true)
		switchHostList(app, &inventoryIndex, inventoryGroups, listHostsGroup)
	})
}

func initializeKubeJumpHostConfig(inventoryGroups []InventoryGroup, inventoryIndex int) (string, error) {
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

func executeCommand(infoArgs []string, connectionArgs []string) {
	cmdInfo := exec.Command(infoArgs[0], infoArgs[1:]...)
	cmdInfo.Stdout = os.Stdout
	cmdInfo.Stdin = os.Stdin
	cmdInfo.Stderr = os.Stderr

	if err := cmdInfo.Run(); err != nil {
		fmt.Println("Connection info print error:", err)
		os.Exit(1)
	}

	cmdConnection := exec.Command(connectionArgs[0], connectionArgs[1:]...)
	cmdConnection.Stdout = os.Stdout
	cmdConnection.Stdin = os.Stdin
	cmdConnection.Stderr = os.Stderr

	if err := cmdConnection.Run(); err != nil {
		fmt.Println("Connection command execute error:", err)
		os.Exit(1)
	}
}
