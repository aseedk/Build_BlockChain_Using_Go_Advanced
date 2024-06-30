package src

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

// Block struct which contains data of a block in the blockchain
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

// CreateBlock function to create a new block in the blockchain
func CreateBlock(transactions []*Transaction, prevHash []byte) *Block {
	// Create a new block with the provided data and previous hash
	block := &Block{[]byte{}, transactions, prevHash, 0}

	// Create a proof of work for the block
	proofOfWork := NewProof(block)

	// Run the proof of work algorithm to get the nonce and hash of the block
	block.Nonce, block.Hash = proofOfWork.Run()

	// Return the block
	return block
}

// HashTransactions function to hash the transactions in the block
func (b *Block) HashTransactions() []byte {
	// Initialize a new bytes buffer
	var transactionsHashes [][]byte
	var transactionHash [32]byte

	// Iterate over the transactions in the block
	for _, transaction := range b.Transactions {
		transactionsHashes = append(transactionsHashes, transaction.ID)
	}

	// Create a new hash with the hashes of the transactions
	transactionHash = sha256.Sum256(bytes.Join(transactionsHashes, []byte{}))

	// Return the hash of the transactions
	return transactionHash[:]
}

// Serialize function to serialize the block
func (b *Block) Serialize() []byte {
	// Initialize a new bytes buffer
	var res bytes.Buffer

	// Create a new gob encoder with the bytes buffer
	encoder := gob.NewEncoder(&res)

	// Encode the block
	err := encoder.Encode(b)
	Handle(err)

	// Return the bytes of the serialized block
	return res.Bytes()
}

// Deserialize function to deserialize the block
func Deserialize(data []byte) *Block {
	// Initialize a new block
	var block Block

	// Create a new gob decoder with the data
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// Decode the data into the block
	err := decoder.Decode(&block)
	Handle(err)

	// Return the block
	return &block
}
