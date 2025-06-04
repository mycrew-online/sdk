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
	defer monitorClient.Close() // Initialize handlers
	monitorHandler := handlers.NewMonitorHandler(monitorClient)

	// Set up HTTP routes
	http.HandleFunc("/", monitorHandler.HandleIndex)
	http.HandleFunc("/api/monitor", monitorHandler.HandleMonitorAPI)
	http.HandleFunc("/api/camera", monitorHandler.HandleCameraStateToggle)
	http.HandleFunc("/api/external-power", monitorHandler.HandleExternalPowerToggle)
	http.HandleFunc("/api/battery1", monitorHandler.HandleBattery1Toggle)
	http.HandleFunc("/api/battery2", monitorHandler.HandleBattery2Toggle)
	http.HandleFunc("/api/apu-master", monitorHandler.HandleApuMasterSwitchToggle)
	http.HandleFunc("/api/apu-start", monitorHandler.HandleApuStartButtonToggle)
	http.HandleFunc("/api/aircraft-exit", monitorHandler.HandleAircraftExitToggle)
	http.HandleFunc("/api/cabin-no-smoking", monitorHandler.HandleCabinNoSmokingToggle)
	http.HandleFunc("/api/cabin-seatbelts", monitorHandler.HandleCabinSeatbeltsToggle)
	http.HandleFunc("/api/cabin-no-smoking-set", monitorHandler.HandleCabinNoSmokingSet)
	http.HandleFunc("/api/cabin-seatbelts-set", monitorHandler.HandleCabinSeatbeltsSet)
	http.HandleFunc("/api/system", monitorClient.GetSystemEventsHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // Start the web server
	fmt.Println("🚀 Starting web server on http://localhost:8081")
	fmt.Println("   ✈️ Comprehensive Flight Monitoring Suite (36+ variables)")
	fmt.Println("   🌤️ Environmental: Temperature, Pressure, Wind, Visibility, Precipitation")
	fmt.Println("   🧭 Navigation: Position, Altitude, Heading, Speed, GPS Data")
	fmt.Println("   ⚡ Systems: Camera Control, External Power, Battery Status, APU")
	fmt.Println("   ⏰ Simulation: Time Data, Simulation Rate, Flight Status")
	fmt.Println("   🔄 Real-time updates every second via SimConnect")
	fmt.Println()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Failed to start web server: %v", err)
	}
}
