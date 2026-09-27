package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"unicode/utf8"
)

// This WASI command has only JSON stdin/stdout and analyzes the supplied note.
func main() {
	var in struct {
		Note      string         `json:"note"`
		Operation string         `json:"operation"`
		From      int            `json:"fromVersion"`
		To        int            `json:"toVersion"`
		Data      map[string]any `json:"data"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		panic(err)
	}
	hash := sha256.Sum256([]byte(in.Note))
	if in.Operation == "blora.data.migrate" {
		if in.To < 1 || in.To > 2 || in.Data == nil {
			os.Exit(2)
		}
		in.Data["noteFormat"] = in.To
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"schemaVersion": in.To, "data": in.Data}); err != nil {
			panic(err)
		}
		return
	}
	if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"characters": utf8.RuneCountInString(in.Note), "words": len(strings.Fields(in.Note)), "sha256": hex.EncodeToString(hash[:])}); err != nil {
		panic(err)
	}
}
