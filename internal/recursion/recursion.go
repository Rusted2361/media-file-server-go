package recursion

import (
    "fmt"
	"time"
	"os"
	"log"
	"media-file-server-go/internal/helpers"
)
////////////////////////////Recursive functions///////////////////////
		//////////////////////////////////////////////////////

// this will recursively check for clusterid and ipfs id
func HeartBeat() {
	for {
		// Check the local IPFS Cluster and IPFS node status
		clusterResponseLocal, _ := helpers.GetClusterID()
		ipfsResponseLocal, _ := helpers.GetIpfsId()

		// If either local IPFS Cluster or IPFS node is not running, exit the application
		if len(clusterResponseLocal) == 0 || len(ipfsResponseLocal) == 0 {
			fmt.Println("Ipfs Cluster or Ipfs is not running locally.")
			//exit
			os.Exit(1)

		}


		// Check the global (online) IPFS Cluster and IPFS node status
		clusterResponseOnline, _ := helpers.GetClusterID()
		ipfsResponseLocalOnline, _ := helpers.GetIpfsId()

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