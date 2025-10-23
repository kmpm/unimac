// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kmpm/unimac/public/unimac"
	"github.com/unpoller/unifi"
)

const (
	DefaultHost = "https://unifi"
)

var (
	hostFlag     = flag.String("h", DefaultHost, "host address for controller")
	usernameFlag = flag.String("u", "", "Username (default is a secret)")
	passwordFlag = flag.String("p", "", "Password (default is a secret)")

	clientsCmd = flag.NewFlagSet("clients", flag.ExitOnError)
	sortFlag   = clientsCmd.Bool("sort", false, "sort my MAC")
	outputFlag = clientsCmd.String("output", "", "filename to output to. [*.xlsx, *.json, *.csv]")

	devicesCmd       = flag.NewFlagSet("devices", flag.ExitOnError)
	deviceOutputFlag = devicesCmd.String("output", "", "filename to output to. [*.xlsx, *.json, *.csv]")
)

func check(err error) {
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
}

func connect(user, pass, url string) (*unifi.Unifi, error) {
	c := &unifi.Config{
		User: user,
		Pass: pass,
		URL:  url,
		// Log with log.Printf or make your own interface that accepts (msg, fmt)
		ErrorLog: log.Printf,
		// DebugLog: log.Printf,
	}
	uni, err := unifi.NewUnifi(c)

	return uni, err
}

func main() {
	flag.Parse()

	args := flag.Args()
	if *usernameFlag == "" {
		*usernameFlag = os.Getenv("UNIMAC_USER")
	}
	if *passwordFlag == "" {
		*passwordFlag = os.Getenv("UNIMAC_PASSWORD")
	}
	hostEnv := os.Getenv("UNIMAC_HOST")
	if hostEnv != "" && (*hostFlag == "" || *hostFlag == DefaultHost) {
		*hostFlag = hostEnv
	}

	switch args[0] {
	case "devices":
		doDevices(args)
	case "clients":
		doClients(args)
	case "version":
		versionRun(args[1:])
	case "licenses":
		licensesRun(args[1:])
	default:
		log.Fatalf("[ERROR] unkown command '%s'", args[0])
	}
}

func mustConnect() (*unifi.Unifi, []*unifi.Site) {
	uni, err := connect(*usernameFlag, *passwordFlag, *hostFlag)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println("Connected to ", *hostFlag)

	sites, err := uni.GetSites()
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println(len(sites), "Unifi Sites Found")

	return uni, sites
}

func doClients(args []string) {
	uni, sites := mustConnect()
	check(clientsCmd.Parse(args[1:]))

	var renderer unimac.ClientRender
	ext := ".table"
	if *outputFlag != "" {
		ext = filepath.Ext(*outputFlag)
	}
	switch ext {
	case ".xlsx":
		renderer = unimac.ClientExcel
	case ".json":
		renderer = unimac.ClientJSON
	case ".csv":
		renderer = unimac.ClientCsv
	case ".table":
		renderer = unimac.ClientTable
	default:
		log.Fatalf("unsupported extension for %s", *outputFlag)
	}
	err := unimac.GenerateClients(uni, sites, renderer, *outputFlag)
	check(err)

}

func doDevices(args []string) {
	uni, sites := mustConnect()
	check(devicesCmd.Parse(args[1:]))
	devices, err := unimac.GenerateDevices(uni, sites)
	check(err)

	if *deviceOutputFlag == "" {
		unimac.DeviceTable(devices)
	} else {
		ext := filepath.Ext(*deviceOutputFlag)
		switch ext {
		case ".json":
			err := unimac.WriteJSON(devices, *deviceOutputFlag)
			if err != nil {
				log.Fatalf("error writing '%s': %v", *deviceOutputFlag, err)
			}
		case ".csv":
			unimac.DevicesCsv(devices, *deviceOutputFlag)
		case ".xlsx":
			unimac.DevicesExcel(devices, *deviceOutputFlag)
		default:
			log.Fatalf("unsupported extension for %s", *deviceOutputFlag)
		}
	}
}
