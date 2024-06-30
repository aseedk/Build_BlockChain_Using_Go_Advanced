package src

import (
	"bytes"
	"crypto/sha256"
)

// Block struct which contains data of a block in the blockchain
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

// DeriveHash method to derive the hash of the block
func (b *Block) DeriveHash() {
	// Concatenate the data and previous hash of the block
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})

	// Derive the hash of the block using SHA-256
	hash := sha256.Sum256(info)

	// Set the hash of the block
	b.Hash = hash[:]
}

// CreateBlock function to create a new block in the blockchain
func CreateBlock(data string, prevHash []byte) *Block {
	// Create a new block with the provided data and previous hash
	block := &Block{[]byte{}, []byte(data), prevHash}

	// Derive the hash of the block
	block.DeriveHash()

	// Return the block
	return block
}
