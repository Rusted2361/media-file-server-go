package api

import (
	"fmt"
	"sort"
	"net/http"
	"io/ioutil"
	"io"
    "github.com/gin-gonic/gin"
	"media-file-server-go/internal/helpers"
	//"media-file-server-go/internal/recursion"
)

// RegisterRoutes registers API routes
func RegisterRoutes(router *gin.Engine) {
    router.GET("/api/file/node/status", getStatus)
    router.GET("/api/file/view/access-play/:accessKey", playVideo)
    router.GET("/api/file/view/access-play/:accessKey/:token", playVideo)
	router.GET("/api/file/view/access/:accessKey", getAccessFile)
	router.GET("/api/file/view/access/:accessKey/:token", getAccessFile)
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
	//Access fileType property
	fileType, ok := accessData["fileType"].(string)
	if !ok {
		// Handle the case where "fileName" key is not present or has an unexpected type
		fmt.Println("Error: 'fileType' key not found or has an unexpected type")
		return
	} else {
		fmt.Println("fileType:", fileType)
	}
	// Concatenate strings to form the path
	path := "videos/" + RespAccessKey + fileName
	fmt.Println("path:", path)
	
	// Setting response headers for content type and filename
	c.Writer.Header().Set("Content-Type", fileType)
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`filename="%s"`, fileName))

	// Create a pipe
    pr, pw := io.Pipe()
	// Start a goroutine to produce data and write to the pipe
	go func() {
        defer pw.Close()
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
        } else{
			fmt.Println("cid:",cid)
		}
					// Making an HTTP GET request to fetch file data from IPFS
					url := fmt.Sprintf("http://46.101.133.110:8080/api/v0/cat/%s", cid)

					// Make an HTTP GET request
					respone, err := http.Get(url)
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
					fmt.Println("fileRespone:", fileRespone)
		//Access accessData->data property
		accessData_data, ok := accessData["data"].(string)
		if !ok {
			// Handle the case where "data" key is not present or has an unexpected type
			fmt.Println("Error: 'data' key not found or has an unexpected type")
			return
		}else {
			fmt.Println("accessData_data:", accessData_data)
		}
		//Access accessData->secretKey property
		accessData_secretKey, ok := accessData["secretKey"].(string)
		if !ok {
			// Handle the case where "secretKey" key is not present or has an unexpected type
			fmt.Println("Error: 'secretKey' key not found or has an unexpected type")
			return
		}else {
			fmt.Println("accessData_secretKey:", accessData_secretKey)
		}
		//Access accessData->accessKey property
		accessData_accessKey, ok := accessData["accessKey"].(string)
		if !ok {
			// Handle the case where "accessKey" key is not present or has an unexpected type
			fmt.Println("Error: 'accessKey' key not found or has an unexpected type")
			return
		}else {
			fmt.Println("accessData_accessKey:", accessData_accessKey)
		}
		//Access accessData->iv property
		accessData_iv, ok := accessData["iv"].(string)
		if !ok {
			// Handle the case where "iv" key is not present or has an unexpected type
			fmt.Println("Error: 'iv' key not found or has an unexpected type")
			return
		}else {
			fmt.Println("accessData_iv:", accessData_iv)
		}
		//Access accessData->salt property
		accessData_salt, ok := accessData["salt"].(string)
		if !ok {
			// Handle the case where "salt" key is not present or has an unexpected type
			fmt.Println("Error: 'salt' key not found or has an unexpected type")
			return
		}else {
			fmt.Println("accessData_salt:", accessData_salt)
		}

    }
	}()
	// Pipe the reader to the response writer
    io.Copy(c.Writer, pr)

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

    // Check if type assertions were successful
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
            } else{
                fmt.Println("cid:",cid)
            }
            // Making an HTTP GET request to fetch file data from IPFS
            url := fmt.Sprintf("http://46.101.133.110:8080/api/v0/cat/%s", cid)
            // Make an HTTP GET request
            respone, err := http.Get(url)
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
			iv := accessData["iv"].(string)
            
		
            // Decrypting data using a custom function
			decryptedData, err := helpers.DecryptedSecretKeyAndFile(
				accessData["data"].(string), 
				accessData["secretKey"].(string), 
				accessData["accessKey"].(string), 
				iv, 
				accessData["salt"].(string),
				[]byte(fileRespone),
			)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			
			// fmt.Println(string(decryptedData[:1000]))
			
			// Write the decrypted data to the pipe
			pw.Write(decryptedData)
			
        }
    }()
    // Pipe the reader to the response writer
    io.Copy(c.Writer, pr)
}

func downloadFile(c *gin.Context) {
	// Implement the logic for the downloadFile function
	// ...
	c.String(http.StatusOK, "File content for download")
}