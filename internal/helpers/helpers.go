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
	"crypto/rand"
	"hash"
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

// Function to generate a secret key for encryption using PBKDF2
func generateSecretKeyForEncryption(secretKeyString string, userSalt string) ([]byte, error) {
	// Derive the key using PBKDF2 with provided salt and other parameters
	derivedKey := pbkdf2.Key([]byte(secretKeyString), []byte(userSalt), 1000, 32, sha256.New)

	// Return the derived key
	return derivedKey, nil
}

// Function to convert a hex string to a byte slice
func fromHexString(hexString string) ([]byte, error) {
	return hex.DecodeString(hexString)
}
func deriveKey(passphrase string, salt []byte) ([]byte, []byte) {
	if salt == nil {
		salt = make([]byte, 8)
		// http://www.ietf.org/rfc/rfc2898.txt
		// Salt.
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(passphrase), salt, 1000, 32, sha256.New), salt
}
func deriveKeyInternal(passphrase string, salt []byte, h func() hash.Hash) ([]byte, []byte) {
	return pbkdf2.Key([]byte(passphrase), salt, 1000, 32, h), salt
}
func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	padLen := int(data[len(data)-1])
	if padLen > len(data) {
		// Invalid padding
		return nil
	}
	return data[:len(data)-padLen]
}
// Function to decrypt data using AES-GCM
func DecryptedSecretKeyAndFile(data, secretKey, accessKey, iv, fileData, userSalt string) (string, error) {
    hexsalt, _ := hex.DecodeString(userSalt)
	hexdata, _ := hex.DecodeString(data)
	hexaccessKey, _ := hex.DecodeString(accessKey)
	
	fmt.Println("hexdata:",hexdata)
	fmt.Println("accessKey:",accessKey)
	trimaccessKey := hexaccessKey[:32]
	fmt.Println("trimaccessKey:",trimaccessKey)

//gcm method
	key, _ := deriveKey(secretKey, []byte(hexsalt))
	b, _ := aes.NewCipher(key)
	
	// Import 16 bytes nonce
    aesgcm, err := cipher.NewGCMWithNonceSize(b, 32)
    if err != nil {
        panic(err.Error())
    }
	//aesgcm, _ := cipher.NewGCM(b)
	// Decrypt the data
	decryptedData, err := aesgcm.Open(nil, trimaccessKey, hexdata, nil)
	if err != nil {
		return "", err
	}

	// Convert the decrypted data to string
	decryptedString := string(decryptedData)

	return decryptedString, nil
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