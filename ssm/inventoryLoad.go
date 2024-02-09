package main

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
)

func loadInventoryGroups() ([]InventoryGroup, error) {
	inventoryPath := os.Getenv("SSHMANAGER_INVENTORY")

	if inventoryPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		inventoryPath = usr.HomeDir + "/inventory.json"
	}

	file, err := os.ReadFile(inventoryPath)
	if err != nil {
		return nil, err
	}

	var inventoryGroups []InventoryGroup
	if err := json.Unmarshal(file, &inventoryGroups); err != nil {
		return nil, err
	}

	for _, ig := range inventoryGroups {
		if err := ig.Validate(); err != nil {
			log.Fatalf("Validation error in inventory group '%s': %v\nPlease check your inventory file: %s", ig.Name, err, inventoryPath)
		}
	}
	return inventoryGroups, nil
}
