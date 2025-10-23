// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package unimac

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/unpoller/unifi"
	"github.com/xuri/excelize/v2"
)

const (
	CLIENT_MAC      = "MAC"
	CLIENT_IP       = "IP"
	CLIENT_HOSTNAME = "Hostname"
	CLIENT_NAME     = "Name"
	CLIENT_NETWORK  = "Network"
	CLIENT_SWITCH   = "Switch"
	CLIENT_SWPORT   = "SwPort"
	CLIENT_AP       = "AP"
	CLIENT_RSSI     = "RSSI"
	CLIENT_NOTE     = "Note"
	CLIENT_SITE     = "Site"
	CLIENT_LASTSEEN = "Last Seen"
)

var (
	client_fields = []string{
		CLIENT_MAC, CLIENT_IP, CLIENT_HOSTNAME, CLIENT_NAME,
		CLIENT_SITE, CLIENT_NETWORK, CLIENT_SWITCH, CLIENT_SWPORT,
		CLIENT_AP, CLIENT_RSSI, CLIENT_LASTSEEN, CLIENT_NOTE,
	}
)

type ClientRender func(io.Writer, string, []*unifi.Client, map[string]*unifi.USW, map[string]*unifi.UAP) error

// GenerateClients takes a list of sites ange extracts
// information abouts clients and outputs depending on
// outputFlag
func GenerateClients(uni *unifi.Unifi, sites []*unifi.Site, renderer ClientRender, filename string) error {
	clients, err := uni.GetClients(sites)
	if err != nil {
		return fmt.Errorf("error getting clients: %v", err)
	}
	fmt.Println(len(clients), "Clients connected")
	devices, err := uni.GetDevices(sites)
	if err != nil {
		return fmt.Errorf("error matching devices: %v", err)
	}

	// if *sortFlag {
	// 	sort.SliceStable(clients, func(i, j int) bool {
	// 		return clients[i].Mac < clients[j].Mac
	// 	})
	// }

	// get a map of switches so that we can add information later
	switchmap := make(map[string]*unifi.USW)
	for _, sw := range devices.USWs {
		switchmap[sw.Mac] = sw
	}

	// get a map of access points
	apmap := make(map[string]*unifi.UAP)
	for _, ap := range devices.UAPs {
		apmap[ap.Mac] = ap
	}

	fmt.Printf("\t to %d switches and %d access points\n", len(devices.USWs), len(devices.UAPs))

	if filename != "" {
		f := mustCreateFile(filename)
		defer f.Close()
		return renderer(f, filename, clients, switchmap, apmap)
	} else {
		return renderer(os.Stdout, filename, clients, switchmap, apmap)
	}

}

func hydrateClient(client *unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	if client == nil {
		panic("client is nil")
	}
	if client.SwMac != "" {
		if sw, ok := switchmap[client.SwMac]; ok {
			client.SwName = sw.Name
		} else {
			// panic(fmt.Errorf("no switch found for %s", client.SwMac))
			client.SwName = client.SwMac
		}
	}
	if client.ApMac != "" {
		if ap, ok := apmap[client.ApMac]; ok {
			client.ApName = ap.Name
		} else {
			client.ApName = client.ApMac
		}
	}
}

func getClientValue(client *unifi.Client, name string) string {
	switch name {
	case CLIENT_MAC:
		return client.Mac
	case CLIENT_IP:
		return client.IP
	case CLIENT_HOSTNAME:
		return client.Hostname
	case CLIENT_NAME:
		return client.Name
	case CLIENT_NETWORK:
		return client.Network
	case CLIENT_SWITCH:
		return client.SwName
	case CLIENT_SWPORT:
		return client.SwPort.String()
	case CLIENT_AP:
		return client.ApName
	case CLIENT_RSSI:
		return fmt.Sprintf("%f", client.Rssi.Val)
	case CLIENT_NOTE:
		return client.Note
	case CLIENT_SITE:
		return client.SiteName
	case CLIENT_LASTSEEN:
		return time.Unix(int64(client.LastSeen.Val), 0).Format("2006-01-02 15:04:05")
	default:
		return "#UNSUPPORTED"
	}
}

func ClientJSON(out io.Writer, filename string, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) error {
	data, err := json.MarshalIndent(clients, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON %v", err)
	}
	_, err = out.Write(data)
	if err != nil {
		return fmt.Errorf("error writing JSON to output %v", err)
	}
	return nil
}

// clientTable outputs on screen in table format
func ClientTable(out io.Writer, filename string, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) error {
	const padding = 3
	w := tabwriter.NewWriter(out, 10, 0, padding, ' ', 0)
	header := ""
	format := ""
	for _, field := range client_fields {
		header += field + "\t"
		format += "%s\t"
	}
	format += "\n"
	fmt.Println("header", header)
	fmt.Println("format", format)
	fmt.Fprintln(w, header)

	values := make([]any, len(client_fields))
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		for i, field := range client_fields {
			values[i] = getClientValue(client, field)
		}
		fmt.Fprintf(w, format, values...)
	}
	w.Flush()
	return nil
}

// clientExcel outputs to .xslx -file
func ClientExcel(out io.Writer, filename string, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) error {
	f := excelize.NewFile()
	// index := f.NewSheet("Sheet1")

	cns := getColumns(client_fields...)

	sname := f.GetSheetName(0)
	for _, field := range client_fields {
		if err := f.SetCellValue(sname, cns[field](1), field); err != nil {
			return fmt.Errorf("error setting cell value: %w", err)
		}
	}

	row := 2
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		r := &runner{}
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_MAC](row), client.Mac)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_IP](row), client.IP)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_HOSTNAME](row), client.Hostname)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_NAME](row), client.Name)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_SITE](row), client.SiteName)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_NETWORK](row), client.Network)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_SWITCH](row), client.SwName)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_SWPORT](row), client.SwPort.Val)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_AP](row), client.ApName)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_RSSI](row), client.Rssi)
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_LASTSEEN](row), getClientValue(client, CLIENT_LASTSEEN))
		})
		r.run(func() error {
			return f.SetCellValue(sname, cns[CLIENT_NOTE](row), client.Note)
		})
		if r.err != nil {
			return fmt.Errorf("error setting cell value row %d: %w", row, r.err)
		}

		// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t\n",
		// 	client.Mac, client.IP, client.Hostname, client.Name,
		// 	client.SwName, &client.SwPort,
		// 	client.ApName, client.Rssi,
		// 	client.Note)
		row++
	}
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("error saving xlsx file: %w", err)
	}
	return nil
}

func ClientCsv(out io.Writer, filename string, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) error {

	w := csv.NewWriter(out)
	record := make([]string, len(client_fields))
	copy(record, client_fields)

	if err := w.Write(record); err != nil {
		return fmt.Errorf("error writing CSV header: %w", err)
	}
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		for i, field := range client_fields {
			record[i] = getClientValue(client, field)
		}

		if err := w.Write(record); err != nil {
			return fmt.Errorf("error writing CSV record: %w", err)
		}
	}
	w.Flush()
	return nil
}
