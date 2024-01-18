// function arrayToBuffer(array) {
//     const buffer = new ArrayBuffer(array.length);
//     const view = new Uint8Array(buffer);
  
//     for (let i = 0; i < array.length; i++) {
//       view[i] = array[i];
//     }
  
//     return buffer;
//   }
  
//   const inputArray = [26 144 180 154 143 101 15 144 231 98 160 247 196 211 50 31 79 5 128 162 233 51 230 241 240 117 179 127 174 3 232 208];
  
//   const outputBuffer = arrayToBuffer(inputArray);
  
//   console.log(outputBuffer);
function arrayToBuffer(array) {
    const buffer = new ArrayBuffer(array.length);
    const view = new Uint8Array(buffer);
  
    for (let i = 0; i < array.length; i++) {
      view[i] = array[i];
    }
  
    return buffer;
  }
  
  const inputString = "[62 101 73 53 109 72 78 78 111 109 49 52 90 101 49 80 82 119 97 106 85 102 98 110 116 71 98 106 74 69 49 104]";
  const inputArray = inputString.match(/\d+/g).map(Number);
  
  const outputBuffer = arrayToBuffer(inputArray);
  
  console.log(outputBuffer);
  