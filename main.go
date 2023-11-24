package main

import (
	"fmt"
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	//"os"
	//"io"
	"io/ioutil"
	//"time"
	//"time"
//	"sort"
	//"strconv"
	//"strings"
	"github.com/gin-gonic/gin"
	
)
// Struct to represent the data sent in the POST request
type NodeDetails struct {
	IPAddress      string `json:"ipAddress"`
	IPFSClusterID  string `json:"ipfsClusterId"`
	IPFSID         string `json:"ipfsId"`
}


type Allocation struct {
	Allocations string `json:"allocations"`
	CID         string   `json:"cid"`
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	Size        float64  `json:"size"`
}

type FileMetaData struct {
	Allocations []string `json:"allocations"`
	CID         string             `json:"cid"`
	Index       int                `json:"index"`
	Name        string             `json:"name"`
	Size        float64            `json:"size"`
}

type AccessData struct {
	__v             int    `json:"__v"`
	_id            string`json:"_id"`
	AccessKey      string `json:"accessKey"`
	AccessType     string `json:"accessType"`
	AccessUserEmail string `json:"accessUserEmail"`
	AccessUserId   string `json:"accessUserId"`
	BucketId       string `json:"bucketId"`
	CreatedAt      string `json:"createdAt"`
	Data           string `json:"data"` 
	FileId         string `json:"fileId"`
	FileMetaData   []FileMetaData `json:"fileMetaData"`
	FileName        string           `json:"fileName"`
	FileSize        float64          `json:"fileSize"`
	FileType        string           `json:"fileType"`
	IV              string           `json:"iv"`
	ObjectType      string           `json:"objectType"`
	Salt            string           `json:"salt"`
	SecretKey       string           `json:"secretKey"`
	SharedStatus    bool             `json:"sharedStatus"`
	Status          string           `json:"status"`
	TokenSalt       string           `json:"tokenSalt"`
	UpdatedAt       string           `json:"updatedAt"`
	UserAvatar      string           `json:"userAvatar"`
	UserID          string           `json:"userId"`
	UserName        string           `json:"userName"`
}

type Response struct {
	Data AccessData `json:"data"`
	Message       string `json:"message"`
	Status        int    `json:"status"`
	Success       bool   `json:"success"`
}
/////////////////////routes bind with specific function behind it///////////////////
		//////////////////////////////////////////////////////
func main() {

	router := gin.Default()

	router.GET("/api/file/node/status", getStatus)
	router.GET("/api/file/view/access-play/:accessKey", playVideo)
	router.GET("/api/file/view/access-play/:accessKey/:token", playVideo)
	

	if err := router.Run(":3008"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}
	
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
	AccessDataResponse, err := verifyAccessToken(accessKey, token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	} else {
		fmt.Println("AccessDataResponse value:", AccessDataResponse)
	}

	// Convert the map to JSON
	jsonData, err := json.Marshal(AccessDataResponse)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Unmarshal the JSON data into the struct
	var response Response
	err = json.Unmarshal(jsonData, &response)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Now 'response' should contain your structured data
	fmt.Printf("%+v\n", response)

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
//func verifyAccessToken(accessKey, token string) (*AccessDataResponse, error){
func verifyAccessToken(accessKey, token string) (map[string]interface{}, error) {
	// Define the request payload
	requestData := map[string]string{"accessKey": accessKey, "token": token}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	// Send a request to verify the access token
	resp, err := http.Post(
		"https://storagechain-be.invo.zone/api/file/access/verify-token",
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
	// var response AccessDataResponse
    // if err := json.Unmarshal(responseBody, &response); err != nil {
    //     return nil, err
    // }

	return responseData, nil
    //return &response, nil
}