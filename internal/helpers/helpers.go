package helpers

import (
	"bytes"
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
		url = "http://localhost:5001/api/v0/id"
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
		return "", err
	}
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
		url = "http://localhost:9094/id"
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
		return "", err
	}
	var clusterid ClusterID
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
	resp, err := client.Post(
		"https://storagechain-be.invo.zone/api/file/access/verify-token",
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
func decryptFile(decryptedKey, trimiv, fileData []byte) ([]byte, error) {
	
	// Create a new AES block cipher with the key
	b, err := aes.NewCipher(decryptedKey)
	if err != nil {
		return nil, err
	}
	
	// Create a GCM cipher mode
	aesgcm, err := cipher.NewGCMWithNonceSize(b, 32)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	decryptedData, err := aesgcm.Open(nil, trimiv, fileData, nil)
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
    fmt.Println("Data:",data)
	fmt.Println("secretKey:",secretKey)
	fmt.Println("accessKey:",accessKey)
	fmt.Println("iv:",iv)
	fmt.Println("userSalt:",userSalt)
	fmt.Println("fileData:",fileData[:100])
	//Nonce and data to decrypt Master Key
	//nonce/iv to decrypt key
	hexaccessKey, _ := hex.DecodeString(accessKey)
	trimaccessKey := hexaccessKey[:32]

	//data to decrypt key
	hexdata, _ := hex.DecodeString(data)
	
	//Nonce and data to decrypt original data
	//nonce/iv to decrypt data
	hexiv, _ :=hex.DecodeString(iv)
	trimiv := hexiv[:32]

	//fileData contains the original data to be decrypted

	//gcm method
	key:= deriveKey(secretKey, userSalt)

	//cipher block generation from derived key
	b, _ := aes.NewCipher(key)

	//gcm generation from 32 bytes nonstandard nonce
    aesgcm, err := cipher.NewGCMWithNonceSize(b, 32)
    if err != nil {
        panic(err.Error())
    }

	// Decrypt the key
	decryptedKey, err := aesgcm.Open(nil, trimaccessKey, hexdata, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("DecryptedKey accessed",decryptedKey)

	//Decrypt the Data
	decryptedData, err := decryptFile(decryptedKey, trimiv, fileData)
	if err != nil {
		fmt.Println("Error2:", err)
		return nil, err
	}
	fmt.Println("Decrypted Data accessed")

	//return Decrypted Data
	return decryptedData, nil
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