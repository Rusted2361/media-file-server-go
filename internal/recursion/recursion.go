package recursion

import (
	"time"
	"fmt"
	"path/filepath"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"log"
	"media-file-server-go/internal/helpers"
)
////////////////////////////Recursive functions///////////////////////
		//////////////////////////////////////////////////////
/////Data Structures//////
type NodeDetailsResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    struct {
		IPFSID        string `json:"ipfsId"`
		IPFSClusterID string `json:"ipfsClusterId"`
		IPAddress     string `json:"ipAddress"`
	} `json:"data"`
	Status int `json:"status"`
}
type UpdateNodeDetailsRequest struct {
	IPAddress      string `json:"ipAddress"`
	IPFSClusterID  string `json:"ipfsClusterId"`
	IPFSID         string `json:"ipfsId"`
}

/////Constants//////
const retryDelay = 5 * time.Second;
const maxRetries = 2;
const hostURL = "https://storagechain-be.invo.zone/api";
//const hostURL = "https://api.storagechain.io/api";

// this will recursively check for clusterid and ipfs id
func HeartBeat() {
	for {
		// Check the local IPFS Cluster and IPFS node status
		clusterResponseLocal, _ := helpers.GetClusterID()
		ipfsResponseLocal, _ := helpers.GetIpfsId()

		// If either local IPFS Cluster and IPFS node is not running, exit the application
		if len(clusterResponseLocal) == 0 || len(ipfsResponseLocal) == 0 {
			log.Print("🌐 Ipfs Cluster or Ipfs is not running locally.")
			//exit
			os.Exit(1)

		}

		ipaddress, err := helpers.GetIPAddress()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		// Check the global (online) IPFS Cluster and IPFS node status
		clusterResponseOnline, _ := helpers.GetClusterID(ipaddress)
		ipfsResponseLocalOnline, _ := helpers.GetIpfsId(ipaddress)

		// If either global IPFS Cluster and IPFS node is not running, exit the application
		if len(clusterResponseOnline) == 0 || len(ipfsResponseLocalOnline) == 0 {
			log.Print("🌐 Ipfs Cluster or Ipfs is not running globally.")
			//exit
			os.Exit(1)

		}

		interval := 15
		// Display a message in the terminal
		log.Print("💌🧚‍♀️💗🌨🥡🍥 Heartbeat check completed. Waiting for the next check after " + fmt.Sprintf("%v", interval) + " seconds...")
		// Sleep for 5 seconds before the next heartbeat
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

//Savenode details to DB
func SaveNodeDetails(retries int) {
	if retries == maxRetries {
		log.Print("❌ Retries", maxRetries, "times but didn't succeed")
		//os.Exit(1)
	}

	retry := func() {
		time.Sleep(retryDelay)
		SaveNodeDetails(retries + 1)
	}
	//get ipaddress
	ipAddress, err := helpers.GetIPAddress()
	if err != nil {
		log.Print("Error getting IP address:", err)
		return
		retry()
	}
	log.Print("🛰 IP Address:", ipAddress)
	//get ipfs id
	ipfsID, err := helpers.GetIpfsId()
	if err != nil {
		log.Print("Error getting IPFS ID:", err)
		retry()
		return
	}
	log.Print("🚀 IPFS ID:", ipfsID)
	//get cluster id
	ipfsClusterID, err := helpers.GetClusterID()
	if err != nil {
		log.Print("Error getting IPFS Cluster ID:", err)
		retry()
		return
	}
	log.Print("🚀 IPFS Cluster ID:", ipfsClusterID)
	
	nodeDetailsResponse, err := GetNodeDetails(ipAddress)
	if err != nil {
		log.Print("Error getting node details:", err)
		retry()
		return
	}
	// log.Print("Node Details Response:", nodeDetailsResponse)

	if nodeDetailsResponse.Data.IPFSClusterID != "" {
		log.Print("✨ Node details are already updated")
		// Return success message or handle as needed
		return
	}
	log.Print("nodeDetailsResponse.Data.IPFSClusterID:", nodeDetailsResponse.Data.IPFSClusterID)

	if ipfsClusterID == "" || ipfsID == "" {
		log.Print("IPFS ID or IPFS Cluster ID not found. Retrying...")
		retry()
		return
	}

	updateNodeDetails := UpdateNodeDetailsRequest{
		IPAddress:     ipAddress,
		IPFSClusterID: ipfsClusterID,
		IPFSID:        ipfsID,
	}

	err = UpdateNode(updateNodeDetails)
	if err != nil {
		log.Print("Error updating node details:", err)
		retry()
		return
	}
	log.Print("Node details updated successfully 😉")
}

// cleanVideoDirectory removes video files older than 6 hours from the specified directory
func CleanVideoDirectory(directory string) {
	for {
		threshold := time.Now().Add(-6 * time.Hour)

		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if the file is a regular file and older than the threshold
			if info.Mode().IsRegular() && info.ModTime().Before(threshold) {
				fmt.Printf("Deleting video file 📁: %s\n", path)
				if err := os.Remove(path); err != nil {
					fmt.Printf("Error deleting file 📁 %s: %v\n", path, err)
				}
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error cleaning video directory: %v\n", err)
		}
		interval := 30
		// Display a message in the terminal
		log.Print("🗑 junk deletion check completed. Waiting for the next check after " + fmt.Sprintf("%v", interval) + " seconds...")
		// Sleep for 30 seconds before the next junk deletion
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func GetNodeDetails(ipAddress string) (*NodeDetailsResponse, error) {
	// Implement logic to get node details using HTTP GET request
	resp, err := http.Get(fmt.Sprintf("%s/node/node-details/%s", hostURL, ipAddress))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var nodeDetailsResponse NodeDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&nodeDetailsResponse)
	if err != nil {
		return nil, err
	}

	return &nodeDetailsResponse, nil
}

func UpdateNode(updateNodeDetails UpdateNodeDetailsRequest) error {
	// Implement logic to update node details using HTTP POST request
	reqBody, err := json.Marshal(updateNodeDetails)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/node/update-node-details", hostURL), "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle the response if needed

	return nil
}