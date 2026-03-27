// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package unimac

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/unpoller/unifi"
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
	CLIENT_OUI      = "OUI"
	CLIENT_OS       = "OS"
	CLIENT_OS_CLASS = "OS Class"
)

var (
	client_fields = []string{
		CLIENT_MAC, CLIENT_IP, CLIENT_HOSTNAME, CLIENT_NAME,
		CLIENT_OUI, CLIENT_OS, CLIENT_OS_CLASS,
		CLIENT_SITE, CLIENT_NETWORK, CLIENT_SWITCH, CLIENT_SWPORT,
		CLIENT_AP, CLIENT_RSSI, CLIENT_LASTSEEN, CLIENT_NOTE,
	}
	// 87, 15 - Vyos routern
	os_name_map = map[int]string{
		0:  "Unidentified",
		1:  "Others",
		3:  "Linux",
		12: "Mac OS X",
		18: "Windows",
		24: "iPhone",
		25: "Apple IOS",
		56: "Android",
		68: "Raspbian",
		87: "Vyos ?",
	}

	os_class_map = map[int]map[int]string{
		0: {
			0: "Unidentified",
		},
		12: {
			7: "Apple MacBook",
		},
		18: {
			15: "HP Laptop",
		},
		24: {
			16: "Apple iPhone 14 Pro Max",
		},
		25: {
			16: "Apple iPhone",
		},
		87: {
			15: "Vyos Router?",
		},
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
