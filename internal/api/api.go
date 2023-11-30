package api

import (
	"fmt"
	"sort"
	"net/http"
    "github.com/gin-gonic/gin"
	"media-file-server-go/internal/helpers"
)

// RegisterRoutes registers API routes
func RegisterRoutes(router *gin.Engine) {
    router.GET("/api/file/node/status", getStatus)
    router.GET("/api/file/view/access-play/:accessKey", playVideo)
    router.GET("/api/file/view/access-play/:accessKey/:token", playVideo)
	router.GET("/api/file/view/access/:accessKey/:token", getAccessFile)
	router.GET("/api/file/download/:accessKey/:token", downloadFile)
}

///////////////////////Functions behind each API are defined here////////////////
		//////////////////////////////////////////////////////
	
func getStatus(c *gin.Context) {
	clusterID, err := helpers.GetClusterID()
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
	AccessDataResponse, err := helpers.VerifyAccessToken(accessKey, token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//This method uses map interfaces to deal with response data
    accessData, ok := AccessDataResponse["data"].(map[string]interface{})
    if ok {
         fmt.Println("accessData value is accessed")
     } else {
         fmt.Println("accessData is not a valid map")
     }
    fileMetaDataValue, ok := accessData["fileMetaData"].([]interface{})
    if ok {
     fmt.Println("fileMetaDataValue is a valid array")
    } else {
     fmt.Println("Data is not a string")
    }
    // Custom sorting function
    sort.Slice(fileMetaDataValue, func(i, j int) bool {
     indexI, okI := fileMetaDataValue[i].(map[string]interface{})["index"].(float64)
     indexJ, okJ := fileMetaDataValue[j].(map[string]interface{})["index"].(float64)
     // Check if type assertions were successful
     if okI && okJ {
         return int(indexI) < int(indexJ)
     }
     // Handle the case where type assertions failed
     return false
    })
    // Storing sorted data in ipfsMetaData
    ipfsMetaData := fileMetaDataValue
    // Print the sorted ipfsMetaData
    fmt.Println("Sorted ipfsMetaData:", ipfsMetaData)
	// Access accessKey property
	RespAccessKey, ok := accessData["accessKey"].(string)
	if !ok {
		// Handle the case where "accessKey" key is not present or has an unexpected type
		fmt.Println("Error: 'accessKey' key not found or has an unexpected type")
		return
	} else {
		fmt.Println("accessKey:", RespAccessKey)
	}

	//Access fileName property
	fileName, ok := accessData["fileName"].(string)
	if !ok {
		// Handle the case where "fileName" key is not present or has an unexpected type
		fmt.Println("Error: 'fileName' key not found or has an unexpected type")
		return
	} else {
		fmt.Println("fileName:", fileName)
	}

	// Concatenate strings to form the path
	path := "videos/" + RespAccessKey + fileName
	fmt.Println("path:", path)

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