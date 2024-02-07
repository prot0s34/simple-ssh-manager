package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
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
					jumpHost := inventoryGroups[inventoryIndex].JumpHostConfig.Hostname
					jumpHostUsername := inventoryGroups[inventoryIndex].JumpHostConfig.Username
					jumpHostPassword := inventoryGroups[inventoryIndex].JumpHostConfig.Password
					kubeconfig := inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath
					namespace := inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace
					svc := inventoryGroups[inventoryIndex].KubeJumpHostConfig.Service
					executeSSHKubeJumpCommand(kubeconfig, namespace, svc, jumpHost, jumpHostUsername, jumpHostPassword, host.Username, host.Password, host.Hostname)
				case 2: // Kube
					app.Stop()
					kubeconfig := inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath
					namespace := inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace
					svc := inventoryGroups[inventoryIndex].KubeJumpHostConfig.Service
					executeSSHKubeCommand(kubeconfig, namespace, svc, host.Username, host.Password, host.Hostname)
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
