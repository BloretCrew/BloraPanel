package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	var in struct {
		Mode  string `json:"mode"`
		Value int    `json:"value"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		panic(err)
	}
	switch in.Mode {
	case "loop":
		for {
		}
	case "overflow":
		fmt.Print(strings.Repeat("x", 128<<10))
	case "isolation":
		_, err := os.ReadFile("/etc/passwd")
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"fileDenied": err != nil, "environment": len(os.Environ())})
	default:
		_ = json.NewEncoder(os.Stdout).Encode(map[string]int{"squared": in.Value * in.Value})
	}
}
