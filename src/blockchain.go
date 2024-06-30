package src

// BlockChain struct which represents a single blockchain
type BlockChain struct {
	Blocks []*Block
}

// InitBlockChain function to initialize a new blockchain
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{CreateGenesisBlock()}}
}

// CreateGenesisBlock function to create the first block in the blockchain
func CreateGenesisBlock() *Block {
	return CreateBlock("Genesis", []byte{})
}

// AddBlock method to add a new block to the blockchain
func (bc *BlockChain) AddBlock(data string) {
	// Get the previous block
	prevBlock := bc.Blocks[len(bc.Blocks)-1]

	// Create a new block with the provided data and previous hash
	newBlock := CreateBlock(data, prevBlock.Hash)

	// Append the new block to the blockchain
	bc.Blocks = append(bc.Blocks, newBlock)
}
