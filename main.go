package main

import (
	"fmt"
	"bytes"
	"log"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"io/ioutil"
	"time"
	"github.com/gin-gonic/gin"
)
// Struct to represent the data sent in the POST request
type NodeDetails struct {
	IPAddress      string `json:"ipAddress"`
	IPFSClusterID  string `json:"ipfsClusterId"`
	IPFSID         string `json:"ipfsId"`
}
/////////////////////routes bind with specific function behind it///////////////////
		//////////////////////////////////////////////////////
func main() {

	
	router := gin.Default()

	router.GET("/api/file/node/status", getStatus)
	router.GET("/api/file/view/access-play/:accessKey/:token", playVideo)
	router.GET("/api/file/view/access/:accessKey/:token", getAccessFile)
	router.GET("/api/file/download/:accessKey/:token", downloadFile)

	if err := router.Run(":3008"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}
	//////////////////Test helper function////////////////
	//////////////////////////////////////////////////////
	testHelperFunctions()
	

	
}

///////////////////////Functions behind each API are defined here////////////////
		//////////////////////////////////////////////////////
	
func getStatus(c *gin.Context) {
	clusterID, err := getClusterID()
	if err != nil {
		// Handle the error as needed
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cluster ID"})
		return
	}

	// Convert the byte slice to a string
	clusterIDString := string(clusterID)

	// Check if clusterIDString is not empty
	isClusterOnline := clusterIDString != ""

	// Send the response
	c.JSON(http.StatusOK, gin.H{"isClusterOnline": isClusterOnline})
}

func playVideo(c *gin.Context) {
	// Implement the logic for the playVideo function
	// ...
	c.JSON(http.StatusOK, gin.H{"message": "Video playing"})
}

func getAccessFile(c *gin.Context) {
	// Implement the logic for the getAccessFile function
	// ...
	c.String(http.StatusOK, "File content")
}

func downloadFile(c *gin.Context) {
	// Implement the logic for the downloadFile function
	// ...
	c.String(http.StatusOK, "File content for download")
}

/////////////////////Helper Functions are defined here////////////////
		//////////////////////////////////////////////////////
	
//get ipaddress array
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

// getIpfsId fetches ID from an IPFS node based on the given IP address.
func getIpfsId(ipAddress string) (string, error) {
	// Construct the URL for the IPFS node's /api/v0/id endpoint
	url := fmt.Sprintf("http://%s:5001/api/v0/id", ipAddress)

	// Make an HTTP GET request to the IPFS node
	response, err := http.Get(url)
	if err != nil {
		// Return an empty string and the error if the request fails
		return "", err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// Return an empty string and the error if reading the body fails
		return "", err
	}

	// Convert the response body to a string and return it
	return string(body), nil
}

// Function to get ID from an IPFS cluster based on the environment
func getClusterID() ([]byte, error) {
	// Get the environment variable or default to 'development'
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Define the URL based on the environment
	var url string
	switch env {
	case "development":
		url = "http://localhost:9094/id"
	default:
		url = "http://cluster-internal.io:9094/id"
	}

	// Make an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Print the status URL to the console
	fmt.Println("Status URL:", url)

	// Return the response body
	return body, nil
}

// Function to  update ip address, ipfs cluster id, ipfs id
func updateNodeDetailsOnRecursiveCall(ipAddress, ipfsClusterId, ipfsId string) {
	// Create a JSON payload from the provided data
	payload := NodeDetails{
		IPAddress:     ipAddress,
		IPFSClusterID: ipfsClusterId,
		IPFSID:        ipfsId,
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Make an HTTP POST request to update node details
	url := fmt.Sprintf("%s/node/update-node-details", os.Getenv("API_SERVER_URL"))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		// Handle the HTTP request error
		fmt.Println("Error making HTTP request:", err)

		// Retry after 5 seconds
		time.Sleep(5 * time.Second)
		updateNodeDetailsOnRecursiveCall(ipAddress, ipfsClusterId, ipfsId)
		return
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error updating node details. Status:", resp.Status)

		// Retry after 5 seconds
		time.Sleep(5 * time.Second)
		updateNodeDetailsOnRecursiveCall(ipAddress, ipfsClusterId, ipfsId)
		return
	}

	// Log the result of the update operation
	fmt.Println("Node details updated successfully")
}
