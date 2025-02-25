package utils

func (nh *NoiseHandler) encode (data []byte) []byte {
  var headerLength int = len (WA_NOISE_HEADER)
  var dataLength int = len(data)
  var frame []byte = make([]byte, headerLength + 3 + dataLength)
  if !nh.sentIntro {
    copy (frame[:headerLength], WA_NOISE_HEADER)
  }
  frame[headerLength] = byte(dataLength >> 16)
  frame[headerLength + 1] = byte(dataLength >> 8)
  frame[headerLength + 2] = byte(dataLength)
  copy(frame[headerLength + 3:], data)
  return frame
}

func (nh *NoiseHandler) decode (data []byte) []byte {
  // We don't need the length of the frame for now
  return data[FRAME_LENGTH_HEADER:]
}
