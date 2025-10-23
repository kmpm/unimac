// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
)

var Version = "v0.0.0-dev"
var BuildTime = "unknown"
var Commit = "unknown"

var (
	versionCmd = flag.NewFlagSet("version", flag.ExitOnError)
)

func printVersion() {
	fmt.Printf("unimac %s\n", Version)
}

func versionRun(arguments []string) {
	check(versionCmd.Parse(arguments))
	printVersion()

}
