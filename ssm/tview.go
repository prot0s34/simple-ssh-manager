package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func updateHostList(app *tview.Application, list *tview.List, hosts []TargetHost, inventoryName string) {
	for _, host := range hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
		list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)
		list.Box.SetBorderAttributes(tcell.RuneBoard)
	}
	list.AddItem("", "'Q' to Quit, hjkl or ← ↓ ↑ → to navigate, 'Enter' to connect", 'q', func() {
		app.Stop()
	})
}

func createHostList(app *tview.Application, ctx *AppContext) *tview.List {
	list := tview.NewList()

	selectedInventoryGroup := ctx.Inventory[*ctx.InventoryIndex]

	for _, host := range selectedInventoryGroup.TargetHost {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
	}

	list.SetTitle("[black:darkcyan]" + selectedInventoryGroup.Name + "[white:-]").SetTitleAlign(tview.AlignLeft)
	list.Box.SetBorderAttributes(tcell.RuneBoard)

	list.AddItem("", "'Q' to Quit, hjkl or ← ↓ ↑ → to navigate, 'Enter' to connect", 'q', func() {
		app.Stop()
	})

	return list
}

func switchHostList(app *tview.Application, list *tview.List, ctx *AppContext) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if inModalDialog {
			return event
		}

		if event.Key() == tcell.KeyLeft || event.Rune() == 'h' {
			*ctx.InventoryIndex = (*ctx.InventoryIndex - 1 + len(ctx.Inventory)) % len(ctx.Inventory)
			list.Clear()
			updateHostList(app, list, ctx.Inventory[*ctx.InventoryIndex].TargetHost, ctx.Inventory[*ctx.InventoryIndex].Name)
		} else if event.Key() == tcell.KeyRight || event.Rune() == 'l' {
			*ctx.InventoryIndex = (*ctx.InventoryIndex + 1) % len(ctx.Inventory)
			list.Clear()
			updateHostList(app, list, ctx.Inventory[*ctx.InventoryIndex].TargetHost, ctx.Inventory[*ctx.InventoryIndex].Name)
		}

		if event.Rune() == 'j' && list.GetCurrentItem() < list.GetItemCount()-1 {
			list.SetCurrentItem(list.GetCurrentItem() + 1)
		} else if event.Rune() == 'k' && list.GetCurrentItem() > 0 {
			list.SetCurrentItem(list.GetCurrentItem() - 1)
		}

		return event
	})
}

func setHostSelected(app *tview.Application, list *tview.List, ctx *AppContext) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {

		TargetHost := ctx.Inventory[*ctx.InventoryIndex].TargetHost[index]
		JumpHost := ctx.Inventory[*ctx.InventoryIndex].JumpHost
		KubeJumpHost := ctx.Inventory[*ctx.InventoryIndex].KubeJumpHost
		if TargetHost.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		inModalDialog = true
		dialog := tview.NewModal().
			SetText("Choose a jumphost option for host: " + TargetHost.Name).
			AddButtons([]string{"None", "Kube❯Jump", "Kube", "Jump", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				inModalDialog = false
				switch buttonIndex {
				case 0: // None
					app.Stop()
					executeSSHCommand(TargetHost.Username, TargetHost.Password, TargetHost.Hostname)
				case 1: // Kube + Jump
					app.Stop()
					jumpHost := JumpHost.Hostname
					jumpHostUsername := JumpHost.Username
					jumpHostPassword := JumpHost.Password
					kubeconfig := KubeJumpHost.KubeconfigPath
					namespace := KubeJumpHost.Namespace
					svc := KubeJumpHost.Service
					servicePort := KubeJumpHost.ServicePort
					localPort := KubeJumpHost.LocalPort
					executeSSHKubeJumpCommand(kubeconfig, namespace, svc, servicePort, localPort, jumpHost, jumpHostUsername, jumpHostPassword, TargetHost.Username, TargetHost.Password, TargetHost.Hostname)
				case 2: // Kube
					app.Stop()
					kubeconfig := KubeJumpHost.KubeconfigPath
					namespace := KubeJumpHost.Namespace
					svc := KubeJumpHost.Service
					servicePort := KubeJumpHost.ServicePort
					localPort := KubeJumpHost.LocalPort
					executeSSHKubeCommand(kubeconfig, namespace, svc, servicePort, localPort, TargetHost.Username, TargetHost.Password, TargetHost.Hostname)
				case 3: // Jump
					// fix jump option behavior
					app.Stop()
					jumpHost := JumpHost.Hostname
					// jumpHostUsername := inventory[inventoryIndex].JumpHost.Username
					// jumpHostPassword := inventory[inventoryIndex].JumpHost.Password
					executeSSHCommand(TargetHost.Username, JumpHost.Password, jumpHost)
				case 4: // Cancel
					inModalDialog = false
					app.SetRoot(list, true)
				}
			})

		app.SetRoot(dialog, true)
		switchHostList(app, list, ctx)
	})
}
