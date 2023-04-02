package main

import (
	"fmt"
	"net/http"
)


func main() {
	fmt.Println("Starting server...")
	
	http.HandleFunc("/", home)
	http.HandleFunc("/submit", submit)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
	
}

func home(w http.ResponseWriter, r *http.Request) {
	// Serve the index.html file from the root directory
	http.ServeFile(w, r, "index.html")
}

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
