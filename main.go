package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type WifiDetails struct {
	Name string `json:"name"`
}

type DeviceDetails struct {
	Interface string `json:"interface"`
	Mac       string `json:"mac"`
}

type Mode string

const (
	ip          = "10.42.0.1" // IP address to run the server on (10.42.0.1)
	port        = "8080"      // Port to run the server on
	envFilePath = ".env"
)

var (
	deviceDetails           DeviceDetails // Name of the WiFi device to use (wlx0ccf89299e08)
	wasConnectedToNet       bool          // Whether the device was connected to wifi before
	wasNeverConnectedToWifi bool          // Whether the device was never connected to wifi before
	isConnected             bool          // Whether the device is currently connected to the internet
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

	readEnv()
	switchToAPMode()

	// TODO:Start and stop the server based on AP mode else it will crash
	// Start the server
	go initServer()

	switchBetweenModes() // Start the goroutine to switch between AP and STA modes
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

	return DeviceDetails{Interface: interfaceString, Mac: macString}
}

// Function to read environment variables
func readEnv() {
	// Read the environment variables from the .env file if not found create one
	// Check if the .env file exists
	_, err := os.Stat(envFilePath)
	if os.IsNotExist(err) {
		// Create a new .env file with default values
		err = createDefaultEnvFile(envFilePath)
		if err != nil {
			log.Panic("Error creating .env file:", err)
		}
	}

	// Load the environment variables
	err = godotenv.Load(envFilePath)
	if err != nil {
		log.Panic("Error loading .env file:", err)
	}

	saved_ssid := os.Getenv("WIFI_SSID")
	// saved_psk := os.Getenv("WIFI_PASSWORD")

	// log.Println("SSID: " + saved_ssid)
	// log.Println("PSK: " + saved_psk)

	if saved_ssid == "" {
		fmt.Println("No saved SSID")
		wasNeverConnectedToWifi = true
		wasConnectedToNet = false
	} else {
		fmt.Println("Saved SSID")
		wasNeverConnectedToWifi = false
		wasConnectedToNet = true
	}
}

// Function to create a default .env file
func createDefaultEnvFile(filepath string) error {
	fmt.Println("Creating default .env file")
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write default environment variable values to the file
	defaultEnvContent := []byte(
		`WIFI_SSID=
WIFI_PASSWORD=
`,
	)

	_, err = file.Write(defaultEnvContent)
	if err != nil {
		return err
	}

	return nil
}

// Function to switch between AP and STA modes
func switchBetweenModes() {
	// while statement to keep the program running
	for {
		//wait for 15 seconds before checking for internet connection
		time.Sleep(1 * time.Second)
		isConnected = checkForInternet()
		isEthernetConnected()
		if !isConnected {
			// If the device is not connected to the internet, switch to AP mode
			if wasNeverConnectedToWifi {
				log.Println("Was never connected to wifi")
				switchToAPMode()
			} else {
				if wasConnectedToNet {
					log.Println("Scanning for previously saved wifi")
					switchToSTAMode()
					wasConnectedToNet = false
				} else {
					log.Println("Was connected to wifi but lost connection")
					switchToAPMode()
					wasConnectedToNet = true //helps to switch back to STA mode when internet is back
				}
			}
		} else {
			wasConnectedToNet = true
		}
		//check for internet connection every 5 minutes
		time.Sleep(5 * time.Minute)
	}
}

func scanWifiNetworks() ([]WifiDetails, error) {
	cmd := exec.Command("iwlist", deviceDetails.Interface, "scan")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Process the output and extract the Wi-Fi network names
	networks := extractWifiNetworks(output)
	return networks, nil
}

func extractWifiNetworks(output []byte) []WifiDetails {
	//Extract the wifi networks from the output
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
	exec.Command("nmcli", "connection", "down", ApName)
	// //restart the network manager
	// exec.Command("systemctl", "restart", "NetworkManager.service").Run()
	// //scan for wifi networks
	// exec.Command("nmcli", "dev", "wifi", "rescan").Run()
	// Execute the nmcli command to connect to the specified WiFi network
	out, err := exec.Command("nmcli", "dev", "disconnect", deviceDetails.Interface).CombinedOutput()
	// fmt.Println(string(out))
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(string(out))
	time.Sleep(2 * time.Second)
	// Create the command to execute the shell script
	cmd := exec.Command("/bin/bash", "-c", "./connect_wifi.sh")
	cmd.Stdin = os.Stdin

	// Set the environment variables for SSID and PSK
	cmd.Env = append(os.Environ(), "SSID="+wifiSSID, "PSK="+wifiSSID)

	// Capture the output of the command
	out, err = cmd.Output()
	if err != nil {
		log.Println("Failed to execute the shell script:", err)
		log.Println("Out:", string(out))
	}

	// Print the captured output
	log.Println("Shell script output:")
	log.Println(string(out))

	//if connection was successful
	if strings.Contains(string(out), "successfully activated") {
		log.Println("Successfully connected to", wifiSSID)
		//auto connect to this network on boot
		exec.Command("nmcli", "con", "modify", wifiSSID, "connection.autoconnect", "yes").Run()
		//save the ssid and psk to the .env file
		env, err := godotenv.Unmarshal("WIFI_SSID=" + wifiSSID + "\nWIFI_PASSWORD=" + wifiPSK)
		if err != nil {
			log.Panic("Error unmarshalling .env file:", err)
		}
		err = godotenv.Write(env, envFilePath)
		if err != nil {
			log.Panic("Error writing .env file:", err)
		}

		wasConnectedToNet = true
		wasNeverConnectedToWifi = false
	} else {
		log.Println("!Failed to connect to", wifiSSID)
		wasConnectedToNet = false
	}
}

// Function to check for internet connection
func checkForInternet() bool {
	out, err := exec.Command("ping", "-c", "1", "google.com").Output()
	if err != nil {
		log.Panic(err)
	}
	// fmt.Println(string(out))
	if strings.Contains(string(out), "1 received") {
		fmt.Println("Internet is available")
		return true
	} else {
		fmt.Println("Internet is not available")
		return false
	}
}

// Function to switch to AP mode
func switchToSTAMode() {
	// Switch to STA mode
	fmt.Println("Switching to STA mode")
	// stop the access point
	exec.Command("nmcli", "connection", "down", ApName).Run()

	exec.Command("nmcli", "radio", "wifi", "on").Run()
	//restart network manager
	exec.Command("systemctl", "restart", "NetworkManager").Run()
}

// Function to check if ethernet is connected
func isEthernetConnected() bool {
	// ifconfig eth1 up
	out, err := exec.Command("ifconfig", "eth1").CombinedOutput()
	// log.Println(string(out))
	if err != nil {
		log.Println(err)
	}
	// fmt.Println(string(out))
	if strings.Contains(string(out), "inet") {
		fmt.Println("Ethernet is connected")
		return true
	} else {
		fmt.Println("Ethernet is not connected")
		return false
	}
}
