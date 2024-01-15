package main

import (
	"fmt"
	"os"

	"github.com/rivo/tview"
)

var inModalDialog = false
var inventoryIndex = 0

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	app := tview.NewApplication()

	inventoryGroups, err := loadInventoryGroups()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	inventoryIndex = 0

	listHostsGroup := createHostList(app, inventoryGroups[inventoryIndex].Hosts, inventoryGroups[inventoryIndex].Name)

	setHostListSelected(listHostsGroup, inventoryGroups[inventoryIndex].Hosts, app, inventoryGroups, listHostsGroup)

	navigateBetweenInventoryGroups(app, &inventoryIndex, inventoryGroups, listHostsGroup)

	if err := app.SetRoot(listHostsGroup, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
