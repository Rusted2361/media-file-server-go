// Hexadecimal string
const hexString = "b849e8198b02345acb21856f6fca8f5d2fcbb18e0e410d4d9904fb8a1a6137a4349a5d3e9750cdfd43b30ad6a4bdb2c3";

// Convert hex string to ArrayBuffer
const arrayBuffer = hexStringToBuffer(hexString);

// Print the ArrayBuffer
console.log(arrayBuffer);

// Function to convert hex string to ArrayBuffer
function hexStringToBuffer(hexString) {
  const hexArray = hexString.match(/.{1,2}/g) || [];
  const byteBuffer = new Uint8Array(hexArray.map(byte => parseInt(byte, 16))).buffer;
  return byteBuffer;
}
