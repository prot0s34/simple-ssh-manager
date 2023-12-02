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

func setHostListSelectedFunc(list *tview.List, hosts []Host, app *tview.Application, inventoryGroups []InventoryGroup, listHostsGroup *tview.List) {
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
					cmd := exec.Command("sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", host.Username+"@"+host.Hostname)
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 1: // Kube + Jump
					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath == "" {
						fmt.Println("Error: KubeconfigPath is missing in the inventory.")
						os.Exit(1)
					}

					clientset, err := initKubernetesClient(inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath)
					if err != nil {
						fmt.Println("Error initializing Kubernetes client:", err)
						os.Exit(1)
					}

					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName == "" {
						podName, err := findPodByKeyword(clientset, inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace, inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodNameTemplate)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName = podName
					}

					app.Stop()

					cmd := exec.Command("kubectl", "--kubeconfig", inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath, "exec", "-it", inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName, "--", "sshpass", "-p", inventoryGroups[inventoryIndex].JumpHostConfig.Password, "ssh", "-o", "StrictHostKeyChecking no", "-q", "-t", inventoryGroups[inventoryIndex].JumpHostConfig.Username+"@"+inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "sshpass", "-p", "'"+host.Password+"'", "ssh", "-o", "'StrictHostKeyChecking no'", "-q", host.Username+"@"+host.Hostname)

					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 2: // Kube
					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath == "" {
						fmt.Println("Error: KubeconfigPath is missing in the inventory.")
						os.Exit(1)
					}

					clientset, err := initKubernetesClient(inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath)
					if err != nil {
						fmt.Println("Error initializing Kubernetes client:", err)
						os.Exit(1)
					}

					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName == "" {
						podName, err := findPodByKeyword(clientset, inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace, inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodNameTemplate)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName = podName
					}

					app.Stop()
					cmd := exec.Command("kubectl", "--kubeconfig", inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath, "exec", "-it", inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName, "--", "sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", "-q", host.Username+"@"+host.Hostname)

					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 3: // Jump
					app.Stop()
					cmd := exec.Command("sshpass", "-p", inventoryGroups[inventoryIndex].JumpHostConfig.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", "-q", inventoryGroups[inventoryIndex].JumpHostConfig.Username+"@"+inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "sshpass", "-p", host.Password, "ssh", "-o", "'StrictHostKeyChecking no'", "-t", "-q", host.Username+"@"+host.Hostname)
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 4: // Cancel
					inModalDialog = false
					app.SetRoot(listHostsGroup, true)
				}
			})

		app.SetRoot(dialog, true)
		navigateBetweenInventoryGroups(app, &inventoryIndex, inventoryGroups, listHostsGroup)
	})
}

func navigateBetweenInventoryGroups(app *tview.Application, inventoryIndex *int, inventoryGroups []InventoryGroup, listHostsGroup *tview.List) {
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
