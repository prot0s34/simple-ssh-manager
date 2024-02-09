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

	inventoryGroups, err := loadInventoryGroups()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	ctx := &AppContext{
		InventoryIndex: &inventoryIndex,
		Inventory:      inventoryGroups,
	}

	app := tview.NewApplication()

	list := createHostList(app, ctx)

	switchHostList(app, list, ctx)

	setHostSelected(app, list, ctx)

	if err := app.SetRoot(list, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
