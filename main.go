package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/unpoller/unifi"
	"github.com/xuri/excelize/v2"
)

var outputFlag string
var urlFlag string
var usernameFlag string
var passwordFlag string
var sortFlag bool

func init() {
	flag.StringVar(&outputFlag, "output", "", "filename to output to. [*.xlsx, *.json, *.csv]")
	flag.StringVar(&outputFlag, "o", "", "shorthand for output")
	flag.StringVar(&urlFlag, "host", "https://unifi.lcl.kapi.se", "host address for controller")
	flag.StringVar(&usernameFlag, "u", "", "Username (default is a secret)")
	flag.StringVar(&passwordFlag, "p", "", "Password (default is a secret)")
	flag.BoolVar(&sortFlag, "sort", false, "sort my MAC")
}

func connect(user, pass, url string) (*unifi.Unifi, error) {
	c := &unifi.Config{
		User: user,
		Pass: pass,
		URL:  url,
		// Log with log.Printf or make your own interface that accepts (msg, fmt)
		ErrorLog: log.Printf,
		// DebugLog: log.Printf,
	}
	uni, err := unifi.NewUnifi(c)

	return uni, err
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

func clientTable(clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 10, 0, padding, ' ', 0)
	fmt.Fprintln(w, "Mac\tIP\tHostname\tName\tNetwork\tSwitch\tPort\tAP\tRSSI\tNote\t")
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t\n",
			client.Mac, client.IP, client.Hostname, client.Name,
			client.Network,
			client.SwName, &client.SwPort,
			client.ApName, client.Rssi,
			client.Note)
	}
	w.Flush()
}

func clientExcel(clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	f := excelize.NewFile()
	// index := f.NewSheet("Sheet1")
	sname := f.GetSheetName(0)
	f.SetCellValue(sname, "A1", "MAC")
	f.SetCellValue(sname, "B1", "IP")
	f.SetCellValue(sname, "C1", "Hostname")
	f.SetCellValue(sname, "D1", "Name")
	f.SetCellValue(sname, "E1", "Network")
	f.SetCellValue(sname, "F1", "Switch")
	f.SetCellValue(sname, "G1", "Port")
	f.SetCellValue(sname, "H1", "AP")
	f.SetCellValue(sname, "I1", "RSSI")
	f.SetCellValue(sname, "J1", "Note")
	row := 2
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		f.SetCellValue(sname, fmt.Sprintf("A%d", row), client.Mac)
		f.SetCellValue(sname, fmt.Sprintf("B%d", row), client.IP)
		f.SetCellValue(sname, fmt.Sprintf("C%d", row), client.Hostname)
		f.SetCellValue(sname, fmt.Sprintf("D%d", row), client.Name)
		f.SetCellValue(sname, fmt.Sprintf("G%d", row), client.Network)
		f.SetCellValue(sname, fmt.Sprintf("F%d", row), client.SwName)
		f.SetCellValue(sname, fmt.Sprintf("G%d", row), client.SwPort.Val)
		f.SetCellValue(sname, fmt.Sprintf("H%d", row), client.ApName)
		f.SetCellValue(sname, fmt.Sprintf("I%d", row), client.Rssi)
		f.SetCellValue(sname, fmt.Sprintf("J%d", row), client.Note)

		// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t\n",
		// 	client.Mac, client.IP, client.Hostname, client.Name,
		// 	client.SwName, &client.SwPort,
		// 	client.ApName, client.Rssi,
		// 	client.Note)
		row++
	}
	if err := f.SaveAs(outputFlag); err != nil {
		log.Println(err)
	}
}

func clientCsv(clients []*unifi.Client, switchmap map[string]*unifi.USW, apmap map[string]*unifi.UAP) {
	f, err := os.Create(outputFlag)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	record := make([]string, 10)
	record[0] = "MAC"
	record[1] = "IP"
	record[2] = "Hostname"
	record[3] = "Name"
	record[4] = "Network"
	record[5] = "Switch"
	record[6] = "Port"
	record[7] = "AP"
	record[8] = "RSSI"
	record[9] = "Note"
	if err := w.Write(record); err != nil {
		log.Fatalln(err)
	}
	for _, client := range clients {
		hydrateClient(client, switchmap, apmap)
		w.Write([]string{
			client.Mac, client.IP, client.Hostname, client.Name,
			client.Network,
			client.SwName, client.SwPort.Txt,
			client.ApName, fmt.Sprintf("%d", client.Rssi),
			client.Note,
		})
	}
	w.Flush()
}

func main() {
	flag.Parse()

	if usernameFlag == "" {
		usernameFlag = "peterm"
	}
	if passwordFlag == "" {
		passwordFlag = "bsI14OOp%O*B"
	}
	uni, err := connect(usernameFlag, passwordFlag, urlFlag)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	sites, err := uni.GetSites()
	if err != nil {
		log.Fatalln("Error:", err)
	}
	clients, err := uni.GetClients(sites)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	devices, err := uni.GetDevices(sites)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	switchmap := make(map[string]*unifi.USW)
	for _, sw := range devices.USWs {
		switchmap[sw.Mac] = sw
	}

	apmap := make(map[string]*unifi.UAP)
	for _, ap := range devices.UAPs {
		apmap[ap.Mac] = ap
	}

	// log.Println(len(sites), "Unifi Sites Found: ", sites)
	// log.Println(len(clients), "Clients connected:")
	if sortFlag {
		sort.SliceStable(clients, func(i, j int) bool {
			return clients[i].Mac < clients[j].Mac
		})
	}

	if outputFlag == "" {
		clientTable(clients, switchmap, apmap)
	} else {
		ext := filepath.Ext(outputFlag)
		switch ext {
		case ".xlsx":
			clientExcel(clients, switchmap, apmap)
		case ".json":
			file, _ := json.MarshalIndent(clients, "", "    ")
			_ = ioutil.WriteFile(outputFlag, file, 0644)
		case ".csv":
			clientCsv(clients, switchmap, apmap)
		default:
			log.Fatalf("unsupported extension for %s", outputFlag)
		}
	}
}
