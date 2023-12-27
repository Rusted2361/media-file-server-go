package main

import (
	"encoding/hex"
	"fmt"
)

func main() {
	// Hexadecimal string
	hexString := "b849e8198b02345acb21856f6fca8f5d2fcbb18e0e410d4d9904fb8a1a6137a4349a5d3e9750cdfd43b30ad6a4bdb2c3"

	// Convert hex string to byte slice
	byteSlice, err := hex.DecodeString(hexString)
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return
	}

	// Print the byte slice
	fmt.Printf("Byte slice: %v\n", byteSlice)

	// Convert byte slice to ArrayBuffer-like structure
	arrayBuffer := Uint8ArrayFromSlice(byteSlice)

	// Print the ArrayBuffer-like structure
	fmt.Printf("ArrayBuffer-like structure: %v\n", arrayBuffer)
}

// Uint8ArrayFromSlice converts a Go byte slice to an ArrayBuffer-like structure
func Uint8ArrayFromSlice(slice []byte) []uint8 {
	arrayBuffer := make([]uint8, len(slice))
	for i, b := range slice {
		arrayBuffer[i] = b
	}
	return arrayBuffer
}
