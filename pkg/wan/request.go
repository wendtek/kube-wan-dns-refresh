package wan

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// GetWanIP will return the current WAN IP address by making a request to a public IP service
func GetWanIP() (string, error) {
	// Make request to https://icanhazip.com
	resp, err := http.Get("https://icanhazip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Convert the body to a string and remove any trailing whitespace
	ip := strings.TrimSpace(string(body))

	// Validate that the body is an IP address
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid WAN address returned: %s", ip)
	}

	return ip, nil
}
