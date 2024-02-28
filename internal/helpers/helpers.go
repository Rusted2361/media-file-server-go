package helpers

import (
	"bytes"
	"log"
	"encoding/json"
	"time"
	"fmt"
	"io/ioutil"
	"os"
	"io"
	"strconv"
	"strings"
	"net/http"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/pbkdf2"
	"github.com/gin-gonic/gin"
)
const hostURLstage = "https://staging-be.storagechain.io/api";
const hostURLdev = "https://storagechain-be.invo.zone/api";
const hostURLlive = "https://api.storagechain.io/api";
type IpfsID struct {
    Id string
}

type ClusterID struct {
    Id string
}


/////////////////////Helper Functions are defined here////////////////
		//////////////////////////////////////////////////////
//get ip address
func GetIPAddress() (string, error) {
	req, err := http.Get("https://httpbin.org/ip")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Extract the IP address
	ipAddress, ok := result["origin"].(string)
	if !ok {
		return "", fmt.Errorf("Unable to extract IP address from the response")
	}

	return ipAddress, nil
}

// getIpfsId fetches ID from an IPFS node based on the given IP address.
func GetIpfsId(ipAddress ...string) (string, error) {
	// Construct the URL for the IPFS node's /api/v0/id endpoint
	var url string
	payload := []byte("")
	if len(ipAddress) > 0 {
		url = fmt.Sprintf("http://%s:5001/api/v0/id", ipAddress)
		
	} else {
		url = "http://127.0.0.1:5001/api/v0/id"
		//url = "http://135.181.55.235:5001/api/v0/id"
	}

	// Make an HTTP GET request to the IPFS node
	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		// Return an empty string and the error if the request fails
		return "", err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// Return an empty string and the error if reading the body fails
		fmt.Println("Error getting IpfsID:", err)
		return "", err
	}
	log.Printf("response body of IpfsID: %s", string(body))

	var ipfsid IpfsID
    json.Unmarshal(body, &ipfsid)
	// Convert the response body to a string and return it
	return ipfsid.Id, nil
}

// Function to get ID from an IPFS cluster based on the environment
func GetClusterID(ipAddress ...string) (string, error) {

	var url string

	if len(ipAddress) > 0 {
		url = fmt.Sprintf("http://%s:9094/id", ipAddress)
	} else {
		//url = "http://135.181.55.235:9084/id"
		url = "http://127.0.0.1:9094/id"
	}

	// Make an HTTP GET request to the IPFS node
	response, err := http.Get(url)
	if err != nil {
		// Return an empty string and the error if the request fails
		return "", err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// Return an empty string and the error if reading the body fails
		fmt.Println("Error getting ClusterID:", err)
		return "", err
	}
	var clusterid ClusterID
	log.Printf("response body of http://127.0.0.1:9094/id: %s", string(body))
    json.Unmarshal(body, &clusterid)
	// Convert the response body to a string and return it
	return clusterid.Id, nil
}

// Function to verify access token and fetch data
func VerifyAccessToken(accessKey, token string) (map[string]interface{}, error) {
	// Define the request payload
	requestData := map[string]string{"accessKey": accessKey, "token": token}

	// Create an HTTP client with a timeout (adjust the timeout duration accordingly)
	client := &http.Client{Timeout: 5 * time.Second}

	// Encode the request payload directly into the request body
	reqBody := new(bytes.Buffer)
	if err := json.NewEncoder(reqBody).Encode(requestData); err != nil {
		return nil, err
	}

	// Send a POST request to verify the access token
	url := fmt.Sprintf("%s/file/access/verify-token", hostURLstage)
	resp, err := client.Post(
		url,
		"application/json",
		reqBody,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non-successful HTTP status code: %d", resp.StatusCode)
	}

	// Parse the JSON response
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, err
	}

	return responseData, nil
}

// Function to decrypt filedata using decrypted key iv and filedata
func decryptor(key, trimiv, data []byte) ([]byte, error) {
	// Create a new AES block cipher with the key
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// Create a GCM cipher mode
	aesgcm, err := cipher.NewGCMWithNonceSize(b, 32)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	decryptedData, err := aesgcm.Open(nil, trimiv, data, nil)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}

// Function to derive pbkf2 keyfrom secret key and salt
func deriveKey(secretKey string, userSalt string) ([]byte) {
	
	// Derive the key using PBKDF2 with provided salt and other parameters
	derivedKey := pbkdf2.Key([]byte(secretKey), []byte(userSalt), 1000, 32, sha256.New)
	// Return the derived key
	return derivedKey
}

// Function to decrypt key and then decrypt data using AES-GCM
func DecryptedSecretKeyAndFile(data, secretKey, accessKey, iv, userSalt string, fileData []byte) ([]byte, error) {
    
	//Nonce and data to decrypt Master Key
	//nonce/iv to decrypt key
	trimaccessKey := []byte(accessKey)
	//data to decrypt key
	hexdata, _ := hex.DecodeString(data)

	//nonce/iv to decrypt data
	trimiv := []byte(iv)
	//fileData contains the original data to be decrypted

	//gcm method
	key:= deriveKey(secretKey, userSalt)

	decryptedKey, err := decryptor(key, trimaccessKey, hexdata)
	if err != nil {
		return nil, err
	}
	fmt.Println("DecryptedKey accessed",decryptedKey)
	
	//Decrypt the Data
	decryptedData, err := decryptor(decryptedKey, trimiv, fileData)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	fmt.Println("Decrypted Data accessed",decryptedData[:100])
	//return Decrypted Data
	return decryptedData, nil
}

//Function to check filepath
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

//Function to get filesize
func GetFileSize(filePath string) (float64) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return float64(fileInfo.Size())
}

//Function to decrypt and download
func DownloadAndWriteChunks(ipfsMetaData []interface{}, accessData map[string]interface{}, path string, c *gin.Context) error  {
    file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return err
	}
	defer file.Close()
		
	for i := 0; i < len(ipfsMetaData); i++ {
		fmt.Println("len(ipfsMetaData)",len(ipfsMetaData))
		metaData := ipfsMetaData[i].(map[string]interface{})
		cid, ok := metaData["cid"].(string)
		if !ok {
			fmt.Println("Error: 'cid' key not found or has an unexpected type")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid CID format"})
			return err
		}
		url := fmt.Sprintf("http://89.117.72.26:8080/api/v0/cat/%s", cid)
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Error:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file data from IPFS"})
			return err
		}
		defer response.Body.Close()
		fileResponse, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file response"})
			return err
		}
		decryptedData, err := DecryptedSecretKeyAndFile(
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
			return err
		}
		file.Write(decryptedData)
	}
	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file information:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file information"})
		return err
	}
	fileSize := stat.Size()
	rangeHeader := c.Request.Header.Get("Range")
	if rangeHeader != "" {
		HandleRangeRequest(c, path, fileSize)
		return err
	}
	HandleFullRequest(c, path, fileSize)

	return nil
}

//Function to stream local video
func StreamVideo(path string, c *gin.Context) {
    fmt.Println("~ File exists locally. Streaming...")
    file, err := os.Open(path)
    if err != nil {
        c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer file.Close()
    stat, err := file.Stat()
    if err != nil {
        c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    fileSize := stat.Size()
    rangeHeader := c.Request.Header.Get("Range")
    if rangeHeader != "" {
        HandleRangeRequest(c, path, fileSize)
        return
    }
    HandleFullRequest(c, path, fileSize)
}

//Function to handle partial content
func HandleRangeRequest(c *gin.Context, path string, fileSize int64) {
	parts := c.Request.Header.Get("Range")[6:] // Remove "bytes=" prefix
	rangeValues := strings.Split(parts, "-")

	start, err := strconv.ParseInt(rangeValues[0], 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid start position in Range header"})
		return
	}

	var end int64
	if rangeValues[1] != "" {
		end, err = strconv.ParseInt(rangeValues[1], 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid end position in Range header"})
			return
		}
	} else {
		end = fileSize - 1
	}

	chunksize := end - start + 1

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Writer.Header().Set("Accept-Ranges", "bytes")
	c.Writer.Header().Set("Content-Length", strconv.FormatInt(chunksize, 10))
	c.Writer.Header().Set("Content-Type", "video/mp4")

	c.Writer.WriteHeader(http.StatusPartialContent)

	file.Seek(start, io.SeekStart)

	io.CopyN(c.Writer, file, chunksize)
}

//Function to handle full content
func HandleFullRequest(c *gin.Context, path string, fileSize int64) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	c.Writer.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Writer.Header().Set("Content-Type", "video/mp4")

	c.Writer.WriteHeader(http.StatusOK)

	io.Copy(c.Writer, file)
}