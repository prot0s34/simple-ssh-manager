package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func assembleHostList(app *tview.Application, list *tview.List, ctx *AppContext, createNew bool) *tview.List {
	if createNew {
		list = tview.NewList()
	} else {
		list.Clear()
	}

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
			assembleHostList(app, list, ctx, false)
		} else if event.Key() == tcell.KeyRight || event.Rune() == 'l' {
			*ctx.InventoryIndex = (*ctx.InventoryIndex + 1) % len(ctx.Inventory)
			list.Clear()
			assembleHostList(app, list, ctx, false)
		}

		if event.Rune() == 'j' && list.GetCurrentItem() < list.GetItemCount()-1 {
			list.SetCurrentItem(list.GetCurrentItem() + 1)
		} else if event.Rune() == 'k' && list.GetCurrentItem() > 0 {
			list.SetCurrentItem(list.GetCurrentItem() - 1)
		}

		return event
	})
}

func selectListItem(app *tview.Application, list *tview.List, ctx *AppContext) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {

		TargetHost := ctx.Inventory[*ctx.InventoryIndex].TargetHost[index]
		JumpHost := ctx.Inventory[*ctx.InventoryIndex].JumpHost
		KubeJumpHost := ctx.Inventory[*ctx.InventoryIndex].KubeJumpHost
		showModalListItem(app, list, TargetHost, JumpHost, KubeJumpHost)
		switchHostList(app, list, ctx)
	})
}

func showModalListItem(app *tview.Application, list *tview.List, TargetHost TargetHost, JumpHost JumpHost, KubeJumpHost KubeJumpHost) {
	inModalDialog = true
	dialog := tview.NewModal().
		SetText("Choose a jumphost option for host: " + TargetHost.Name).
		AddButtons([]string{"None", "Kube❯Jump", "Kube", "Jump", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			inModalDialog = false
			switch buttonIndex {
			case 0: // None
				app.Stop()
				connectTarget(TargetHost)
			case 1: // Kube + Jump
				app.Stop()
				connectKubeJumpTarget(KubeJumpHost, JumpHost, TargetHost)
			case 2: // Kube
				app.Stop()
				connectKubeTarget(KubeJumpHost, TargetHost)
			case 3: // Jump
				app.Stop()
				connectJumpTarget(JumpHost, TargetHost)
			case 4: // Cancel
				inModalDialog = false
				app.SetRoot(list, true)
			}
		})

	app.SetRoot(dialog, true)
}
