package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	dns "github.com/miles-w-3/ddns/internal/dns"
)

func main() {
	// Get the authorization token from environment variable
	client, err := dns.NewCloudflareClient("https://api.cloudflare.com/client")

	if client == nil || err != nil {
		fmt.Printf("Failed to initialize client: %v\n", err)
		return
	}

	result, err := client.GetCurrentIP()
	if err != nil {
		fmt.Printf("Failed to get current ip in DNS zone: %s\n", err.Error())
		return
	}

	watcher, err := dns.NewIPWatcher(result)
	if err != nil {
		fmt.Printf("Failed to initialize IP watcher: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the monitoring in a goroutine
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go watcher.Start(ctx)

	// Main loop to handle IP changes
	for {
		select {
		case newIP := <-watcher.IPChangeChannel():
			// Handle the IP change in the main thread
			fmt.Println("Main thread: Detected IP change to", newIP)
			err := client.UpdateIP(newIP)
			if err != nil {
				fmt.Printf("Failed to update target record: %v\n", err)
			}
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
			cancel() // Stop the monitor
			return
		}
	}
}
