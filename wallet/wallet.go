package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

// Wallet struct which contains the private key and the public key
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Address function to generate the address of the wallet
func (wallet Wallet) Address() []byte {
	// Generate the public key hash
	publicKeyHash := PublicKeyHash(wallet.PublicKey)

	// Create a new payload with the version and the public key hash
	payload := append([]byte{version}, publicKeyHash...)

	// Generate the checksum
	checksum := Checksum(payload)

	// Create a new payload with the checksum
	fullPayload := append(payload, checksum...)

	// Encode the payload with base58
	address := Base58Encode(fullPayload)

	fmt.Printf("Public key: %x\n", wallet.PublicKey)
	fmt.Printf("Public key hash: %x\n", publicKeyHash)
	fmt.Printf("Address: %x\n", address)

	// Return the address
	return address

}

// NewKeyPair function to create a new key pair
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	// Generate a curve which will be used to generate the key pair
	curve := elliptic.P256()

	// Generate a privateKey key using the curve and the random reader
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)

	// Handle the error
	if err != nil {
		log.Panic(err)
	}
	// Get the publicKey key from the privateKey key
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	// Return the privateKey key and the publicKey key
	return *privateKey, publicKey
}

// MakeWallet function to create a new wallet
func MakeWallet() *Wallet {
	// Generate a new key pair
	privateKey, publicKey := NewKeyPair()

	// Create a new wallet with the key pair
	wallet := Wallet{privateKey, publicKey}

	// Return the wallet
	return &wallet
}

// PublicKeyHash function to generate the public key hash
func PublicKeyHash(publicKey []byte) []byte {
	// Generate the public key hash
	publicKeyHash := sha256.Sum256(publicKey)

	// Initialize a new ripemd160 hasher
	hasher := ripemd160.New()

	// Write the public key hash to the hasher
	_, err := hasher.Write(publicKeyHash[:])

	// Handle the error
	if err != nil {
		log.Panic(err)
	}

	// Get the public key hash
	publicRipeMd := hasher.Sum(nil)

	// Return the public key hash
	return publicRipeMd
}

// Checksum function to generate the checksum
func Checksum(payload []byte) []byte {
	// Generate the first checksum
	firstChecksum := sha256.Sum256(payload)

	// Generate the second checksum
	secondChecksum := sha256.Sum256(firstChecksum[:])

	// Return the first four bytes of the second checksum
	return secondChecksum[:checksumLength]
}

func (wallet Wallet) MarshalJSON() ([]byte, error) {
	mapStringAny := map[string]any{
		"PrivateKey": map[string]any{
			"D": wallet.PrivateKey.D,
			"PublicKey": map[string]any{
				"X": wallet.PrivateKey.PublicKey.X,
				"Y": wallet.PrivateKey.PublicKey.Y,
			},
			"X": wallet.PrivateKey.X,
			"Y": wallet.PrivateKey.Y,
		},
		"PublicKey": wallet.PublicKey,
	}
	return json.Marshal(mapStringAny)
}
