package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	ApName = "SS-AP"
	ApPsk  = "12345678"
)

func switchToAPMode() {
	// Check if the device is already in AP mode
	if checkIfAPMode() {
		fmt.Println("Device is already in AP mode")
		return
	}

	fmt.Println("Switching to AP mode...")

	if !checkIfApNameRegistered() {
		fmt.Println("Registering AP name...")
		out, err := exec.Command("nmcli", "device", "wifi", "hotspot", "ssid", ApName, "password", ApPsk).Output()
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(string(out))
		// rename the connection
		exec.Command("nmcli", "connection", "modify", "Hotspot", "connection.id", ApName).Output()
	}

	fmt.Println("Starting AP...")
	out, err := exec.Command("nmcli", "connection", "up", ApName).Output()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(out))

	fmt.Println("Device switched to AP mode")
}

// Function to check if the device is already in AP mode
func checkIfAPMode() bool {
	out, err := exec.Command("iwconfig", deviceDetails.Interface).CombinedOutput()
	if err != nil {
		log.Panic(err)
	}
	if strings.Contains(string(out), "Mode:Master") {
		return true
	}
	return false
}

// Function to check if ap name is already registered
func checkIfApNameRegistered() bool {
	out, err := exec.Command("nmcli", "con", "show", ApName).CombinedOutput()
	// fmt.Println(string(out))
	if err != nil {
		if strings.Contains(string(out), "no such connection") {
			return false
		} else {
			log.Panic(err)
		}
	}
	if strings.Contains(string(out), "connection.id:                          "+ApName) {
		fmt.Println("AP name already registered")
		return true
	}
	return false
}
