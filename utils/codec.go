package utils

func (nh *NoiseHandler) encode (data []byte) []byte {
  var headerLength int
  var dataLength int = len(data)
  if nh.sentIntro {
    headerLength = 0
  } else {
    headerLength = len(WA_NOISE_HEADER)
  }

  var encodedData []byte = make([]byte, headerLength + FRAME_LENGTH_HEADER + len(data))
  encodedData = append(encodedData, WA_NOISE_HEADER...)
  encodedData[headerLength] = byte(dataLength >> 16)
  encodedData[headerLength + 1] = byte(dataLength >> 8)
  encodedData[headerLength + 2] = byte(dataLength)
  copy (encodedData[headerLength + FRAME_LENGTH_HEADER:], data)

  return encodedData
}

func (nh *NoiseHandler) decode (data []byte) []byte {
  // We don't need the length of the frame for now
  return data[FRAME_LENGTH_HEADER:]
}
