// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/unpoller/unifi"
	"github.com/xuri/excelize/v2"
)

var (
	devicesCmd       = flag.NewFlagSet("devices", flag.ExitOnError)
	deviceOutputFlag = devicesCmd.String("output", "", "filename to output to. [*.xlsx, *.json, *.csv]")
)

type DevicePort struct {
	Mac  string
	Name string
	Port string
}

func (me *DevicePort) String() string {
	return me.Displayname() + " " + me.Port
}

func (me *DevicePort) Displayname() string {
	// fmt.Printf("me %v\n", me)
	// fmt.Printf("name %v\n", me.Name)
	// fmt.Printf("mac %v\n", me.Mac)
	if me == nil {
		fmt.Println("nil")
	}
	if me.Name == "" {
		return me.Mac
	} else {
		return me.Name
	}
}

type Device struct {
	Mac           string
	Name          string
	Type          string
	Site          string
	IP            string
	Uplink        *DevicePort
	Note          string
	ConfigNetwork *unifi.ConfigNetwork
}

func generateDevices(uni *unifi.Unifi, sites []*unifi.Site) {
	unifidevices, err := uni.GetDevices(sites)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	// clients, err := uni.GetClients(sites)
	// if err != nil {
	// 	log.Fatalln("Error getting clients:", err)
	// }

	switchmap := make(map[string]*unifi.USW)
	dlmap := make(map[string]*DevicePort)
	for _, sw := range unifidevices.USWs {
		switchmap[sw.Mac] = sw
		// fmt.Println("sw: ", sw.Mac)
		for _, dl := range sw.DownlinkTable {
			// fmt.Printf("\t %s, %s\n", dl.Mac, dl.PortIdx.String())
			dlmap[dl.Mac] = &DevicePort{Mac: sw.Mac, Name: sw.Name, Port: dl.PortIdx.String()}
		}
	}

	apmap := make(map[string]*unifi.UAP)
	for _, ap := range unifidevices.UAPs {
		apmap[ap.Mac] = ap
	}

	// gwmap := make(map[string]*unifi.USG)
	// for _, sg := range devices.USGs {
	// 	gwmap[sg.Mac] = sg
	// }
	// clientmap := make(map[string]*unifi.Client)
	// for _, cl := range clients {
	// 	clientmap[cl.Mac] = cl
	// }
	var devices []*Device
	before := len(devices)
	fmt.Printf("\t with %d USGs", len(unifidevices.USGs))
	withUSGs(unifidevices, &devices)
	fmt.Printf(" added %d \n", len(devices)-before)

	before = len(devices)
	fmt.Printf("\t with %d USWs", len(unifidevices.USWs))
	withUSWs(unifidevices, &devices, dlmap)
	fmt.Printf(" added %d\n", len(devices)-before)

	fmt.Printf("\t with %d UAPs", len(unifidevices.UAPs))
	before = len(devices)
	withUAPs(unifidevices, &devices, dlmap)
	fmt.Printf(" added %d\n", len(devices)-before)

	fmt.Printf("\t with %d UXGs\n", len(unifidevices.UXGs))
	for _, xg := range unifidevices.UXGs {
		ul := dlmap[xg.Mac]
		d := &Device{
			Mac:    xg.Mac,
			Site:   xg.SiteName,
			Name:   xg.Name,
			IP:     xg.IP,
			Type:   "UXG",
			Uplink: ul, //DevicePort{Mac: ap.Uplink.Mac, Port: strconv.Itoa(ap.Uplink.UplinkRemotePort)},
		}
		devices = append(devices, d)
	}

	if *deviceOutputFlag == "" {
		deviceTable(devices)
	} else {
		ext := filepath.Ext(*deviceOutputFlag)
		switch ext {
		case ".json":
			err := writeJSON(devices, *deviceOutputFlag)
			if err != nil {
				log.Fatalf("error writing '%s': %v", *deviceOutputFlag, err)
			}
		case ".csv":
			devicesCsv(devices, *deviceOutputFlag)
		case ".xlsx":
			devicesExcel(devices, *deviceOutputFlag)
		default:
			log.Fatalf("unsupported extension for %s", *deviceOutputFlag)
		}
	}
}

func withUSGs(unifidevices *unifi.Devices, devices *[]*Device) {

	for _, sg := range unifidevices.USGs {
		if sg == nil {
			continue
		}
		ul := &DevicePort{Mac: sg.Uplink.Mac, Port: sg.Uplink.PortIdx.String()}

		d := &Device{
			Mac:           sg.Mac,
			Site:          sg.SiteName,
			Name:          sg.Name,
			IP:            sg.IP,
			Type:          "USG",
			Uplink:        ul,
			ConfigNetwork: sg.ConfigNetwork,
		}

		*devices = append(*devices, d)
	}
}

func withUSWs(unifidevices *unifi.Devices, devices *[]*Device, dlmap map[string]*DevicePort) {

	for _, sw := range unifidevices.USWs {
		d := &Device{
			Mac:           sw.Mac,
			Site:          sw.SiteName,
			Name:          sw.Name,
			IP:            sw.IP,
			Type:          "USW",
			ConfigNetwork: sw.ConfigNetwork,
		}

		if val, ok := dlmap[sw.Mac]; ok {
			d.Uplink = val
		} else {
			d.Uplink = &DevicePort{Mac: sw.Uplink.Mac, Port: sw.Uplink.NumPort.String()}
			d.Note += "root"
		}

		*devices = append(*devices, d)
	}
}

func withUAPs(unifidevices *unifi.Devices, devices *[]*Device, dlmap map[string]*DevicePort) {
	for _, ap := range unifidevices.UAPs {
		ul := dlmap[ap.Mac]
		d := &Device{
			Mac:  ap.Mac,
			Site: ap.SiteName,
			Name: ap.Name,
			IP:   ap.IP,
			Type: "UAP",
			// Uplink: *ul, //DevicePort{Mac: ap.Uplink.Mac, Port: strconv.Itoa(ap.Uplink.UplinkRemotePort)},
			// ConfigNetwork: &unifi.ConfigNetwork{IP: ap.ConfigNetwork.IP, Type: ap.ConfigNetwork.Type},
		}
		if ul != nil {
			d.Uplink = ul
		}
		if ap.ConfigNetwork != nil {
			d.ConfigNetwork = ap.ConfigNetwork
		}
		*devices = append(*devices, d)
	}
}

func deviceTable(devices []*Device) {

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 10, 0, padding, ' ', 0)
	fmt.Fprintln(w, "Mac\tType\tSite\tIP\tName\tNetwork\tUplink\tUpPort\tConfigIP\tNote\t")
	template := "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n"
	for _, d := range devices {
		ul_name := "nil"
		ul_port := "nil"
		if d.Uplink != nil {
			ul_name = d.Uplink.Displayname()
			ul_port = d.Uplink.Port
		}
		fmt.Fprintf(w, template,
			d.Mac, d.Type, d.Site,
			d.IP,
			d.Name,
			"",
			ul_name,
			ul_port,
			d.ConfigNetwork.IP,
			d.Note,
		)
	}
	w.Flush()
}

func devicesCsv(devices []*Device, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	record := make([]string, 10)
	record[0] = "MAC"
	record[1] = "Type"
	record[2] = "Site"
	record[3] = "IP"
	record[4] = "Name"
	record[5] = "Network"
	record[6] = "Uplink"
	record[7] = "UpPort"
	record[8] = "ConfigIP"
	record[9] = "Note"
	if err := w.Write(record); err != nil {
		log.Fatalln(err)
	}
	for _, d := range devices {
		ul_name := "nil"
		ul_port := "nil"
		if d.Uplink != nil {
			ul_name = d.Uplink.Displayname()
			ul_port = d.Uplink.Port
		}
		cfn_ip := "nil"
		if d.ConfigNetwork != nil {
			cfn_ip = d.ConfigNetwork.IP
		}
		err := w.Write([]string{
			d.Mac, d.Type, d.Site, d.IP, d.Name, "",
			ul_name, ul_port,
			cfn_ip,
			d.Note,
		})
		check(err)
	}
	w.Flush()
}

func devicesExcel(devices []*Device, filename string) {
	f := excelize.NewFile()
	sname := f.GetSheetName(0)
	check(f.SetCellValue(sname, "A1", "MAC"))
	check(f.SetCellValue(sname, "B1", "Type"))
	check(f.SetCellValue(sname, "C1", "Site"))
	check(f.SetCellValue(sname, "D1", "IP"))
	check(f.SetCellValue(sname, "E1", "Name"))
	check(f.SetCellValue(sname, "F1", "Network"))
	check(f.SetCellValue(sname, "G1", "Uplink"))
	check(f.SetCellValue(sname, "H1", "UpPort"))
	check(f.SetCellValue(sname, "I1", "ConfigIP"))
	check(f.SetCellValue(sname, "J1", "Note"))
	row := 2
	for _, d := range devices {
		ul_name := "nil"
		ul_port := "nil"
		if d.Uplink != nil {
			ul_name = d.Uplink.Displayname()
			ul_port = d.Uplink.Port
		}
		check(f.SetCellValue(sname, fmt.Sprintf("A%d", row), d.Mac))
		check(f.SetCellValue(sname, fmt.Sprintf("B%d", row), d.Type))
		check(f.SetCellValue(sname, fmt.Sprintf("C%d", row), d.Site))
		check(f.SetCellValue(sname, fmt.Sprintf("D%d", row), d.IP))
		check(f.SetCellValue(sname, fmt.Sprintf("E%d", row), d.Name))
		check(f.SetCellValue(sname, fmt.Sprintf("F%d", row), ""))
		check(f.SetCellValue(sname, fmt.Sprintf("G%d", row), ul_name))
		check(f.SetCellValue(sname, fmt.Sprintf("H%d", row), ul_port))
		if d.ConfigNetwork != nil {
			check(f.SetCellValue(sname, fmt.Sprintf("I%d", row), d.ConfigNetwork.IP))
		} else {
			check(f.SetCellValue(sname, fmt.Sprintf("I%d", row), "nil"))
		}
		check(f.SetCellValue(sname, fmt.Sprintf("J%d", row), d.Note))
		row++
	}
	if err := f.SaveAs(*deviceOutputFlag); err != nil {
		log.Println(err)
	}
}
