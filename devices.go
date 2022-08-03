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
)

var (
	devicesCmd       = flag.NewFlagSet("devices", flag.ExitOnError)
	deviceOutputFlag = devicesCmd.String("output", "", "filename to output to. [*.csv]")
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
	if me.Name == "" {
		return me.Mac
	} else {
		return me.Name
	}
}

type Device struct {
	Mac    string
	Name   string
	Type   string
	Site   string
	IP     string
	Uplink DevicePort
	Note   string
}

func generateDevices(uni *unifi.Unifi, sites []*unifi.Site) {
	unifidevices, err := uni.GetDevices(sites)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	clients, err := uni.GetClients(sites)
	if err != nil {
		log.Fatalln("Error getting clients:", err)
	}

	switchmap := make(map[string]*unifi.USW)
	dlmap := make(map[string]DevicePort)
	for _, sw := range unifidevices.USWs {
		switchmap[sw.Mac] = sw
		// fmt.Println("sw: ", sw.Mac)
		for _, dl := range sw.DownlinkTable {
			// fmt.Printf("\t %s, %s\n", dl.Mac, dl.PortIdx.String())
			dlmap[dl.Mac] = DevicePort{Mac: sw.Mac, Name: sw.Name, Port: dl.PortIdx.String()}
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
	clientmap := make(map[string]*unifi.Client)
	for _, cl := range clients {
		clientmap[cl.Mac] = cl
	}

	var devices []*Device
	for _, sg := range unifidevices.USGs {
		d := &Device{
			Mac:    sg.Mac,
			Site:   sg.SiteName,
			Name:   sg.Name,
			IP:     sg.ConfigNetwork.IP,
			Type:   "USG",
			Uplink: DevicePort{Mac: sg.Uplink.Mac, Port: sg.Uplink.PortIdx.String()},
		}
		devices = append(devices, d)
	}
	for _, sw := range unifidevices.USWs {

		d := &Device{
			Mac:  sw.Mac,
			Site: sw.SiteName,
			Name: sw.Name,
			IP:   sw.ConfigNetwork.IP,
			Type: "USW",
		}
		if val, ok := dlmap[sw.Mac]; ok {
			d.Uplink = val
		} else {
			d.Uplink = DevicePort{Mac: sw.Uplink.Mac, Port: sw.Uplink.NumPort.String()}
			d.Note += "root"
		}

		devices = append(devices, d)
	}
	for _, ap := range unifidevices.UAPs {
		ul := dlmap[ap.Mac]
		d := &Device{
			Mac:    ap.Mac,
			Site:   ap.SiteName,
			Name:   ap.Name,
			IP:     ap.ConfigNetwork.IP,
			Type:   "UAP",
			Uplink: ul, //DevicePort{Mac: ap.Uplink.Mac, Port: strconv.Itoa(ap.Uplink.UplinkRemotePort)},
		}
		devices = append(devices, d)
	}

	for _, xg := range unifidevices.UXGs {
		ul := dlmap[xg.Mac]
		d := &Device{
			Mac:    xg.Mac,
			Site:   xg.SiteName,
			Name:   xg.Name,
			IP:     xg.ConfigNetwork.IP,
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
		case ".csv":
			devicesCsv(devices, *deviceOutputFlag)
		default:
			log.Fatalf("unsupported extension for %s", *deviceOutputFlag)
		}
	}
}

func deviceTable(devices []*Device) {

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 10, 0, padding, ' ', 0)
	fmt.Fprintln(w, "Mac\tType\tSite\tIP\tName\tNetwork\tUplink\tPort\tTBD\tNote\t")
	template := "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n"
	for _, d := range devices {
		fmt.Fprintf(w, template,
			d.Mac, d.Type, d.Site,
			d.IP,
			d.Name,
			"",
			d.Uplink.Displayname(),
			d.Uplink.Port,
			"",
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
	record[7] = "Port"
	record[8] = "TBD"
	record[9] = "Note"
	if err := w.Write(record); err != nil {
		log.Fatalln(err)
	}
	for _, d := range devices {
		w.Write([]string{
			d.Mac, d.Type, d.Site, d.IP, d.Name, "",
			d.Uplink.Displayname(), d.Uplink.Port,
			"",
			d.Note,
		})
	}
	w.Flush()
}
