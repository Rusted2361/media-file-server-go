// Input string
const inputString = "<89 50 4e 47 0d 0a 1a 0a 00 00 00 0d 49 48 44 52 00 00 0d 20 00 00 06 c0 08 06 00 00 00 cd 7a 19 4b 00 00 0c 6b 69 43 43 50 49 43 43 20 50 72 6f 66 69 6c 65 00 00 48 89 95 57 07 58 53 c9 16 9e 5b 92 90 90 d0 02 11 90 12 7a 13 44 7a 91 12 42 8b 20 20 55 b0 11 92 40 42 89 31 21 a8 d8 91 45 05 d7 2e 22>";
// Remove '<' and '>' and split the string into an array of hexadecimal values
const hexValues = inputString.slice(1, -1).split(' ');

// Convert hexadecimal values to integers
const byteValues = hexValues.map(hex => parseInt(hex, 16));

// Print the resulting byte array
console.log(byteValues);
