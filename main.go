package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type WifiDetails struct {
	Name string `json:"name"`
}

func main() {
	fmt.Println("Starting server...")
	fmt.Println("localhost:8080")

	http.HandleFunc("/", home)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/wifilist", getWifiList)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

// Function to handle the home page
func home(w http.ResponseWriter, r *http.Request) {
	// Serve the index.html file from the root directory
	http.ServeFile(w, r, "index.html")
}

// Function to handle the form submission
func submit(w http.ResponseWriter, r *http.Request) {
	// Read the values of the wifiSSID and password fields from the request
	wifiSSID := r.FormValue("wifiSSID")
	wifiPSK := r.FormValue("wifiPSK")

	// Print the values to the console to confirm they are correct
	fmt.Println("SSID:", wifiSSID)
	fmt.Println("Password:", wifiPSK)

	// Redirect the user back to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Function to get the list of available WiFi networks and return them as JSON
func getWifiList(w http.ResponseWriter, r *http.Request) {
	// Execute the nmcli command to get the list of WiFi networks
	out, err := exec.Command("nmcli", "-f", "SSID", "dev", "wifi", "list").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Parse the output and create a slice of Wifi objects
	log.Println("-----Available WiFi Networks-----")
	wifiList := make([]WifiDetails, 0)
	lines := strings.Split(string(out), "\n")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			wifiList = append(wifiList, WifiDetails{Name: line})
			log.Println(line)
		}
	}

	// Marshal the slice to JSON and write to the response
	wifiListJson, err := json.Marshal(wifiList)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(wifiListJson)
}
