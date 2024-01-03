package api

import (
	"fmt"
	"sort"
	//"sync"
	"net/http"
	"io/ioutil"
	"io"
	"os"
    "github.com/gin-gonic/gin"
	"media-file-server-go/internal/helpers"
)

// RegisterRoutes registers API routes
func RegisterRoutes(router *gin.Engine) {
    router.GET("/api/file/node/status", getStatus)
    router.GET("/api/file/view/access-play/:accessKey", playVideo)
    router.GET("/api/file/view/access-play/:accessKey/:token", playVideo)
	router.GET("/api/file/view/access/:accessKey", getAccessFile)
	router.GET("/api/file/view/access/:accessKey/:token", getAccessFile)
	router.GET("/api/file/download/:accessKey/", downloadFile)
	router.GET("/api/file/download/:accessKey/:token", downloadFile)
}

///////////////////////Functions behind each API are defined here////////////////
		//////////////////////////////////////////////////////
	
func getStatus(c *gin.Context) {
	//test ipaddress
	ipaddress, err := helpers.GetIPAddress()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ipaddress: %s\n", ipaddress)

	//test ipfs id
	ipfsid, err := helpers.GetIpfsId()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ipfsid: %s\n", ipfsid)

	//test clusterid function
	clusterid, err := helpers.GetClusterID()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("clusterid: %s\n", clusterid)
	
	//get clusterID function
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
    // This method uses map interfaces to deal with response data
    accessData, ok := AccessDataResponse["data"].(map[string]interface{})
    if ok {
        fmt.Println("accessData value is accessed")
    }
    fileMetaDataValue, ok := accessData["fileMetaData"].([]interface{})
    if ok {
        fmt.Println("fileMetaDataValue is a valid array")
    }
    // Custom sorting function
    sort.Slice(fileMetaDataValue, func(i, j int) bool {
        indexI, okI := fileMetaDataValue[i].(map[string]interface{})["index"].(float64)
        indexJ, okJ := fileMetaDataValue[j].(map[string]interface{})["index"].(float64)
        // Check if type assertions are successful
        if okI && okJ {
            return int(indexI) < int(indexJ)
        }
        // Handle the case where type assertions failed
        return false
    })
    // Storing sorted data in ipfsMetaData
    ipfsMetaData := fileMetaDataValue
    fmt.Println("ipfsMetaData sorted")
    path := fmt.Sprintf("videos/%s%s", accessData["accessKey"].(string), accessData["fileName"].(string))
    if _, err := os.Stat(path); os.IsNotExist(err) {
        file, err := os.Create(path)
        if err != nil {
            fmt.Println("Error creating file:", err)
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
            return
        }
        defer file.Close()
        for i := 0; i < len(ipfsMetaData); i++ {
            metaData := ipfsMetaData[i].(map[string]interface{})
            cid, ok := metaData["cid"].(string)
            if !ok {
                fmt.Println("Error: 'cid' key not found or has an unexpected type")
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid CID format"})
                return
            }
            url := fmt.Sprintf("http://46.101.133.110:8080/api/v0/cat/%s", cid)
            response, err := http.Get(url)
            if err != nil {
                fmt.Println("Error:", err)
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file data from IPFS"})
                return
            }
            defer response.Body.Close()
            fileResponse, err := io.ReadAll(response.Body)
            if err != nil {
                fmt.Println("Error:", err)
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file response"})
                return
            }
            decryptedData, err := helpers.DecryptedSecretKeyAndFile(
                accessData["data"].(string),
                accessData["secretKey"].(string),
                accessData["accessKey"].(string),
                accessData["iv"].(string),
                accessData["salt"].(string),
                fileResponse,
            )
            if err != nil {
                fmt.Println("Error:", err)
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt file data"})
                return
            }
            file.Write(decryptedData)
        }
        stat, err := file.Stat()
        if err != nil {
            fmt.Println("Error getting file information:", err)
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file information"})
            return
        }
        fileSize := stat.Size()
        rangeHeader := c.Request.Header.Get("Range")
        if rangeHeader != "" {
            helpers.HandleRangeRequest(c, path, fileSize)
            return
        }
        helpers.HandleFullRequest(c, path, fileSize)
    } else {
        // File exists locally. Stream video...
        fmt.Println(":rocket: ~ File exists locally. Streaming...")
        file, err := os.Open(path)
        if err != nil {
            fmt.Println("Error opening file:", err)
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
            return
        }
        defer file.Close()
        stat, err := file.Stat()
        if err != nil {
            fmt.Println("Error getting file information:", err)
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file information"})
            return
        }
        fileSize := stat.Size()
        rangeHeader := c.Request.Header.Get("Range")
        if rangeHeader != "" {
            helpers.HandleRangeRequest(c, path, fileSize)
            return
        }
        helpers.HandleFullRequest(c, path, fileSize)
    }
}

func getAccessFile(c *gin.Context) {
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
     }
    fileMetaDataValue, ok := accessData["fileMetaData"].([]interface{})
    if ok {
     fmt.Println("fileMetaDataValue is a valid array")
    }

    // Custom sorting function
    sort.Slice(fileMetaDataValue, func(i, j int) bool {
     indexI, okI := fileMetaDataValue[i].(map[string]interface{})["index"].(float64)
     indexJ, okJ := fileMetaDataValue[j].(map[string]interface{})["index"].(float64)

    // Check if type assertions are successful
    if okI && okJ {
     	return int(indexI) < int(indexJ)
    }
    // Handle the case where type assertions failed
    	return false
    })

    // Storing sorted data in ipfsMetaData
    ipfsMetaData := fileMetaDataValue
    fmt.Println("ipfsMetaData sorted")	

	// Setting response headers for content type and filename
	c.Writer.Header().Set("Content-Type", accessData["fileType"].(string))
	c.Writer.Header().Set("Content-Disposition", accessData["fileName"].(string))
	//c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`filename="%s"`, accessData["fileName"].(string)))

	// Create a pipe
    pr, pw := io.Pipe()

    // Start a goroutine to produce data and write to the pipe
    go func() {
        defer pw.Close()

		// Create an HTTP client with a timeout of 5 seconds
		client := &http.Client{}

        // Looping through ipfsMetaData and fetching file data
        for i := 0; i < len(ipfsMetaData); i++ {
            // Type-assert ipfsMetaData[i] to a map[string]interface{}
            metaData, ok := ipfsMetaData[i].(map[string]interface{})
            if !ok {
                // Handle the case where type assertion fails
                fmt.Println("Error: ipfsMetaData is not a valid map")
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid metadata format"})
                return
            }

            // Fetch the "cid" value from the map
            cid, ok := metaData["cid"].(string)
            if !ok {
                // Handle the case where "cid" key is not present or has an unexpected type
                fmt.Println("Error: 'cid' key not found or has an unexpected type")
                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid CID format"})
                return
            }

            // Making an HTTP GET request to fetch file data from IPFS
            url := fmt.Sprintf("http://46.101.133.110:8080/api/v0/cat/%s", cid)
			
            // Make an HTTP GET request
            respone, err := client.Get(url)
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
            defer respone.Body.Close()

            // Read the response body
            fileRespone, err := ioutil.ReadAll(respone.Body)
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
			
            // Decrypting data using a custom function
			decryptedData, err := helpers.DecryptedSecretKeyAndFile(
				accessData["data"].(string), 
				accessData["secretKey"].(string), 
				accessData["accessKey"].(string), 
				accessData["iv"].(string), 
				accessData["salt"].(string),
				[]byte(fileRespone),
			)
			if err != nil {
				fmt.Println("Error1:", err)
				return
			}
			
			// Write the decrypted data to the pipe
			pw.Write(decryptedData)
			
        }
    }()
    // Pipe the reader to the response writer
    io.Copy(c.Writer, pr)
}

func downloadFile(c *gin.Context) {
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
		 }
		fileMetaDataValue, ok := accessData["fileMetaData"].([]interface{})
		if ok {
		 fmt.Println("fileMetaDataValue is a valid array")
		}
	
		// Custom sorting function
		sort.Slice(fileMetaDataValue, func(i, j int) bool {
		 indexI, okI := fileMetaDataValue[i].(map[string]interface{})["index"].(float64)
		 indexJ, okJ := fileMetaDataValue[j].(map[string]interface{})["index"].(float64)
	
		// Check if type assertions are successful
		if okI && okJ {
			 return int(indexI) < int(indexJ)
		}
		// Handle the case where type assertions failed
			return false
		})
	
		// Storing sorted data in ipfsMetaData
		ipfsMetaData := fileMetaDataValue
		fmt.Println("ipfsMetaData sorted")
	
		// Setting response headers for content type and filename
		c.Writer.Header().Set("Content-Type", accessData["fileType"].(string))
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, accessData["fileName"].(string)))
	
		// Create a pipe
		pr, pw := io.Pipe()
	
		// Start a goroutine to produce data and write to the pipe
		go func() {
			defer pw.Close()
	
			// Create an HTTP client with a timeout of 5 seconds
			client := &http.Client{}
	
			// Looping through ipfsMetaData and fetching file data
			for i := 0; i < len(ipfsMetaData); i++ {
				// Type-assert ipfsMetaData[i] to a map[string]interface{}
				metaData, ok := ipfsMetaData[i].(map[string]interface{})
				if !ok {
					// Handle the case where type assertion fails
					fmt.Println("Error: ipfsMetaData is not a valid map")
					return
				}
	
				// Fetch the "cid" value from the map
				cid, ok := metaData["cid"].(string)
				if !ok {
					// Handle the case where "cid" key is not present or has an unexpected type
					fmt.Println("Error: 'cid' key not found or has an unexpected type")
					return
				}
	
				// Making an HTTP GET request to fetch file data from IPFS
				url := fmt.Sprintf("http://46.101.133.110:8080/api/v0/cat/%s", cid)
				
				// Make an HTTP GET request
				respone, err := client.Get(url)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				defer respone.Body.Close()
	
				// Read the response body
				fileRespone, err := ioutil.ReadAll(respone.Body)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				
				// Decrypting data using a custom function
				decryptedData, err := helpers.DecryptedSecretKeyAndFile(
					accessData["data"].(string), 
					accessData["secretKey"].(string), 
					accessData["accessKey"].(string), 
					accessData["iv"].(string), 
					accessData["salt"].(string),
					[]byte(fileRespone),
				)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				
				// Write the decrypted data to the pipe
				pw.Write(decryptedData)
				
			}
		}()
		// Pipe the reader to the response writer
		io.Copy(c.Writer, pr)
}