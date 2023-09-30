package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"time"

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
		list.AddItem(host.Name, "", 0, nil)
	}

	var connectingText *tview.TextView

	// Define the function to connect to the selected host
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventory.Hosts[index]

		// Clear the list and show connecting indicator
		list.Clear()
		connectingText = tview.NewTextView().
			SetText("Connecting...").
			SetTextAlign(tview.AlignCenter)
		app.SetRoot(connectingText, true)

		// Create a channel to handle the result of the connection attempt
		resultCh := make(chan error)

		// Connect to the host in a Goroutine
		go func() {
			resultCh <- connectToHost(host)
		}()

		// Wait for the connection result or timeout
		select {
		case err := <-resultCh:
			if err != nil {
				showErrorAndReturnToList(app, list, err)
			}
		case <-time.After(15 * time.Second):
			showErrorAndReturnToList(app, list, fmt.Errorf("Connection timed out"))
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

// connectToHost connects to the specified host using sshpass.
func connectToHost(host Host) error {
	cmd := exec.Command("sshpass", "-p", host.Password, "ssh", host.Username+"@"+host.Hostname)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// showErrorAndReturnToList displays an error message and returns to the host list.
func showErrorAndReturnToList(app *tview.Application, list *tview.List, err error) {
	errorText := tview.NewTextView().
		SetText("Error: " + err.Error()).
		SetTextAlign(tview.AlignCenter)

	returnButton := tview.NewTextView().
		SetText("Return to Hosts List (Press Enter)").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	app.SetRoot(
		tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(errorText, 0, 1, true).
			AddItem(returnButton, 1, 1, false),
		true,
	)

	// Wait for Enter key press to return to the host list
	app.SetInputCapture(func(event *tview.Event) *tview.Event {
		if event.Key == tview.KeyEnter {
			app.SetRoot(list, true)
		}
		return event
	})
}
