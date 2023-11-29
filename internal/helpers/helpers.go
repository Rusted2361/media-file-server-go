package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)
/////////////////////Helper Functions are defined here////////////////
		//////////////////////////////////////////////////////

// getIpfsId fetches ID from an IPFS node based on the given IP address.
func GetIpfsId(ipAddress ...string) (string, error) {
	// Construct the URL for the IPFS node's /api/v0/id endpoint
	var url string

	if len(ipAddress) > 0 {
		url = fmt.Sprintf("http://%s:9094/id", ipAddress[0])
	} else {
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

	// Convert the response body to a string and return it
	return string(body), nil
}

// Function to get ID from an IPFS cluster based on the environment
func GetClusterID(ipAddress ...string) ([]byte, error) {
	
	// if ipAddress == "" {
	// 	ipAddress = "localhost"
	// }
	// // Construct the URL
	// url := fmt.Sprintf("http://%s:9094/id", ipAddress)

	var url string

	if len(ipAddress) > 0 {
		url = fmt.Sprintf("http://%s:9094/id", ipAddress[0])
	} else {
		url = "http://localhost:9094/id"
	}


	// Make an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Print the status URL to the console
	fmt.Println("Status URL:", url)

	
	// Return the response body
	return body, nil
}

//function 
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