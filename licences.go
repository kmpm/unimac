package main

import (
	"flag"
	"fmt"
)

var (
	licencesCmd = flag.NewFlagSet("licenses", flag.ExitOnError)
)

func printMIT(title string, copyrights ...string) {

	text := `MIT LICENSE.
%s
Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.
`
	var t string
	for _, c := range copyrights {
		t += fmt.Sprintf("Copyright (c) %s\n", c)
	}
	fmt.Printf("\n\n-- %s --\n", title)
	fmt.Printf(text, t)
}

func printBSD3(title string, copyrights ...string) {
	text := `BSD 3-Clause License

%s
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of the copyright holder nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.
`
	var t string
	for _, c := range copyrights {
		t += fmt.Sprintf("Copyright (c) %s\n", c)
	}
	fmt.Printf("\n\n-- %s --\n", title)
	fmt.Printf(text, t)

}

func printDisclaimer(prefix, postfix string) {
	text := `THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.`
	fmt.Print(prefix)
	fmt.Println(text)
	fmt.Print(postfix)
}

func licensesRun(arguments []string) {
	check(licencesCmd.Parse(arguments))
	printMIT("unimac", "2021 Peter Magnusson <me@kmpm.se>")
	printDisclaimer("\n", "\n")

	fmt.Println("-- direct dependencies --")
	printMIT("github.com/unpoller/unifi", "2018-2020 David Newhall II", "2016 Garrett Bjerkhoel")
	printBSD3("github.com/qax-os/excelize", "2016-2022 The excelize Authors.")
}
