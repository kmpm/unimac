package main

import (
	"encoding/json"
	"os"
)

func writeJSON(data any, filename string) error {
	file, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, file, 0644)
}
