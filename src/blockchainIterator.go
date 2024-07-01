package src

import "github.com/dgraph-io/badger"

// BlockChainIterator struct to iterate over the blockchain
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// Iterator method to create a new iterator for the blockchain
func (blockChain *BlockChain) Iterator() *BlockChainIterator {
	// Create a new iterator with the last hash and the database
	iterator := &BlockChainIterator{blockChain.LastHash, blockChain.Database}

	// Return the iterator
	return iterator
}

// Next method to move to the next block in the blockchain
func (iterator *BlockChainIterator) Next() *Block {
	var (
		block *Block
		item  *badger.Item
		val   []byte
		err   error
	)

	// Read the block from the database
	err = iterator.Database.View(func(txn *badger.Txn) error {
		// Get the block from the database
		item, err = txn.Get(iterator.CurrentHash)

		// Handle the error
		Handle(err)

		// Decode the block from the database
		val, err = item.Value()

		// Handle the error
		Handle(err)

		block = Deserialize(val)

		// return err if any
		return err
	})

	// Handle the error
	Handle(err)

	// Move to the previous block
	iterator.CurrentHash = block.PrevHash

	// Return the block
	return block
}
