package unimac

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/unpoller/unifi/v5"
	"github.com/xuri/excelize/v2"
)

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
		for _, field := range client_fields {
			switch field {
			case CLIENT_MAC:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_MAC](row), client.Mac)
				})
			case CLIENT_IP:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_IP](row), client.IP)
				})
			case CLIENT_HOSTNAME:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_HOSTNAME](row), client.Hostname)
				})
			case CLIENT_NAME:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_NAME](row), client.Name)
				})
			case CLIENT_SITE:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_SITE](row), client.SiteName)
				})
			case CLIENT_NETWORK:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_NETWORK](row), client.Network)
				})
			case CLIENT_SWITCH:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_SWITCH](row), client.SwName)
				})
			case CLIENT_SWPORT:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_SWPORT](row), client.SwPort.Val)
				})
			case CLIENT_AP:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_AP](row), client.ApName)
				})
			case CLIENT_RSSI:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_RSSI](row), client.Rssi.Val)
				})
			case CLIENT_LASTSEEN:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_LASTSEEN](row), getClientValue(client, CLIENT_LASTSEEN))
				})
			case CLIENT_NOTE:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_NOTE](row), client.Note)
				})
			case CLIENT_OUI:
				r.run(func() error {
					return f.SetCellValue(sname, cns[CLIENT_OUI](row), client.Oui)
				})
			case CLIENT_OS:
				r.run(func() error {

					if v, ok := os_name_map[client.OsName.Int()]; ok {
						return f.SetCellValue(sname, cns[CLIENT_OS](row), v)
					}
					return f.SetCellValue(sname, cns[CLIENT_OS](row), client.OsName.Val)
				})
			case CLIENT_OS_CLASS:
				r.run(func() error {
					if ov, ok := os_class_map[client.OsName.Int()]; ok {
						if oc, ok := ov[client.OsClass.Int()]; ok {
							return f.SetCellValue(sname, cns[CLIENT_OS_CLASS](row), oc)
						}
					}
					return f.SetCellValue(sname, cns[CLIENT_OS_CLASS](row), client.OsClass)
				})

			}
			if r.err != nil {
				return fmt.Errorf("error setting cell value row %d: %w", row, r.err)
			}
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
