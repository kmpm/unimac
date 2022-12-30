// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unpoller/unifi"
)

const (
	DefaultHost = "https://unifi"
)

var (
	hostFlag     = flag.String("h", DefaultHost, "host address for controller")
	usernameFlag = flag.String("u", "", "Username (default is a secret)")
	passwordFlag = flag.String("p", "", "Password (default is a secret)")
)

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
		uni, sites := mustConnect()
		check(devicesCmd.Parse(args[1:]))
		generateDevices(uni, sites)

	case "clients":
		uni, sites := mustConnect()
		check(clientsCmd.Parse(args[1:]))
		generateClients(uni, sites)
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
