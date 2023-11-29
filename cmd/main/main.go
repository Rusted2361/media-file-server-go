package main

import (
	"fmt"
	// "time"
	"github.com/gin-gonic/gin"
	"media-file-server-go/internal/api"
	// "media-file-server-go/internal/recursion"
	
)

/////////////////////routes bind with specific function behind it///////////////////
		//////////////////////////////////////////////////////
func main() {

	//////uncomment this function to immediately shut service if cluster id not found
	// Create a new Goroutine for the heartBeat function
	//go heartBeat()

	router := gin.Default()

	// Register API routes
    api.RegisterRoutes(router)
	//comment this one once you uncomment below recursion
	if err := router.Run(":3008"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}

	
	///////////uncomment this function to enable heartbeat function with 5 second delay will not immediately shut down/////////////
	//Run the Gin server on port 3008 in a Goroutine
	// go func() {
	// 	if err := router.Run(":3008"); err != nil {
	// 		fmt.Println("Failed to start the server:", err)
	// 	}
	// }()

	// // Wait for 5 seconds
	// <-time.After(5 * time.Second)

	// // Start the heartBeat function after 5 seconds
	// go recursion.HeartBeat()

	// // Block the main Goroutine so that the program doesn't exit
	// select {}
	
}


