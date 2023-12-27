package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
)

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Generate a random nonce
	nonce := make([]byte, chacha20poly1305.NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Create a new ChaCha20-Poly1305 cipher
	c, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Encrypt the plaintext
	ciphertext := c.Seal(nil, nonce, plaintext, nil)

	// Append the nonce to the ciphertext
	ciphertext = append(ciphertext, nonce...)

	return ciphertext, nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Extract the nonce from the end of the ciphertext
	nonce := ciphertext[len(ciphertext)-chacha20poly1305.NonceSize:]
	ciphertext = ciphertext[:len(ciphertext)-chacha20poly1305.NonceSize]

	// Create a new ChaCha20-Poly1305 cipher
	c, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Decrypt the ciphertext
	plaintext, err := c.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("nonce:", nonce)
	return plaintext, nil
}

func main() {
	// Example key (32 bytes)
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		fmt.Println("Error generating key:", err)
		os.Exit(1)
	}

	// Example plaintext
	plaintext := []byte("Hello, ChaCha20-Poly1305!")

	// Encrypt
	ciphertext, err := encrypt(plaintext, key)
	if err != nil {
		fmt.Println("Error encrypting:", err)
		os.Exit(1)
	}

	fmt.Println("Ciphertext:", ciphertext)
	fmt.Println("key:", key)
	

	// Decrypt
	decrypted, err := decrypt(ciphertext, key)
	if err != nil {
		fmt.Println("Error decrypting:", err)
		os.Exit(1)
	}

	fmt.Println("Decrypted:", string(decrypted))
}
