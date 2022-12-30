// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT
package main

import (
	"testing"
)

func Test_getColumns(t *testing.T) {
	type args struct {
		name []string
	}
	type pair struct {
		name   string
		result string
	}

	samples1 := []string{"MAC", "IP", "Hostname", "Name", "Network", "Switch", "Port", "AP"}

	tests := []struct {
		name string
		args args
		want []pair
	}{
		{"0=A", args{samples1}, []pair{{"MAC", "A1"}}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getColumns(tt.args.name...)
			for _, v := range tt.want {
				got2 := got[v.name](1)
				if got2 != v.result {
					t.Errorf("got[%s](1) = %s, want %s", v.name, got2, v.result)
				}
			}
		})
	}
}
