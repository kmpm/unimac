package main

import (
	"encoding/json"
	"io/ioutil"
)

func writeJSON(data any, filename string) error {
	file, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, file, 0644)
}
