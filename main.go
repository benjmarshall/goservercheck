package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/0xAX/notificator"
)

var notify *notificator.Notificator

func main() {

	// Set options for the notifyier
	notify = notificator.New(notificator.Options{AppName: "GoServerCheck"})

	// Default name for the server list file
	serverListFile := ".goservercheck"

	// Create a server list struct which will use the reader dfunction and default server file name
	serverList := newServerList(reader, serverListFile)

	// Read the server list
	serverList.read()
	fmt.Println(serverList.sl)

	// Run the server checks
	ok, errorsList := runChecks(*serverList)

	// If any of the checks failed to connect to the server issue an error. (note this does not include http errors)
	if !ok {
		notifyAndExit(errorsList)
	}

	// If we received any http errors then issue an error notice, else give the all clear
	if errorsList != "" {
		notify.Push("GoServerCheck: Problems Detected", errorsList, "", notificator.UR_CRITICAL)
	} else {
		notify.Push("GoServerCheck: Servers OK", "All web pages reached.", "", notificator.UR_NORMAL)
	}

}

// reader defines the functionality for reading a server list file
func reader(serverListFile string) []string {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(usr.HomeDir + "/" + serverListFile)
	if err != nil {
		notifyAndExit("Can't find ~/" + serverListFile)
	}
	defer f.Close()

	var serverList []string
	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		serverList = append(serverList, fileScanner.Text())
	}
	return serverList

}

// runChecks runs the server checks for each server in the list
func runChecks(serverList serverListType) (bool, string) {

	var errorsList string

	for _, url := range serverList.sl {
		status, statusString := checkServer(url)
		if status == -1 {
			errorsList = errorsList + fmt.Sprintf("Error occured connecting to %s:\n%v", url, statusString)
			return false, errorsList
		} else if status == 1 {
			errorsList = errorsList + fmt.Sprintf("%s, appears to be down. Resonse was:%s\n", url, statusString)
		}
	}
	return true, errorsList

}

// checkServer performs the actual server check
func checkServer(url string) (int, string) {

	// Check URL format
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	resp, err := http.Get(url)
	if err != nil {
		return -1, err.Error()
	}
	defer resp.Body.Close()
	status := resp.Status

	if resp.StatusCode != 200 {
		return 1, status
	}

	return 0, status
}

func notifyAndExit(s string) {

	notify.Push("GoServerCheck: Error", s, "", notificator.UR_CRITICAL)
	os.Exit(-1)
}
