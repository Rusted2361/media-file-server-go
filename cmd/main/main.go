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
	
	go recursion.SaveNodeOsDetails(0)
	// Introduce a 5-second delay before starting the HeartBeat function
	time.Sleep(10 * time.Second)
	go recursion.HeartBeat()

	router := gin.Default()
	// Register API routes
    api.RegisterRoutes(router)
	//comment this one once you uncomment below recursion
	if err := router.Run(":3009"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}
}