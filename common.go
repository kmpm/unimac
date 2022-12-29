package main

import (
	"encoding/json"
	"log"
	"os"
)

func writeJSON(data any, filename string) error {
	file, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, file, 0644)
}

func check(err error) {
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
}
