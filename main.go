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
	ok, errorsList := runChecks(*serverList, 5)

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
func runChecks(serverList serverListType, numWorkers int) (bool, string) {

	var errorsList string

	chURL := make(chan string, len(serverList.sl))
	chResp := make(chan serverResponse, len(serverList.sl))

	for i := 0; i < numWorkers; i++ {
		go checkServer(chURL, chResp)
	}

	for _, url := range serverList.sl {
		chURL <- url
	}

	close(chURL)

	for i := 0; i < len(serverList.sl); i++ {
		response := <-chResp
		if response.returnCode == -1 {
			errorsList = errorsList + fmt.Sprintf("Error occured connecting to %s:\n%v", response.url, response.status)
			return false, errorsList
		} else if response.returnCode == 1 {
			errorsList = errorsList + fmt.Sprintf("%s, appears to be down. Resonse was:%s\n", response.url, response.status)
		}
	}

	close(chResp)

	return true, errorsList

}

// checkServer performs the actual server check
func checkServer(chURL chan string, chResp chan serverResponse) {

	for url := range chURL {

		// Check URL format
		fullURL := url
		if !strings.HasPrefix(url, "http") {
			fullURL = "http://" + fullURL
		}

		resp, err := http.Get(fullURL)
		if resp == nil || err != nil {
			chResp <- serverResponse{url, -1, err.Error()}
		} else {
			defer resp.Body.Close()
			status := resp.Status
			if resp.StatusCode != 200 {
				chResp <- serverResponse{url, 1, status}
			} else {
				chResp <- serverResponse{url, 0, status}
			}
		}

	}
}

func notifyAndExit(s string) {

	notify.Push("GoServerCheck: Error", s, "", notificator.UR_CRITICAL)
	os.Exit(-1)
}
