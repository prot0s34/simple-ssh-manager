package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Host represents a host entry in the inventory.
type Host struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Username string `json:"username,omitempty"` // Added "omitempty" to make it optional
	Password string `json:"password"`
}

// Inventory represents the list of hosts.
type Inventory struct {
	InventoryName string `json:"inventory_name"`
	Hosts         []Host `json:"hosts"`
}

func main() {
	// Load the first inventory from the environment variable or default location
	inventory1, err := loadInventoryByIndex(1)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Load the second inventory (index 2)
	inventory2, err := loadInventoryByIndex(2)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := tview.NewApplication()

	// Create a list of hosts for the TUI
	listHostsGroupFirst := createHostList(app, inventory1, inventory1.InventoryName)
	listHostsGroupSecond := createHostList(app, inventory2, inventory2.InventoryName)

	// Quit-list for the TUI
	listQuit := tview.NewList()
	listQuit.AddItem("Quit", "Press Q to exit", 'q', func() {
		app.Stop()
	})

	flex := tview.NewFlex().
		AddItem(listHostsGroupFirst, 0, 1, true).
		AddItem(listHostsGroupSecond, 0, 1, true).
		AddItem(listQuit.SetBorder(true), 10, 1, false)

		// Define the function to connect to the selected host
	setHostListSelectedFunc(listHostsGroupFirst, inventory1, app)

	// Set the selected function for the second list
	setHostListSelectedFunc(listHostsGroupSecond, inventory2, app)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft {
			// Switch the focus to the next panel.
			app.SetFocus(listHostsGroupFirst)
		}
		if event.Key() == tcell.KeyRight {
			// Switch the focus to the next panel.
			app.SetFocus(listHostsGroupSecond)
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

}

// loadInventory loads the inventory from the specified JSON file.
func loadInventory(inventoryPath string) (*Inventory, error) {
	file, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return nil, err
	}

	var inventory Inventory
	if err := json.Unmarshal(file, &inventory); err != nil {
		return nil, err
	}

	// Set the default username to the current OS username if not defined
	currentUser, _ := user.Current()
	for i := range inventory.Hosts {
		if inventory.Hosts[i].Username == "" {
			inventory.Hosts[i].Username = currentUser.Username
		}
	}

	return &inventory, nil
}

func createHostList(app *tview.Application, inventory *Inventory, inventoryName string) *tview.List {
	list := tview.NewList()
	list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)

	for _, host := range inventory.Hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil)
	}

	list.SetBorder(true)
	list.Box.SetBorderAttributes(tcell.RuneBoard).
		SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)

	list.AddItem("", "Press 'Q' to Quit, < or > to change hosts list", 'q', func() {
		app.Stop()
	})

	return list
}

func setHostListSelectedFunc(list *tview.List, inventory *Inventory, app *tview.Application) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventory.Hosts[index]

		if host.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		// Close the TUI application
		app.Stop()

		// Launch the SSH connection in the default terminal using sshpass
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

func loadInventoryByIndex(index int) (*Inventory, error) {
	// Define a map to store inventory paths
	inventoryPaths := map[int]string{
		1: os.Getenv("SSHMANAGER_INVENTORY" + strconv.Itoa(index)),
		2: os.Getenv("SSHMANAGER_INVENTORY2" + strconv.Itoa(index)),
	}

	// Determine the inventory path based on the index
	inventoryPath, found := inventoryPaths[index]
	if !found {
		return nil, fmt.Errorf("Invalid index")
	}

	if inventoryPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		inventoryPath = usr.HomeDir + "/inventory" + strconv.Itoa(index) + ".json"
	}

	inventory, err := loadInventory(inventoryPath)
	if err != nil {
		return nil, err
	}

	return inventory, nil
}
