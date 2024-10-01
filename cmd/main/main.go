package main

import (
	"log"
	"media-file-server-go/internal/api"
	"media-file-server-go/internal/recursion"
	"time"

	"github.com/gin-gonic/gin"
)

// ///////////////////routes bind with specific function behind it///////////////////
// ////////////////////////////////////////////////////
func main() {
	// Delay for 10 seconds before starting HeartBeat
	time.Sleep(6 * time.Second)

	// Start SaveNodeDetails immediately
	go recursion.SaveNodeDetails(0)

	// Register API routes
	router := gin.Default()
	api.RegisterRoutes(router)

	// Run HeartBeat in a goroutine with recovery
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("HeartBeat recovered from panic: %v", r)
					}
				}()

				// Start HeartBeat
				recursion.HeartBeat()
			}()

			// Sleep for a short interval before restarting HeartBeat
			time.Sleep(2 * time.Second)
		}
	}()

	// Start video deletion task
	go recursion.CleanVideoDirectory("videos")

	if err := router.Run(":3008"); err != nil {
		log.Println("Failed to start the server:", err)
	}
}
