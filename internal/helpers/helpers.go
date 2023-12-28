package helpers

import (
	"bytes"
	"encoding/json"
	//"net"
	"fmt"
	"os"
	"io"
	//"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	//"crypto/rand"
	//"hash"
	//"encoding/base64"
	"golang.org/x/crypto/pbkdf2"
	"github.com/gin-gonic/gin"
	//"golang.org/x/crypto/chacha20poly1305"
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
			requestBody, err := json.Marshal(requestData)
			if err != nil {
				return nil, err
			}
		
			// Send a request to verify the access token
			resp, err := http.Post(
				"https://storagechain-be.invo.zone/api/file/access/verify-token",
				"application/json",
				bytes.NewBuffer(requestBody),
			)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		
			// Read the response body
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
		
			// Parse the JSON response
			var responseData map[string]interface{}
			err = json.Unmarshal(responseBody, &responseData)
			if err != nil {
				return nil, err
			}
		
			return responseData, nil
}


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
	
	fmt.Println("Decrypted Data accessed")
	
	return decryptedData, nil
}

func deriveKey(secretKey string, userSalt string) ([]byte) {
	
	// Derive the key using PBKDF2 with provided salt and other parameters
	derivedKey := pbkdf2.Key([]byte(secretKey), []byte(userSalt), 1000, 32, sha256.New)

	// Return the derived key
	return derivedKey
}

// Function to decrypt key and then decrypt data using AES-GCM
func DecryptedSecretKeyAndFile(data, secretKey, accessKey, iv, userSalt string, fileData []byte) ([]byte, error) {
    
	//nonce/iv to decrypt key
	hexaccessKey, _ := hex.DecodeString(accessKey)
	trimaccessKey := hexaccessKey[:32]
	//data to decrypt key
	hexdata, _ := hex.DecodeString(data)
	
	//nonce/iv to decrypt data
	hexiv, _ :=hex.DecodeString(iv)
	trimiv := hexiv[:32]
	//fileData contains the original data to be decrypted

	//gcm method
	key:= deriveKey(secretKey, userSalt)
	b, _ := aes.NewCipher(key)
	
	//Import 32 bytes nonstandard nonce
    aesgcm, err := cipher.NewGCMWithNonceSize(b, 32)
    if err != nil {
        panic(err.Error())
    }

	// Decrypt the key
	decryptedKey, err := aesgcm.Open(nil, trimaccessKey, hexdata, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("DecryptedKey accessed")

	//Decrypt the Data
	decryptedData, err := decryptFile(decryptedKey, trimiv, fileData)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	
	//return string(decryptedData), nil
	return decryptedData, nil
}

func HandleByteRange(c *gin.Context, path string, fileSize int64) {
	rangeHeader := c.GetHeader("Range")
	parts := strings.Split(strings.ReplaceAll(rangeHeader, "bytes=", ""), "-")
	start, _ := strconv.ParseInt(parts[0], 10, 64)
	end, _ := strconv.ParseInt(parts[1], 10, 64)
	chunkSize := end - start + 1

	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer file.Close()

	c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Writer.Header().Set("Accept-Ranges", "bytes")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", chunkSize))
	c.Writer.Header().Set("Content-Type", "video/mp4")
	c.Writer.WriteHeader(http.StatusPartialContent)

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to seek file",
		})
		return
	}

	io.CopyN(c.Writer, file, chunkSize)
}

func HandleFullContent(c *gin.Context, path string, fileSize int64) {
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Writer.Header().Set("Content-Type", "video/mp4")
	c.Writer.WriteHeader(http.StatusOK)

	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer file.Close()

	io.Copy(c.Writer, file)
}

func HandleExistingFile(c *gin.Context, path string) {
	// Functionality for streaming and response of an existing file
}