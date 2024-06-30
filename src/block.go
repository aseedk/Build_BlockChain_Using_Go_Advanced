package src

// Block struct which contains data of a block in the blockchain
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// CreateBlock function to create a new block in the blockchain
func CreateBlock(data string, prevHash []byte) *Block {
	// Create a new block with the provided data and previous hash
	block := &Block{[]byte{}, []byte(data), prevHash, 0}

	// Create a proof of work for the block
	proofOfWork := NewProof(block)

	// Run the proof of work algorithm to get the nonce and hash of the block
	block.Nonce, block.Hash = proofOfWork.Run()

	// Return the block
	return block
}
