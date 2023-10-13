package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Host struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Username string `json:"username,omitempty"`
	Password string `json:"password"`
}

type Inventory struct {
	InventoryGroup1 []Host `json:"inventory_group1"`
	InventoryGroup2 []Host `json:"inventory_group2"`
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	inventory, err := loadInventory()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := tview.NewApplication()

	listHostsGroupFirst := createHostList(app, inventory.InventoryGroup1, "inventory_group1")
	listHostsGroupSecond := createHostList(app, inventory.InventoryGroup2, "inventory_group2")

	flex := tview.NewFlex().
		AddItem(listHostsGroupFirst, 0, 1, true).
		AddItem(listHostsGroupSecond, 0, 1, true)

	setHostListSelectedFunc(listHostsGroupFirst, inventory.InventoryGroup1, app)
	setHostListSelectedFunc(listHostsGroupSecond, inventory.InventoryGroup2, app)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft {

			app.SetFocus(listHostsGroupFirst)
		}
		if event.Key() == tcell.KeyRight {

			app.SetFocus(listHostsGroupSecond)
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

}

func loadInventory() (*Inventory, error) {
	inventoryPath := os.Getenv("SSHMANAGER_INVENTORY")

	if inventoryPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		inventoryPath = usr.HomeDir + "/inventory.json"
	}

	file, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return nil, err
	}

	var inventory Inventory
	if err := json.Unmarshal(file, &inventory); err != nil {
		return nil, err
	}

	currentUser, _ := user.Current()
	for i := range inventory.InventoryGroup1 {
		if inventory.InventoryGroup1[i].Username == "" {
			inventory.InventoryGroup1[i].Username = currentUser.Username
		}
	}
	for i := range inventory.InventoryGroup2 {
		if inventory.InventoryGroup2[i].Username == "" {
			inventory.InventoryGroup2[i].Username = currentUser.Username
		}
	}

	return &inventory, nil
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

func setHostListSelectedFunc(list *tview.List, hosts []Host, app *tview.Application) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := hosts[index]

		if host.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		app.Stop()

		cmd := exec.Command("sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", host.Username+"@"+host.Hostname)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	})
}
