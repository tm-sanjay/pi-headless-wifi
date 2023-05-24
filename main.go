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

type DeviceDetails struct {
	Interface string `json:"interface"`
	Mac       string `json:"mac"`
}

const (
	ip   = "localhost" // IP address to run the server on (10.42.0.1)
	port = "8080"      // Port to run the server on
)

var (
	deviceDetails DeviceDetails // Name of the WiFi device to use (wlx0ccf89299e08)
)

func main() {
	deviceDetails = getDeviceDetails()
	fmt.Println("Device Details:")
	fmt.Println("  Name", deviceDetails.Interface)
	fmt.Println("  MAC", deviceDetails.Mac)

	fmt.Println("Starting server...")
	fmt.Println(ip + ":" + port)

	// Handle the routes
	http.HandleFunc("/", home)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/wifilist", getWifiList)

	// Start the server
	log.Fatal(http.ListenAndServe(ip+":"+port, nil))
}

// --------------------Page Handlers--------------------
// Function to handle the home page
func home(w http.ResponseWriter, r *http.Request) {
	// Serve the index.html file from the root directory
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "index.html")
		fmt.Println("Home page served")
	} else {
		// If the URL doesn't match the root directory, return 404
		errorHandler(w, r, http.StatusNotFound)
	}
}

// Function to handle the form submission
func submit(w http.ResponseWriter, r *http.Request) {
	// Read the values of the wifiSSID and password fields from the request
	wifiSSID := r.FormValue("wifiSSID")
	wifiPSK := r.FormValue("wifiPSK")

	// Print the values to the console to confirm they are correct
	fmt.Println("SSID:", wifiSSID)
	fmt.Println("Password:", wifiPSK)

	// Connect to the specified WiFi network
	connectToWifi(wifiSSID, wifiPSK)

	// thanks for submitting text
	fmt.Fprintf(w, "Thanks for submitting!")

}

// Function to get the list of available WiFi networks and return them as JSON
func getWifiList(w http.ResponseWriter, r *http.Request) {

	// Set the CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Scan for available WiFi networks
	wifiList, err := scanWifiNetworks()
	if err != nil {
		log.Panic(err)
	}

	// Convert the slice of Wifi objects to JSON
	wifiListJson, err := json.Marshal(wifiList)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Wifi list served")
	fmt.Println(string(wifiListJson))

	// Write the JSON to the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(wifiListJson)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "Page not found")
		fmt.Println("Page not found")
	}
}

// --------------------Helper Functions--------------------

// Function to get the name of the WiFi device to use
func getDeviceDetails() DeviceDetails {
	// Get Interface
	inter, err := exec.Command("iw", "dev").Output()
	if err != nil {
		log.Panic(err)
	}
	interfaceString := string(inter)
	interfaceString = strings.Split(interfaceString, "Interface ")[1]
	interfaceString = strings.Split(interfaceString, "\n")[0]

	//remove if any spaces or newlines
	interfaceString = strings.TrimSpace(interfaceString)

	// Get MAC
	mac, err := exec.Command("cat", "/sys/class/net/"+interfaceString+"/address").CombinedOutput()

	if err != nil {
		log.Panic(err)
	}
	macString := string(mac)
	macString = strings.TrimSpace(macString)

	// Return DeviceDetails
	return DeviceDetails{Interface: interfaceString, Mac: macString}
}

func scanWifiNetworks() ([]WifiDetails, error) {
	cmd := exec.Command("iwlist", deviceDetails.Interface, "scan")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Process the output and extract the Wi-Fi network names
	networks := extractWifiNetworks(output)
	return networks, nil
}

func extractWifiNetworks(output []byte) []WifiDetails {
	// Example implementation assuming "iwlist" command output format:
	wifiDetailsList := make([]WifiDetails, 0)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ESSID:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				network := strings.Trim(parts[1], `" `)
				if network != "" {
					wifiDetailsList = append(wifiDetailsList, WifiDetails{Name: network})
				}
			}
		}
	}

	return wifiDetailsList
}

// Function to connect to the specified WiFi network
func connectToWifi(wifiSSID string, wifiPSK string) {
	// Switch back to station mode
	// exec.Command("nmcli", "radio", "wifi", "off").Run()
	// exec.Command("nmcli", "radio", "wifi", "on").Run()

	// Execute the nmcli command to connect to the specified WiFi network
	out, err := exec.Command("nmcli", "dev", "wifi", "connect", wifiSSID, "password", wifiPSK).CombinedOutput()
	if err != nil {
		log.Panic(err)
	}

	// Print the output to the console to confirm the connection was successful
	fmt.Println(string(out))

	//if connection was successful
	if strings.Contains(string(out), "successfully activated") {
		log.Println("Successfully connected to", wifiSSID)
	} else {
		log.Println("!Failed to connect to", wifiSSID)
	}
}
