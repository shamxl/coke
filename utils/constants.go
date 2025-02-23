package utils


const (
  NOISE_MODE string = "Noise_XX_25519_AESGCM_SHA256\x00\x00\x00\x00"
  WA_MAGIC_NUMBER byte = 6
  WA_DICT_VERSION byte = 2
)

var WA_NOISE_HEADER []byte = []byte{'W', 'A', WA_MAGIC_NUMBER, WA_DICT_VERSION}
