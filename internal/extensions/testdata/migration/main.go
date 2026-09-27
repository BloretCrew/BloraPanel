package main

import (
	"encoding/json"
	"os"
	"time"
)

func main() {
	var input struct {
		Operation string         `json:"operation"`
		From      int            `json:"fromVersion"`
		To        int            `json:"toVersion"`
		Data      map[string]any `json:"data"`
	}
	if json.NewDecoder(os.Stdin).Decode(&input) != nil || input.Operation != "blora.data.migrate" || input.To == 3 {
		os.Exit(2)
	}
	input.Data["migratedFrom"] = input.From
	if input.Data["slow"] == true {
		time.Sleep(2 * time.Second)
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"schemaVersion": input.To, "data": input.Data})
}
