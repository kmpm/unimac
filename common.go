package main

import (
	"encoding/json"
	"fmt"
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

type colname func(row int) string

var cols = [...]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

// getColumns takes a list of names and returns a map of functions
// that will return excel style cell keys An, Bn etc where n is the row.
func getColumns(name ...string) map[string]colname {
	index := make(map[string]colname)

	for k, v := range name {
		col := cols[k]
		index[v] = func(row int) string {
			return fmt.Sprintf("%s%d", col, row)
		}
	}
	return index
}

func mustCreateFile(filename string) *os.File {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalln(err)
	}
	return f
}
