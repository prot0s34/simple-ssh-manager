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

	list := createHostList(app, inventoryGroups[inventoryIndex].Hosts, inventoryGroups[inventoryIndex].Name)

	setHostSelected(list, inventoryGroups[inventoryIndex].Hosts, app, inventoryGroups)

	switchHostList(app, &inventoryIndex, inventoryGroups, list)

	if err := app.SetRoot(list, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
