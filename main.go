package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unpoller/unifi"
)

var (
	urlFlag      = flag.String("host", "https://unifi.lcl.kapi.se", "host address for controller")
	usernameFlag = flag.String("u", "", "Username (default is a secret)")
	passwordFlag = flag.String("p", "", "Password (default is a secret)")
)

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

func main() {
	flag.Parse()

	args := flag.Args()
	fmt.Println("asdfsadf:", args)

	switch args[0] {
	case "devices":
		uni, sites := mustConnect()
		devicesCmd.Parse(args[1:])
		generateDevices(uni, sites)

	case "clients":
		uni, sites := mustConnect()
		clientsCmd.Parse(args[1:])
		generateClients(uni, sites)
	default:
		log.Fatalf("[ERROR] unkown command '%s'", args[0])
	}
}

func mustConnect() (*unifi.Unifi, []*unifi.Site) {
	if *usernameFlag == "" {
		*usernameFlag = os.Getenv("UNIMAC_USER")
	}
	if *passwordFlag == "" {
		*passwordFlag = os.Getenv("UNIMAC_PASSWORD")
	}
	uni, err := connect(*usernameFlag, *passwordFlag, *urlFlag)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println("Connected to ", *urlFlag)

	sites, err := uni.GetSites()
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println(len(sites), "Unifi Sites Found")

	return uni, sites
}
