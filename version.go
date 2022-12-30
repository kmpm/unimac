// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
)

var appVersion = "v0.0.0-dev"

var (
	versionCmd = flag.NewFlagSet("version", flag.ExitOnError)
)

func printVersion() {
	fmt.Printf("unimac %s\n", appVersion)
}

func versionRun(arguments []string) {
	check(versionCmd.Parse(arguments))
	printVersion()

}
