package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

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

// Function to derive pbkf2 key from secret key and salt
func deriveKey(secretKey string, userSalt string) []byte {
	// Derive the key using PBKDF2 with provided salt and other parameters
	derivedKey := pbkdf2.Key([]byte(secretKey), []byte(userSalt), 1000, 32, sha256.New)
	// Return the derived key
	return derivedKey
}

// Function to decrypt key and then decrypt data using AES-GCM
func DecryptedSecretKeyAndFile(data, secretKey, accessKey, iv, userSalt string, fileData []byte) ([]byte, error) {

	// Nonce and data to decrypt Master Key
	// nonce/iv to decrypt key
	trimaccessKey := []byte(accessKey)
	// data to decrypt key
	hexdata, _ := hex.DecodeString(data)

	// nonce/iv to decrypt data
	trimiv := []byte(iv)
	// fileData contains the original data to be decrypted

	// gcm method
	key := deriveKey(secretKey, userSalt)

	decryptedKey, err := decryptor(key, trimaccessKey, hexdata)
	if err != nil {
		return nil, err
	}
	fmt.Println("DecryptedKey accessed", decryptedKey)
	

	// Decrypt the Data
	decryptedData, err := decryptor(decryptedKey, trimiv, fileData)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	fmt.Println("Decrypted Data accessed", decryptedData[:100])
	// return Decrypted Data
	return decryptedData, nil
}

// Function to encrypt data using key, iv, and data
func encryptor(key, trimIV, data []byte) ([]byte, error) {
	// Create a new AES block cipher with the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a GCM cipher mode
	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(trimIV))
	if err != nil {
		return nil, err
	}

	// Encrypt the data
	encryptedData := aesgcm.Seal(nil, trimIV, data, nil)
	
	return encryptedData, nil
}

// Function to encrypt key and data using AES-GCM
func EncryptedSecretKeyAndFileForEncryption(data, secretKey, accessKey, iv, userSalt string, fileData []byte) ([]byte, error) {
	// Nonce/IV to encrypt the key
	trimAccessKey := []byte(accessKey)

	// Nonce/IV to encrypt the data
	trimIV := []byte(iv)

	// Derive the key using PBKDF2
	key := deriveKey(secretKey, userSalt)
	fmt.Println("key:", key)
	//data to decrypt key
	hexdata, _ := hex.DecodeString(data)
	
	// Encrypt the data
	encryptedData, err := encryptor(key, trimIV, fileData)
	if err != nil {
		return nil, err
	}
	// Encrypt the key
	encryptedKey, err := encryptor(key, trimAccessKey, hexdata)
	if err != nil {
		return nil, err
	}
	
	
	
	// Concatenate encrypted key and data
	encryptedResult := append(encryptedKey, encryptedData...)

	return encryptedResult, nil
}

func main() {
	// Sample parameters
	data := "b48b0ec2c2eb77b8e6a57014e70f622193c19b581233a10086aeccd1a7b1d92925ea87e6e02e7e06f898ba979db86465"
	secretKey := "gsMgFSTZ1a5I09m0YBfImxE9hDPtIjws"
	accessKey := "RQ2pfPIQMuxVqCFa2vCVHbZjsQcZANaQ"
	iv := "KVpMsHBPq9qKID0XsZbz7NJe8KZ00dut"
	userSalt := "1f84383d473a70ae1bb9e90341718c7d05dfe338ffe7e699bea1e07db99cd573"
	fileData := []byte("file_data_to_encrypt")

	// Encrypt key and data
	encryptedResult, err := EncryptedSecretKeyAndFileForEncryption(data, secretKey, accessKey, iv, userSalt, fileData)
	if err != nil {
		fmt.Println("Error during encryption:", err)
		return
	}

	// Print the encrypted result
	fmt.Println("Encrypted Result:", encryptedResult)

	// Decrypt the encrypted data
	decryptedResult, err := DecryptedSecretKeyAndFile(data, secretKey, accessKey, iv, userSalt, encryptedResult[32:])
	if err != nil {
		fmt.Println("Error during decryption:", err)
		return
	}

	// Print the decrypted result
	fmt.Println("Decrypted Result:", string(decryptedResult))
}
