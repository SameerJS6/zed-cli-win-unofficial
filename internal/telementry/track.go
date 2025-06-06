package telementry

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sys/windows"
)

const (
	API_URL = "https://zedcli.sameerjs.com/api/phog"
	TIMEOUT = 3 * time.Second
)

var (
	sessionID string
	userID    string
	osVersion string
)

func init() {
	// Generate session ID for this CLI run
	sessionID = uuid.New().String()

	// Generate anonymous user ID based on machine
	hostname, _ := os.Hostname()
	username := os.Getenv("USERNAME")
	machineID := fmt.Sprintf("%s-%s", hostname, username)
	userID = fmt.Sprintf("%x", md5.Sum([]byte(machineID)))[:16]

	// Get detailed OS version
	osVersion = getOSVersion()
}

// getOSVersion returns detailed OS version information
func getOSVersion() string {
	if runtime.GOOS != "windows" {
		return runtime.GOOS
	}

	// Get Windows version information
	version := windows.RtlGetVersion()

	// Map version numbers to friendly names
	switch {
	case version.MajorVersion == 10 && version.BuildNumber >= 22000:
		return "Windows 11"
	case version.MajorVersion == 10 && version.BuildNumber >= 10240:
		return "Windows 10"
	case version.MajorVersion == 6 && version.MinorVersion == 3:
		return "Windows 8.1"
	case version.MajorVersion == 6 && version.MinorVersion == 2:
		return "Windows 8"
	case version.MajorVersion == 6 && version.MinorVersion == 1:
		return "Windows 7"
	default:
		return fmt.Sprintf("Windows %d.%d (Build %d)",
			version.MajorVersion, version.MinorVersion, version.BuildNumber)
	}
}

type Event struct {
	Event      string         `json:"event"`
	Properties map[string]any `json:"properties"`
}

// TrackEvent sends analytics data asynchronously (non-blocking)
func TrackEvent(event string, properties map[string]any) {
	// Add base properties to all events
	enrichedProperties := enrichProperties(properties)

	payload := Event{
		Event:      event,
		Properties: enrichedProperties,
	}

	// Send asynchronously to avoid blocking CLI
	go sendEvent(payload)
}

// enrichProperties adds base properties to every event
func enrichProperties(properties map[string]any) map[string]any {
	base := map[string]any{
		"cli_version": "v1.0.0",
		"os":          runtime.GOOS,
		"os_version":  osVersion,
		"arch":        runtime.GOARCH,
		"session_id":  sessionID,
		"user_id":     userID,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	// Merge user properties with base properties
	for k, v := range properties {
		base[k] = v
	}

	return base
}

// sendEvent sends the event with timeout and error handling
func sendEvent(event Event) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return // Silently fail
	}

	// Create request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	fmt.Println(string(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "zed-cli/v1.0.0")

	client := &http.Client{
		Timeout: TIMEOUT,
	}

	// Send request and ignore response/errors
	// Analytics should never interfere with CLI functionality
	client.Do(req)
}
