package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func getIPAddress() ([]string, error) {
	// Retrieve network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Array to store IP addresses
	addresses := []string{}

	// Loop through each network interface
	for _, iface := range interfaces {
		// Interface should not be a loopback and should be up
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			// Interface addresses
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, err
			}

			// Loop through each address in the interface
			for _, addr := range addrs {
				// Convert network address to IP
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					return nil, err
				}

				// Check if the address is IPv4
				if ip.To4() != nil {
					addresses = append(addresses, ip.String())
				}
			}
		}
	}

	return addresses, nil
}

func getIpfsID(ipAddress string) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:5001/api/v0/id", ipAddress)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getClusterID() (*http.Response, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	url := ""
	switch env {
	case "development":
		url = "http://localhost:9094/id"
	default:
		url = "http://cluster-internal.io:9094/id"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
