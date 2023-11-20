package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"io/ioutil"
	"github.com/gin-gonic/gin"
)

func main() {

	/////////////////////routes bind with specific function behind it///////////////////
		//////////////////////////////////////////////////////
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
	// Implement the logic for the getStatus function
	// ...
	c.JSON(http.StatusOK, gin.H{"isClusterOnline": true})
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



///////////////////////Helper Functions are defined here////////////////
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

// getIpfsId fetches information from an IPFS node based on the given IP address.
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




///////////////////////// Function to test the helper functions////////////////
	//////////////////////////////////////////////////////
func testHelperFunctions() {
	// Call the function to get IP addresses
	ipAddresses, err := getIPAddress()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Print the retrieved IP addresses
	fmt.Println("IP Addresses:")
	for _, ip := range ipAddresses {
		fmt.Println(ip)
	}

	// Test getIpfsId
	ipfsNodeInformation, err := getIpfsId("localhost")
	if err != nil {
		// Handle the error, e.g., print it to the console
		fmt.Println("Error in getIpfsId:", err)
		return
	}

	// Print the information obtained from the IPFS node
	fmt.Println("IPFS Node Information:", ipfsNodeInformation)
}