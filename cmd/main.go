package main

import "github.com/mycrew-online/sdk/pkg/client"

func main() {
	// This is the main entry point for the application.
	// Here we would typically initialize the client, connect to the SimConnect server,
	// and start processing events or commands.
	// For now, we will just print a message to indicate that the application has started.
	println("SimConnect client application starting.")

	sdk := client.New("MySimConnectClient")
	defer sdk.Close()

	if err := sdk.Open(); err != nil {
		println("Failed to open SimConnect client:", err.Error())
		return
	}
}
