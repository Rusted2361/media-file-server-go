package main

import (
	"fmt"
	"time"
	"github.com/gin-gonic/gin"
	"media-file-server-go/internal/api"
	 "media-file-server-go/internal/recursion"
	// "media-file-server-go/internal/helpers"
	
)

/////////////////////routes bind with specific function behind it///////////////////
		//////////////////////////////////////////////////////
func main() {

	//////uncomment this function to immediately shut service if cluster id not found
	// Create a new Goroutine for the heartBeat function
	
	// Start SaveNodeOsDetails immediately
	go recursion.SaveNodeOsDetails(0)

	router := gin.Default()
	// Register API routes
	api.RegisterRoutes(router)

	// Delay for 10 seconds before starting HeartBeat
	time.Sleep(10 * time.Second)
	go recursion.HeartBeat()
	
	//comment this one once you uncomment below recursion
	if err := router.Run(":3009"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}

}