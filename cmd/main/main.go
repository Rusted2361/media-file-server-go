package main

import (
	"fmt"
	//"time"
	"github.com/gin-gonic/gin"
	"media-file-server-go/internal/api"
	//"media-file-server-go/internal/recursion"
	
)

/////////////////////routes bind with specific function behind it///////////////////
		//////////////////////////////////////////////////////
func main() {

	// Delay for 10 seconds before starting HeartBeat
	// time.Sleep(10 * time.Second)
	//go recursion.HeartBeat()

	// Start SaveNodeDetails immediately
	//go recursion.SaveNodeDetails(0)

	// Register API routes
	router := gin.Default()
	api.RegisterRoutes(router)
	if err := router.Run(":3009"); err != nil {
		fmt.Println("Failed to start the server:", err)
	}

}