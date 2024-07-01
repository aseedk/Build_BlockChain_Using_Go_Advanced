package wallet

import (
	"bytes"
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

// MakeWallet function to create a new wallet
func MakeWallet() *Wallet {
	// Generate a new key pair
	privateKey, publicKey := NewKeyPair()

	// Create a new wallet with the key pair
	wallet := Wallet{privateKey, publicKey}

	// Return the wallet
	return &wallet
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

// MarshalJSON function to marshal the wallet to JSON
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

// ValidateAddress function to validate an address
func ValidateAddress(address string) bool {
	// Decode the address with base58
	publicKeyHash := Base58Decode([]byte(address))

	// Get the checksum
	actualChecksum := publicKeyHash[len(publicKeyHash)-checksumLength:]

	// Get the version
	ver := publicKeyHash[0]

	// Get the publicKeyHash without the checksum
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-checksumLength]

	// Get Target checksum
	targetChecksum := Checksum(append([]byte{ver}, publicKeyHash...))

	// Compare the checksums and return the result
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// NewKeyPair function to create a new key pair
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	// Generate a curve which will be used to generate the key pair
	curve := elliptic.P256()

	// Generate a privateKey key using the curve and the random reader
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	privateKey.PublicKey.Curve = curve

	// Handle the error
	if err != nil {
		log.Panic(err)
	}
	// Get the publicKey key from the privateKey key
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	// Return the privateKey key and the publicKey key
	return *privateKey, publicKey
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
