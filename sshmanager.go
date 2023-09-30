package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

	"github.com/gdamore/tcell"
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
	Hosts []Host `json:"hosts"`
}

func main() {
	// Load the inventory from the environment variable or default location
	inventoryPath := os.Getenv("SSHMANAGER_INVENTORY")
	if inventoryPath == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		inventoryPath = usr.HomeDir + "/inventory.json"
	}

	inventory, err := loadInventory(inventoryPath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := tview.NewApplication()

	// Create a list of hosts for the TUI
	listHostsGroupFirst := tview.NewList()
	for _, host := range inventory.Hosts {
		listHostsGroupFirst.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil)
	}
	listHostsGroupFirst.SetBorder(true)
	listHostsGroupFirst.Box.SetBorderAttributes(tcell.RuneBoard).
		SetTitle("[black:red]S[:yellow]i[:green]m[:darkcyan]l[:blue]e[:darkmagenta]-[:red]S[:yellow]S[:green]H[:darkmagenta]-[:blue]M[:red]a[:darkcyan]n[:yellow]a[:yellow]g[:red]e[:blue]r[white:-]").SetTitleAlign(tview.AlignLeft)
	listHostsGroupFirst.AddItem("", "", 'q', func() {
		app.Stop()
	})

	listQuit := tview.NewList()
	listQuit.SetBorder(true)
	listQuit.AddItem("Quit", "Press Q to exit", 'q', func() {
		app.Stop()
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(listHostsGroupFirst, 0, 1, true).
		AddItem(listQuit, 3, 2, false)

	// Define the function to connect to the selected host
	listHostsGroupFirst.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
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
