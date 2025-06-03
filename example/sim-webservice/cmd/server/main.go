package main

import (
	"fmt"
	"log"
	"net/http"

	"sim-webservice/pkg/handlers"
	"sim-webservice/pkg/simconnect"
)

func main() {
	fmt.Println("✈️ MSFS Sim WebService - Starting...")
	fmt.Println("   Real-time simulator data monitoring for Microsoft Flight Simulator")
	fmt.Println("   Open your browser to http://localhost:8080")
	fmt.Println()

	// Initialize SimConnect weather client
	weatherClient := simconnect.NewWeatherClient()
	if err := weatherClient.Connect(); err != nil {
		log.Fatalf("❌ Failed to initialize SimConnect: %v", err)
	}
	defer weatherClient.Close()

	// Initialize handlers
	weatherHandler := handlers.NewWeatherHandler(weatherClient) // Set up HTTP routes
	http.HandleFunc("/", weatherHandler.HandleIndex)
	http.HandleFunc("/api/weather", weatherHandler.HandleWeatherAPI)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Start the web server
	fmt.Println("🚀 Starting web server on http://localhost:8080")
	fmt.Println("   🌍 Phase 1: Environmental monitoring (12 variables)")
	fmt.Println("   📊 Core Weather: Temperature, Pressure, Wind Speed/Direction")
	fmt.Println("   🌧️ Conditions: Visibility, Precipitation, Density Altitude, Ground Elevation")
	fmt.Println("   🧭 Additional: Magnetic Variation, Sea Level Pressure, Air Density")
	fmt.Println("   🔄 Updates every second via SimConnect")
	fmt.Println()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Failed to start web server: %v", err)
	}
}
