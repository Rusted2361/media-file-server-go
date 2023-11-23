package main

import (
	"fmt"
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"io"
	"io/ioutil"
	"time"
	"log"
	"sort"
	"strconv"
	"strings"
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

	//////uncomment this function to immediately shut service if cluster id not found
	// Create a new Goroutine for the heartBeat function
	//go heartBeat()

	router := gin.Default()

	router.GET("/api/file/node/status", getStatus)
	router.GET("/api/file/view/access-play/:accessKey/:token", playVideo)
	router.GET("/api/file/view/access/:accessKey/:token", getAccessFile)
	router.GET("/api/file/download/:accessKey/:token", downloadFile)

	if err := router.Run(":3008"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}
	///////////uncomment this function to enable heartbeat function with 5 second delay will not immediately shut down/////////////
	// Run the Gin server on port 3008 in a Goroutine
	// go func() {
	// 	if err := router.Run(":3008"); err != nil {
	// 		fmt.Println("Failed to start the server:", err)
	// 	}
	// }()

	// // Wait for 5 seconds
	// <-time.After(5 * time.Second)

	// // Start the heartBeat function after 5 seconds
	// go heartBeat()

	// // Block the main Goroutine so that the program doesn't exit
	// select {}
	
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
	
	// Extract access key and token from URL parameters
	accessKey := c.Param("accessKey")
	token := c.Param("token")

	// Verify the access token
	accessData, err := verifyAccessToken(accessKey, token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ipfsMetaData := accessData.FileMetaData
	sort.Slice(ipfsMetaData, func(i, j int) bool {
		return ipfsMetaData[i].Index < ipfsMetaData[j].Index
	})

	path := fmt.Sprintf("videos/%s%s", accessData.AccessKey, accessData.FileName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		writableStream, err := os.Create(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create writable stream",
			})
			return
		}

		for _, meta := range ipfsMetaData {
			fileResponse, err := http.Get(fmt.Sprintf("http://46.101.133.110:8080/api/v0/cat/%s", meta.CID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to fetch file response from IPFS",
				})
				return
			}

			fileBytes, err := io.ReadAll(fileResponse.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to read file response body",
				})
				return
			}

			decryptedData := decryptedSecretKeyAndFile(accessData.Data, accessData.SecretKey, accessData.AccessKey, accessData.IV, fileBytes, accessData.Salt)

			_, err = writableStream.Write(decryptedData)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to write decrypted data to the writable stream",
				})
				return
			}
		}

		writableStream.Close()

		stat, err := os.Stat(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get file statistics",
			})
			return
		}

		fileSize := stat.Size()
		rangeHeader := c.GetHeader("Range")

		if rangeHeader != "" {
			handleByteRange(c, path, fileSize)
		} else {
			handleFullContent(c, path, fileSize)
		}
	} else {
		handleExistingFile(c, path)
	}
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
func getIpfsId(ipAddress ...string) (string, error) {
	// Construct the URL for the IPFS node's /api/v0/id endpoint
	var url string

	if len(ipAddress) > 0 {
		url = fmt.Sprintf("http://%s:9094/id", ipAddress[0])
	} else {
		url = "http://localhost:9094/id"
	}

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
func getClusterID(ipAddress ...string) ([]byte, error) {
	
	// if ipAddress == "" {
	// 	ipAddress = "localhost"
	// }
	// // Construct the URL
	// url := fmt.Sprintf("http://%s:9094/id", ipAddress)

	var url string

	if len(ipAddress) > 0 {
		url = fmt.Sprintf("http://%s:9094/id", ipAddress[0])
	} else {
		url = "http://localhost:9094/id"
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


func decryptedSecretKeyAndFile(data, secretKey, accessKey, iv []byte, fileResponse []byte, salt string) []byte {
	// Replace with your decryption logic
	return fileResponse
}

func verifyAccessToken(accessKey, token string) (map[string]interface{}, error) {
	// Define the request payload
	requestData := map[string]string{"accessKey": accessKey, "token": token}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	// Make the HTTP POST request
	resp, err := http.Post(
		"http://your-api-server/file/access/verify-token",
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var responseData map[string]interface{}
	err = json.Unmarshal(responseBody, &responseData)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func getAccessDataResponse(accessKey, token string) AccessDataResponse {
	// Replace with your logic to get access data from token verification
	return AccessDataResponse{}
}

func handleByteRange(c *gin.Context, path string, fileSize int64) {
	rangeHeader := c.GetHeader("Range")
	parts := strings.Split(strings.ReplaceAll(rangeHeader, "bytes=", ""), "-")
	start, _ := strconv.ParseInt(parts[0], 10, 64)
	end, _ := strconv.ParseInt(parts[1], 10, 64)
	chunkSize := end - start + 1

	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer file.Close()

	c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Writer.Header().Set("Accept-Ranges", "bytes")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", chunkSize))
	c.Writer.Header().Set("Content-Type", "video/mp4")
	c.Writer.WriteHeader(http.StatusPartialContent)

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to seek file",
		})
		return
	}

	io.CopyN(c.Writer, file, chunkSize)
}

func handleFullContent(c *gin.Context, path string, fileSize int64) {
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Writer.Header().Set("Content-Type", "video/mp4")
	c.Writer.WriteHeader(http.StatusOK)

	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer file.Close()

	io.Copy(c.Writer, file)
}

func handleExistingFile(c *gin.Context, path string) {
	// Functionality for streaming and response of an existing file
}





////////////////////////////Recursive functions///////////////////////
		//////////////////////////////////////////////////////
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

// this will recursively check for clusterid and ipfs id
func heartBeat() {
	for {
		// Check the local IPFS Cluster and IPFS node status
		clusterResponseLocal, _ := getClusterID()
		ipfsResponseLocal, _ := getIpfsId()

		// If either local IPFS Cluster or IPFS node is not running, exit the application
		if len(clusterResponseLocal) == 0 || len(ipfsResponseLocal) == 0 {
			fmt.Println("Ipfs Cluster or Ipfs is not running locally.")
			//exit
			os.Exit(1)

		}


		// Check the global (online) IPFS Cluster and IPFS node status
		clusterResponseOnline, _ := getClusterID()
		ipfsResponseLocalOnline, _ := getIpfsId()

		// If either global IPFS Cluster or IPFS node is not running, exit the application
		if len(clusterResponseOnline) == 0 || len(ipfsResponseLocalOnline) == 0 {
			fmt.Println("Ipfs Cluster or Ipfs is not running globally.")
			//exit
			os.Exit(1)

		}

		// Display a message in the terminal
		log.Print("Heartbeat check completed. Waiting for the next check...")
		// Sleep for 5 seconds before the next heartbeat
		time.Sleep(5 * time.Second)
	}
}








