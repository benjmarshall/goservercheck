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

	notify = notificator.New(notificator.Options{AppName: "GoServerCheck"})

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(usr.HomeDir + "/.goservercheck")
	if err != nil {
		notifyAndExit("Can't find ~/.goservecheck")
	}

	var serverList []string
	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		serverList = append(serverList, fileScanner.Text())
	}

	var errorsList string

	for _, url := range serverList {
		status, statusString := checkServer(url)
		if status == -1 {
			notifyAndExit(fmt.Sprintf("Error occured connecting to %s:\n%v", url, statusString))
		} else if status == 1 {
			errorsList = errorsList + fmt.Sprintf("%s, appears to be down. Resonse was:%s\n", url, statusString)
		}
	}
	if errorsList != "" {
		notify.Push("GoServerCheck: Problems Detected", errorsList, "", notificator.UR_CRITICAL)
	} else {
		notify.Push("GoServerCheck: Servers OK", "All web pages reached.", "", notificator.UR_NORMAL)
	}

}

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
