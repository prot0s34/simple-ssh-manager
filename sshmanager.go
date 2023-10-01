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
	// Load the first inventory from the environment variable or default location
	inventoryPath1 := os.Getenv("SSHMANAGER_INVENTORY1")
	if inventoryPath1 == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		inventoryPath1 = usr.HomeDir + "/inventory1.json"
	}
	inventory1, err := loadInventory(inventoryPath1)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Load the second inventory from the environment variable or default location
	inventoryPath2 := os.Getenv("SSHMANAGER_INVENTORY2")
	if inventoryPath2 == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		inventoryPath2 = usr.HomeDir + "/inventory2.json"
	}
	inventory2, err := loadInventory(inventoryPath2)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := tview.NewApplication()

	// Create a list of hosts for the TUI
	listHostsGroupFirst := tview.NewList()
	for _, host := range inventory1.Hosts {
		listHostsGroupFirst.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil)
	}
	listHostsGroupFirst.SetBorder(true)
	listHostsGroupFirst.Box.SetBorderAttributes(tcell.RuneBoard).
		SetTitle("[black:darkcyan]First-Host-Group[white:-]").SetTitleAlign(tview.AlignLeft)
	listHostsGroupFirst.AddItem("", "", 'q', func() {
		app.Stop()
	})

	// Second list of hosts for the TUI
	listHostsGroupSecond := tview.NewList()
	for _, host := range inventory2.Hosts {
		listHostsGroupSecond.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil)
	}
	listHostsGroupSecond.SetBorder(true)
	listHostsGroupSecond.Box.SetBorderAttributes(tcell.RuneBoard).
		SetTitle("[black:red]Second-Host-Group[white:-]").SetTitleAlign(tview.AlignLeft)
	listHostsGroupSecond.AddItem("", "", 'q', func() {
		app.Stop()
	})

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
	listHostsGroupFirst.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventory1.Hosts[index]

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

	listHostsGroupSecond.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventory2.Hosts[index]

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
