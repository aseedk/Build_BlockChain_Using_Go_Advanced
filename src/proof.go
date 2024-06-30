package src

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

// Difficulty is the number of leading zeros that must be present in the hash of a block
const Difficulty = 12

// ProofOfWork struct which contains a block and a target
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// NewProof function to create a new proof of work
func NewProof(b *Block) *ProofOfWork {
	// Initialize a new big.Int with the value of 1
	target := big.NewInt(1)

	// Shift the value of 1 to the left by 256 - Difficulty
	target.Lsh(target, uint(256-Difficulty))

	// Create a new proof of work with the block and the target
	pow := &ProofOfWork{b, target}

	// Return the proof of work
	return pow
}

// Run function to run the proof of work algorithm
func (pow *ProofOfWork) Run() (int, []byte) {
	// Initialize a new big.Int and a byte array to store the hash
	var intHash big.Int
	var hash [32]byte

	// Initialize the nonce to 0
	nonce := 0

	// Iterate over the nonce until a hash with the required number of leading zeros is found
	for nonce < math.MaxInt64 {
		// Initialize the data of the proof of work
		data := pow.InitData(nonce)

		// Calculate the hash of the data
		hash = sha256.Sum256(data)

		// Print the hash in hexadecimal format
		fmt.Printf("\r%x", hash)

		// Set the hash to the big.Int
		intHash.SetBytes(hash[:])

		// Check if the hash is less than the target, if it is then break the loop else increment the nonce
		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Println()

	// Return the nonce and the hash of the block that satisfies the proof of work
	return nonce, hash[:]
}

// Validate function to validate the proof of work
func (pow *ProofOfWork) Validate() bool {
	// initialize a new big.Int to store the hash
	var intHash big.Int

	// Initialize the data of the proof of work using the nonce of the block
	data := pow.InitData(pow.Block.Nonce)

	// Calculate the hash of the data
	hash := sha256.Sum256(data)

	// Set the hash to the big.Int
	intHash.SetBytes(hash[:])

	// Check if the hash is less than the target
	return intHash.Cmp(pow.Target) == -1
}

// InitData function to initialize the data of the proof of work
func (pow *ProofOfWork) InitData(nonce int) []byte {
	// Join the previous hash, data, nonce, and difficulty to create the data
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	// Return the data
	return data
}

// ToHex function to convert integer64 to a byte slice
func ToHex(num int64) []byte {
	// Initialize a new buffer to store the bytes
	buff := new(bytes.Buffer)

	// Write the integer to the buffer in big endian format
	err := binary.Write(buff, binary.BigEndian, num)
	Handle(err)

	// Return the bytes
	return buff.Bytes()
}
