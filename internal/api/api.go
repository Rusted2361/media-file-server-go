package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"media-file-server-go/internal/helpers"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

	//new routes for s3 streamable
	router.GET("/api/v2/file/view/access/:fileId/:accessKey", getAccessFileId)
}

///////////////////////Functions behind each API are defined here////////////////
//////////////////////////////////////////////////////

func getStatus(c *gin.Context) {
	//test ipaddress
	ipaddress, err := helpers.GetIPAddress()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cluster ID"})
		return
	}

	fmt.Printf("ipaddress: %s\n", ipaddress)

	//test ipfs id
	ipfsid, err := helpers.GetIpfsId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cluster ID"})
		return
	}

	fmt.Printf("ipfsid: %s\n", ipfsid)

	//test clusterid function
	clusterid, err := helpers.GetClusterID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cluster ID"})
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
		fmt.Println("accessData value is accessed", accessData)

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
	fmt.Println("ipfsMetaData sorted", ipfsMetaData)

	//path creation for checking video file locally
	path := fmt.Sprintf("videos/%s%s", accessData["accessKey"].(string), accessData["fileName"].(string))

	//bool to store filepath exist or not
	isFileExist := helpers.FileExists(path)

	//float to store original file size
	videoFileSize := accessData["fileSize"].(float64)

	fmt.Println("videoFileSize", videoFileSize)
	localfileSize := helpers.GetFileSize(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("localfileSize", localfileSize)

	var wg sync.WaitGroup

	if len(ipfsMetaData) < 3 {
		if !isFileExist {
			fmt.Println("I am in case 1 and file size is smaller than 3 cids")
			wg.Add(1) // Increment wait group by 2 for two goroutines

			// Goroutine for downloading chunks
			go func() {
				defer wg.Done()
				if err := helpers.DownloadAndWriteChunks(ipfsMetaData, accessData, path, c); err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}()

			wg.Wait()

		} else {
			if helpers.GetFileSize(path) == 0 {
				err := os.Remove(path)
				if err != nil {
					fmt.Println("Error:", err)
				}
			} else if helpers.GetFileSize(path) > 0 {
				fmt.Println("I am in case 2 file size is smaller than 3")
				helpers.StreamVideo(path, c)
			}
		}
	} else {
		if !isFileExist {
			fmt.Println("I am in case 1")
			wg.Add(2) // Increment wait group by 2 for two goroutines

			// Goroutine for downloading chunks
			go func() {
				defer wg.Done()
				if err := helpers.DownloadAndWriteChunks(ipfsMetaData, accessData, path, c); err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}()

			// Goroutine for checking playback condition
			go func() {
				defer wg.Done()
				for {
					// Check if enough data is downloaded to start playback
					if helpers.GetFileSize(path) >= (videoFileSize/float64(len(ipfsMetaData)))*3 {
						fmt.Println("Starting playback...")
						helpers.StreamVideo(path, c)
						fmt.Println("Playback started.")
						return // Exit after streaming the video
					}

					// Wait for some time before checking again
					time.Sleep(5 * time.Second)
				}
			}()
			wg.Wait()
		} else {
			if helpers.GetFileSize(path) < (videoFileSize/float64(len(ipfsMetaData)))*3 {
				err := os.Remove(path)
				if err != nil {
					fmt.Println("Error:", err)
				}
			} else if helpers.GetFileSize(path) >= (videoFileSize/float64(len(ipfsMetaData)))*3 {
				fmt.Println("I am in case 2")
				helpers.StreamVideo(path, c)
			}
		}
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
	fmt.Println("AccessDataResponse", AccessDataResponse)
	//This method uses map interfaces to deal with response data
	accessData, ok := AccessDataResponse["data"].(map[string]interface{})
	fmt.Println("accessData", accessData)
	if ok {
		fmt.Println("accessData value is accessed")
	}
	// Set Cache-Control header
	c.Writer.Header().Set("Cache-Control", "max-age=3600")

	// Set Last-Modified header to the current time
	c.Writer.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	// Extract modifiedSince header from the request
	modifiedSince := c.GetHeader("If-Modified-Since")

	if modifiedSince != "" {
		// Parse the modifiedSince header
		clientModifiedTime, err := time.Parse(http.TimeFormat, modifiedSince)
		if err != nil {
			fmt.Println("Error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse modifiedSince header"})
			return
		}

		// // Compare client's modified time with the current time
		// if time.Now().UTC().After(clientModifiedTime.Add(time.Hour * 1)) {
		// 	// If the resource has been modified, return the resource
		// 	c.JSON(http.StatusOK, gin.H{"message": "Resource modified"})
		// 	return
		// } else {
		// 	// If the resource has not been modified, return 304 Not Modified
		// 	c.Status(http.StatusNotModified)
		// 	return
		// }
		// Compare client's modified time with the current time
		if !(time.Now().UTC().After(clientModifiedTime.Add(time.Hour * 1))) {
			// If the resource has not been modified, return 304 Not Modified
			c.Status(http.StatusNotModified)
			return
		}
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
	fmt.Println("ipfsMetaData sorted", ipfsMetaData)
	fmt.Println("fileName", accessData["fileName"].(string))
	// Setting response headers for content type and filename
	// Set the Content-Type header based on the file type
	c.Writer.Header().Set("Content-Type", accessData["fileType"].(string))
	// Set Content-Disposition header to indicate inline display
	c.Writer.Header().Set("Content-Disposition", "inline; filename=\""+accessData["fileName"].(string)+"\"")
	//c.Writer.Header().Set("Content-Disposition", accessData["fileName"].(string))
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
			url := fmt.Sprintf("http://89.117.72.26:8080/api/v0/cat/%s", cid)

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
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt file data"})
				return
			}

			// Write the decrypted data to the pipe
			pw.Write(decryptedData)

		}
	}()
	// Pipe the reader to the response writer
	io.Copy(c.Writer, pr)
}

func getAccessFileId(c *gin.Context) {
	// Extract access key and token from URL parameters
	accessKey := c.Param("accessKey")
	token := c.Param("token")
	fileId := c.Param("fileId")
	fmt.Println("fileId", fileId)
	// Verify the access token
	AccessDataResponse, err := helpers.VerifyAccessTokenFileId(accessKey, token, fileId)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("AccessDataResponse", AccessDataResponse)

	//This method uses map interfaces to deal with response data
	accessData, ok := AccessDataResponse["data"].(map[string]interface{})
	fmt.Println("accessData", accessData)
	if ok {
		fmt.Println("accessData value is accessed")
	}
	// Set Cache-Control header
	c.Writer.Header().Set("Cache-Control", "max-age=3600")

	// Set Last-Modified header to the current time
	c.Writer.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	// Extract modifiedSince header from the request
	modifiedSince := c.GetHeader("If-Modified-Since")

	if modifiedSince != "" {
		// Parse the modifiedSince header
		clientModifiedTime, err := time.Parse(http.TimeFormat, modifiedSince)
		if err != nil {
			fmt.Println("Error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse modifiedSince header"})
			return
		}

		// // Compare client's modified time with the current time
		// if time.Now().UTC().After(clientModifiedTime.Add(time.Hour * 1)) {
		// 	// If the resource has been modified, return the resource
		// 	c.JSON(http.StatusOK, gin.H{"message": "Resource modified"})
		// 	return
		// } else {
		// 	// If the resource has not been modified, return 304 Not Modified
		// 	c.Status(http.StatusNotModified)
		// 	return
		// }
		// Compare client's modified time with the current time
		if !(time.Now().UTC().After(clientModifiedTime.Add(time.Hour * 1))) {
			// If the resource has not been modified, return 304 Not Modified
			c.Status(http.StatusNotModified)
			return
		}
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
	fmt.Println("ipfsMetaData sorted", ipfsMetaData)
	fmt.Println("fileName", accessData["fileName"].(string))
	// Setting response headers for content type and filename
	// Set the Content-Type header based on the file type
	c.Writer.Header().Set("Content-Type", accessData["fileType"].(string))
	// Set Content-Disposition header to indicate inline display
	c.Writer.Header().Set("Content-Disposition", "inline; filename=\""+accessData["fileName"].(string)+"\"")
	//c.Writer.Header().Set("Content-Disposition", accessData["fileName"].(string))
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
			url := fmt.Sprintf("http://89.117.72.26:8080/api/v0/cat/%s", cid)

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

			fmt.Println("accessData[data]", accessData["data"].(string))
			fmt.Println("accessData[secretKey]", accessData["secretKey"].(string))
			fmt.Println("accessData[accessKey]", accessData["accessKey"].(string))
			fmt.Println("accessData[iv]", accessData["iv"].(string))
			fmt.Println("accessData[salt]", accessData["salt"].(string))
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
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt file data"})
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
			url := fmt.Sprintf("http://89.117.72.26:8080/api/v0/cat/%s", cid)

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
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt file data"})
				return
			}

			// Write the decrypted data to the pipe
			pw.Write(decryptedData)

		}
	}()
	// Pipe the reader to the response writer
	io.Copy(c.Writer, pr)
}
