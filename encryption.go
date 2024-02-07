package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

// Function to derive PBKDF2 key from secret key and salt
func deriveKey(secretKey, userSalt string) []byte {
	// Derive the key using PBKDF2 with provided salt and other parameters
	derivedKey := pbkdf2.Key([]byte(secretKey), []byte(userSalt), 1000, 32, sha256.New)
	// Return the derived key
	return derivedKey
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
func EncryptedSecretKeyAndFile(data, secretKey, accessKey, iv, userSalt string, fileData []byte) ([]byte, error) {
	// Nonce/IV to encrypt the key
	trimAccessKey := []byte(accessKey)

	// Nonce/IV to encrypt the data
	trimIV := []byte(iv)

	// Derive the key using PBKDF2
	key := deriveKey(secretKey, userSalt)

	// Encrypt the key
	encryptedKey, err := encryptor(key, trimAccessKey, []byte(data))
	if err != nil {
		return nil, err
	}

	// Encrypt the data
	encryptedData, err := encryptor(key, trimIV, fileData)
	if err != nil {
		return nil, err
	}

	// Concatenate encrypted key and data
	encryptedResult := append(encryptedKey, encryptedData...)

	return encryptedResult, nil
}

func main() {
	// Sample parameters
	data := "data_to_encrypt"
	secretKey := "secret_key"
	accessKey := "access_key"
	iv := "initialization_vector"
	userSalt := "user_salt"
	fileData := []byte("file_data_to_encrypt")

	// Encrypt key and data
	encryptedResult, err := EncryptedSecretKeyAndFile(data, secretKey, accessKey, iv, userSalt, fileData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the encrypted result
	fmt.Println("Encrypted Result:", encryptedResult)
}
