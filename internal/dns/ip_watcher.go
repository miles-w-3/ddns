package dns

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// IPWatcher monitors changes to the public IP address
type IPWatcher struct {
	interval     time.Duration
	ipChangeChan chan string
	url          string
	client       *http.Client
	currentIP    string
}

// creates a new IP monitor
func NewIPWatcher(startingIP string) (*IPWatcher, error) {
	intervalStr := os.Getenv("POLL_MINUTE_INTERVAL")
	intervalMin := 1
	if intervalStr != "" {
		value, err := strconv.Atoi(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("Failed to read invalid minute interval %s", intervalStr)
		}
		if intervalMin < 1 {
			return nil, fmt.Errorf("Cannot specify invalid poll duration %d", intervalMin)
		}
		intervalMin = value
	}
	fmt.Printf("Interval minutes is %d\n", intervalMin)
	return &IPWatcher{
		interval:     (time.Minute * time.Duration(intervalMin)),
		ipChangeChan: make(chan string, 1), // Buffered channel to prevent blocking
		url:          "https://api.ipify.org",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		currentIP: startingIP,
	}, nil
}

// FetchIP fetches the current public IP from the API
func (m *IPWatcher) FetchIP() (string, error) {
	resp, err := m.client.Get(m.url)
	if err != nil {
		return "", fmt.Errorf("error fetching IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	ip := strings.TrimSpace(string(body))
	return ip, nil
}

// Start begins the IP monitoring process
func (m *IPWatcher) Start(ctx context.Context) {
	log.Println("IP watching started with interval:", m.interval)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("IP monitoring stopped")
			return
		case <-ticker.C:
			newIP, err := m.FetchIP()
			if err != nil {
				log.Println("Error fetching IP:", err)
				continue
			}

			// Check if the IP has changed
			if m.currentIP != newIP {
				log.Printf("IP changed: %s -> %s", m.currentIP, newIP)
				m.currentIP = newIP // Update the local currentIP

				// Send the new IP to the channel (non-blocking)
				select {
				case m.ipChangeChan <- newIP:
					// Successfully sent
				default:
					// Channel is full, but we don't want to block
					log.Println("Warning: IP change notification channel is full")
				}
			}
		}
	}
}

// IPChangeChannel returns the channel that notifies about IP changes
func (m *IPWatcher) IPChangeChannel() <-chan string {
	return m.ipChangeChan
}
