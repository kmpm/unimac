package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
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
	clientsCmd    = flag.NewFlagSet("clients", flag.ExitOnError)
	sortFlag      = clientsCmd.Bool("sort", false, "sort my MAC")
	outputFlag    = clientsCmd.String("output", "", "filename to output to. [*.xlsx, *.json, *.csv]")
	client_fields = []string{
		CLIENT_MAC, CLIENT_IP, CLIENT_HOSTNAME, CLIENT_NAME,
		CLIENT_SITE, CLIENT_NETWORK, CLIENT_SWITCH, CLIENT_SWPORT,
		CLIENT_AP, CLIENT_RSSI, CLIENT_LASTSEEN, CLIENT_NOTE,
	}
)

type clientRender func(io.Writer, []*unifi.Client, map[string]*unifi.USW, map[string]*unifi.UAP)

// generateClients takes a list of sites ange extracts
// information abouts clients and outputs depending on
// outputFlag
func generateClients(uni *unifi.Unifi, sites []*unifi.Site) {
	clients, err := uni.GetClients(sites)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println(len(clients), "Clients connected")

	devices, err := uni.GetDevices(sites)
	if err != nil {
		log.Fatalln("Error:", err)
	}

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

	if *sortFlag {
		sort.SliceStable(clients, func(i, j int) bool {
			return clients[i].Mac < clients[j].Mac
		})
	}

	var renderer clientRender
	ext := ".table"
	if *outputFlag != "" {
		ext = filepath.Ext(*outputFlag)
	}
	switch ext {
	case ".xlsx":
		renderer = clientExcel
	case ".json":
		renderer = clientJSON
	case ".csv":
		renderer = clientCsv
	case ".table":
		renderer = clientTable
	default:
		log.Fatalf("unsupported extension for %s", *outputFlag)
	}
	if *outputFlag != "" {
		f := mustCreateFile(*outputFlag)
		defer f.Close()
		renderer(f, clients, switchmap, apmap)
	} else {
		renderer(os.Stdout, clients, switchmap, apmap)
	}

}

func hydrateClient(client *unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	if client.SwMac != "" {
		sw := switchmap[client.SwMac]
		client.SwName = sw.Name
	}
	if client.ApMac != "" {
		ap := apmap[client.ApMac]
		client.ApName = ap.Name
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
		return fmt.Sprintf("%d", client.Rssi)
	case CLIENT_NOTE:
		return client.Note
	case CLIENT_SITE:
		return client.SiteName
	case CLIENT_LASTSEEN:
		return time.Unix(client.LastSeen, 0).Format("2006-01-02 15:04:05")
	default:
		return "#UNSUPPORTED"
	}
}

func clientJSON(out io.Writer, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	data, err := json.MarshalIndent(clients, "", "    ")
	if err != nil {
		log.Fatalf("error marshalling to JSON %v", err)
	}
	_, err = out.Write(data)
	check(err)
}

// clientTable outputs on screen in table format
func clientTable(out io.Writer, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
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
}

// clientExcel outputs to .xslx -file
func clientExcel(out io.Writer, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	f := excelize.NewFile()
	// index := f.NewSheet("Sheet1")

	cns := getColumns(client_fields...)

	sname := f.GetSheetName(0)
	for _, field := range client_fields {
		check(f.SetCellValue(sname, cns[field](1), field))
	}

	row := 2
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		check(f.SetCellValue(sname, cns[CLIENT_MAC](row), client.Mac))
		check(f.SetCellValue(sname, cns[CLIENT_IP](row), client.IP))
		check(f.SetCellValue(sname, cns[CLIENT_HOSTNAME](row), client.Hostname))
		check(f.SetCellValue(sname, cns[CLIENT_NAME](row), client.Name))
		check(f.SetCellValue(sname, cns[CLIENT_SITE](row), client.SiteName))
		check(f.SetCellValue(sname, cns[CLIENT_NETWORK](row), client.Network))
		check(f.SetCellValue(sname, cns[CLIENT_SWITCH](row), client.SwName))
		check(f.SetCellValue(sname, cns[CLIENT_SWPORT](row), client.SwPort.Val))
		check(f.SetCellValue(sname, cns[CLIENT_AP](row), client.ApName))
		check(f.SetCellValue(sname, cns[CLIENT_RSSI](row), client.Rssi))
		check(f.SetCellValue(sname, cns[CLIENT_LASTSEEN](row), time.Unix(client.LastSeen, 0)))
		check(f.SetCellValue(sname, cns[CLIENT_NOTE](row), client.Note))

		// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t\n",
		// 	client.Mac, client.IP, client.Hostname, client.Name,
		// 	client.SwName, &client.SwPort,
		// 	client.ApName, client.Rssi,
		// 	client.Note)
		row++
	}
	if err := f.SaveAs(*outputFlag); err != nil {
		log.Println(err)
	}
}

func clientCsv(out io.Writer, clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {

	w := csv.NewWriter(out)
	record := make([]string, len(client_fields))
	copy(record, client_fields)

	if err := w.Write(record); err != nil {
		log.Fatalln(err)
	}
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		for i, field := range client_fields {
			record[i] = getClientValue(client, field)
		}

		check(w.Write(record))
	}
	w.Flush()
}
