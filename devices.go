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

	fmt.Printf("\t with %d USGs", len(unifidevices.USGs))
	var devices []*Device
	for _, sg := range unifidevices.USGs {
		d := &Device{
			Mac:           sg.Mac,
			Site:          sg.SiteName,
			Name:          sg.Name,
			IP:            sg.IP,
			Type:          "USG",
			Uplink:        &DevicePort{Mac: sg.Uplink.Mac, Port: sg.Uplink.PortIdx.String()},
			ConfigNetwork: sg.ConfigNetwork,
		}
		devices = append(devices, d)
	}

	fmt.Printf(",%d USWs", len(unifidevices.USWs))
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

		devices = append(devices, d)
	}

	fmt.Printf(", %d UAPs", len(unifidevices.UAPs))
	for _, ap := range unifidevices.UAPs {
		ul := dlmap[ap.Mac]
		d := &Device{
			Mac:    ap.Mac,
			Site:   ap.SiteName,
			Name:   ap.Name,
			IP:     ap.IP,
			Type:   "UAP",
			Uplink: ul, //DevicePort{Mac: ap.Uplink.Mac, Port: strconv.Itoa(ap.Uplink.UplinkRemotePort)},
			// ConfigNetwork: &unifi.ConfigNetwork{IP: ap.ConfigNetwork.IP, Type: ap.ConfigNetwork.Type},
		}
		if ap.ConfigNetwork != nil {
			d.ConfigNetwork = ap.ConfigNetwork
		}
		devices = append(devices, d)
	}

	fmt.Printf(" and %d UXGs\n", len(unifidevices.UXGs))
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

func deviceTable(devices []*Device) {

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 10, 0, padding, ' ', 0)
	fmt.Fprintln(w, "Mac\tType\tSite\tIP\tName\tNetwork\tUplink\tPort\tConfigIP\tNote\t")
	template := "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n"
	for _, d := range devices {
		fmt.Fprintf(w, template,
			d.Mac, d.Type, d.Site,
			d.IP,
			d.Name,
			"",
			d.Uplink.Displayname(),
			d.Uplink.Port,
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

func devicesExcel(devices []*Device, filename string) {
	f := excelize.NewFile()
	sname := f.GetSheetName(0)
	f.SetCellValue(sname, "A1", "MAC")
	f.SetCellValue(sname, "B1", "Type")
	f.SetCellValue(sname, "C1", "Site")
	f.SetCellValue(sname, "D1", "IP")
	f.SetCellValue(sname, "E1", "Name")
	f.SetCellValue(sname, "F1", "Network")
	f.SetCellValue(sname, "G1", "Uplink")
	f.SetCellValue(sname, "H1", "Port")
	f.SetCellValue(sname, "I1", "TBD")
	f.SetCellValue(sname, "J1", "Note")
	row := 2
	for _, d := range devices {
		f.SetCellValue(sname, fmt.Sprintf("A%d", row), d.Mac)
		f.SetCellValue(sname, fmt.Sprintf("B%d", row), d.Type)
		f.SetCellValue(sname, fmt.Sprintf("C%d", row), d.Site)
		f.SetCellValue(sname, fmt.Sprintf("D%d", row), d.IP)
		f.SetCellValue(sname, fmt.Sprintf("E%d", row), d.Name)
		f.SetCellValue(sname, fmt.Sprintf("F%d", row), "")
		f.SetCellValue(sname, fmt.Sprintf("G%d", row), d.Uplink.Displayname())
		f.SetCellValue(sname, fmt.Sprintf("H%d", row), d.Uplink.Port)
		f.SetCellValue(sname, fmt.Sprintf("I%d", row), "")
		f.SetCellValue(sname, fmt.Sprintf("J%d", row), d.Note)
		row++
	}
	if err := f.SaveAs(*deviceOutputFlag); err != nil {
		log.Println(err)
	}
}
