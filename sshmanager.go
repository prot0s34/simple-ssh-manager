package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

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
	list := tview.NewList()
	for _, host := range inventory.Hosts {
		list.AddItem("host:"+host.Name, "user:"+host.Username, 0, nil)
	}

	// Define the function to connect to the selected host
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

	if err := app.SetRoot(list, true).Run(); err != nil {
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
