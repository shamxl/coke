package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/shamxl/coke/waproto"
	"github.com/shamxl/coke/websocket"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"google.golang.org/protobuf/proto"
)

type Keypair struct {
  PublicKey [32]byte
  PrivateKey [32]byte
}

func NewKeypair () *Keypair {
  var privateKey [32]byte
  var publicKey [32]byte
  rand.Read(privateKey[:])
  curve25519.ScalarBaseMult(&publicKey, &privateKey)
  return &Keypair{
    PrivateKey: privateKey,
    PublicKey: publicKey, 
  }
}

func SHA256Sum (data []byte) []byte {
  var hash [32]byte = sha256.Sum256(data)
  return hash[:]
}

func GenerateIV (count uint32) []byte {
  var iv []byte = make([]byte, 4)
  binary.BigEndian.PutUint32(iv, count)
  return iv
}

func ExtractAndExpand (salt, data []byte) ([]byte, []byte, error) {
  var hkdf = hkdf.New(sha256.New, data, salt, nil)
  var read []byte = make([]byte, 32)
  var write []byte = make([]byte, 32)
  if _, err := io.ReadFull(hkdf, write); err != nil {
    return nil, nil, err
  }  
  if _, err := io.ReadFull(hkdf, read); err != nil {
    return nil, nil, err
  }
  return write, read, nil
}


func NewAESGCM (secret []byte) (gcm cipher.AEAD, err error) {
  var block cipher.Block
  block, err = aes.NewCipher(secret)
  if err != nil {
    return
  }
  gcm, err = cipher.NewGCM(block)

  return
}

type NoiseHandler struct {
  hash []byte
  salt []byte
  key cipher.AEAD
  ClientKeypair *Keypair
  ServerKeypair *Keypair
  counter uint32
}


func (nh *NoiseHandler) Authenticate (data []byte) {
  nh.hash = SHA256Sum(append(nh.hash, data...))
}

func (nh *NoiseHandler) MixIntoKey (data []byte) {
  nh.counter = 0
  var write []byte
  var read []byte
  var err error
  write, read, err = ExtractAndExpand(nh.salt, data)
  if err != nil {
    panic (err)
  }

  nh.salt = write
  
  var key cipher.AEAD
  key, err = NewAESGCM(read)
  if err != nil {
    panic (err)
  }

  nh.key = key
}


func (nh *NoiseHandler) MixSharedKey (privateKey, publicKey [32]byte) {
  var secret []byte
  var err error
  secret, err = curve25519.X25519(privateKey[:], publicKey[:])
  if err != nil {
    panic (err)
  }

  nh.MixIntoKey(secret)
}


func (nh *NoiseHandler) PostIncrementCounter () uint32 {
  var count uint32 = atomic.AddUint32(&nh.counter, 1)
  return count - 1
}

func (nh *NoiseHandler) Decrypt (ciphertext []byte) []byte {
  var plaintext []byte
  var err error

  plaintext, err = nh.key.Open(nil, GenerateIV(nh.PostIncrementCounter()), ciphertext, nh.hash)
  if err != nil {
    panic (err)
  }

  return plaintext
}

func (nh *NoiseHandler) Encrypt (plaintext []byte) []byte {
  var ciphertext []byte = nh.key.Seal(nil, GenerateIV(nh.PostIncrementCounter()), plaintext, nh.hash)
  nh.Authenticate(ciphertext)
  return ciphertext
}

func (nh *NoiseHandler) StartHandshake () {
  var mode []byte = []byte(NOISE_MODE)
  if len(mode) == 32 {
    nh.hash = mode
  } else {
    nh.hash = SHA256Sum(mode)
  }

  nh.salt = nh.hash

  var key cipher.AEAD
  var err error
  key, err = NewAESGCM(nh.hash)
  if err != nil {
    panic (err)
  }
  nh.key = key
  nh.Authenticate(WA_NOISE_HEADER)
}


func (nh *NoiseHandler) ProcessHandshake (args *websocket.Args) {
  var err error
  var handshakeMessage *waproto.HandshakeMessage
  err = proto.Unmarshal(args.Message, handshakeMessage)
  if err != nil {
    panic ("Failed to decode handshake message")
  }
  var serverHello *waproto.HandshakeMessage_ServerHello = handshakeMessage.GetServerHello()
  if serverHello == nil {
    panic("Handshake failed, Empty response.")
  }

  // TODO: save these keys in config
  var serverEphmeralKey []byte = serverHello.GetEphemeral()
  var serverEncryptedStaticKey []byte = serverHello.GetStatic()
  var serverEncryptedCertificate []byte = serverHello.GetPayload()

  nh.Authenticate(serverEphmeralKey)
  nh.MixSharedKey(nh.ClientKeypair.PrivateKey, [32]byte(serverEphmeralKey))
  var serverDecryptedStaticKey []byte = nh.Decrypt(serverEncryptedStaticKey)
  nh.MixSharedKey(nh.ClientKeypair.PrivateKey, [32]byte(serverDecryptedStaticKey))
  // TODO: implement server certificate verifier
  var s []byte = nh.Decrypt(serverEncryptedCertificate) // Decrypted certificate of the server
  fmt.Print(s) // :187
}
