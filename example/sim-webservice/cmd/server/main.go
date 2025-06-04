package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"sim-webservice/pkg/handlers"
	"sim-webservice/pkg/simconnect"
)

func main() {
	// Parse command line flags
	dllPath := flag.String("dll", "", "Path to SimConnect.dll (optional, uses default if not specified)")
	flag.Parse()

	fmt.Println("✈️ MSFS Sim WebService - Starting...")
	fmt.Println("   Real-time simulator data monitoring for Microsoft Flight Simulator")
	fmt.Println("   Open your browser to http://localhost:8080")
	fmt.Println()
	// Initialize SimConnect monitor client with optional DLL path
	var monitorClient *simconnect.MonitorClient
	if *dllPath != "" {
		fmt.Printf("   🔧 Using custom DLL path: %s\n", *dllPath)
		monitorClient = simconnect.NewMonitorClientWithDLL(*dllPath)
	} else {
		fmt.Println("   🔧 Using default DLL path")
		monitorClient = simconnect.NewMonitorClient()
	}

	if err := monitorClient.Connect(); err != nil {
		log.Fatalf("❌ Failed to initialize SimConnect: %v", err)
	}
	defer monitorClient.Close()
	// Initialize handlers
	monitorHandler := handlers.NewMonitorHandler(monitorClient) // Set up HTTP routes
	http.HandleFunc("/", monitorHandler.HandleIndex)
	http.HandleFunc("/api/monitor", monitorHandler.HandleMonitorAPI)
	http.HandleFunc("/api/camera", monitorHandler.HandleCameraStateToggle)
	http.HandleFunc("/api/external-power", monitorHandler.HandleExternalPowerToggle)
	http.HandleFunc("/api/battery1", monitorHandler.HandleBattery1Toggle)
	http.HandleFunc("/api/battery2", monitorHandler.HandleBattery2Toggle)
	http.HandleFunc("/api/system", monitorClient.GetSystemEventsHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Start the web server
	fmt.Println("🚀 Starting web server on http://localhost:8080")
	fmt.Println("   🌍 Phase 1: Environmental monitoring (12 variables)")
	fmt.Println("   📊 Environmental Data: Temperature, Pressure, Wind Speed/Direction")
	fmt.Println("   🌧️ Conditions: Visibility, Precipitation, Density Altitude, Ground Elevation")
	fmt.Println("   🧭 Additional: Magnetic Variation, Sea Level Pressure, Air Density")
	fmt.Println("   🔄 Updates every second via SimConnect")
	fmt.Println()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Failed to start web server: %v", err)
	}
}
