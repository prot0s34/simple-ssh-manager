package main

import (
	"encoding/json"
	"io/ioutil"
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

	file, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return nil, err
	}

	var inventoryGroups []InventoryGroup
	if err := json.Unmarshal(file, &inventoryGroups); err != nil {
		return nil, err
	}

	return inventoryGroups, nil
}
