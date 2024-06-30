package src

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	lastHashKey = "lh"
)

// BlockChain struct which represents a single blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// InitBlockChain function to initialize a new blockchain
func InitBlockChain() *BlockChain {
	// Initialize variables
	var (
		lastHash []byte
		item     *badger.Item
		db       *badger.DB
		opts     badger.Options
		err      error
	)

	// Set the options for the badger database
	opts = badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	// Open the badger database
	db, err = badger.Open(opts)

	// Handle the error
	Handle(err)

	// Update the database, if no existing blockchain is found, create a new one
	// If an existing blockchain is found, load it
	err = db.Update(func(txn *badger.Txn) error {
		// Check if the last hash exists in the database
		_, err = txn.Get([]byte(lastHashKey))

		// If the key is not found, create a new blockchain
		if errors.Is(badger.ErrKeyNotFound, err) {
			fmt.Println("No existing blockchain found")
			// Create the genesis block
			genesis := CreateGenesisBlock()
			fmt.Println("Genesis proved")

			// Set the genesis block in the database
			err = txn.Set(genesis.Hash, genesis.Serialize())

			// Handle the error
			Handle(err)

			// Set the last hash in the database
			err = txn.Set([]byte(lastHashKey), genesis.Hash)

			// set the last hash to the hash of the genesis block
			lastHash = genesis.Hash

			// return nil or error if any
			return err
		} else {
			// If the key is found, load the blockchain
			item, err = txn.Get([]byte(lastHashKey))

			// Handle the error
			Handle(err)

			// Get the last hash from the database and set it to the last hash
			err = item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})

			// return err if any
			return err
		}
	})

	// Handle the error
	Handle(err)

	// Create a new blockchain with the last hash and the database
	blockchain := BlockChain{lastHash, db}

	// Return the blockchain
	return &blockchain
}

// CreateGenesisBlock function to create the first block in the blockchain
func CreateGenesisBlock() *Block {
	return CreateBlock("Genesis", []byte{})
}

// AddBlock method to add a new block to the blockchain
func (blockChain *BlockChain) AddBlock(data string) {
	var (
		lastHash []byte
		item     *badger.Item
		err      error
	)

	// Read the last hash from the database
	err = blockChain.Database.View(func(txn *badger.Txn) error {
		// Get the last hash from the database
		item, err = txn.Get([]byte(lastHashKey))

		// Handle the error
		Handle(err)

		// set the last hash to the hash of the last block
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

		// return err if any
		return err
	})

	// Handle the error
	Handle(err)

	// Create a new block with the data and the last hash
	newBlock := CreateBlock(data, lastHash)

	// Update the database with the new block
	err = blockChain.Database.Update(func(txn *badger.Txn) error {
		// Set the new block in the database
		err = txn.Set(newBlock.Hash, newBlock.Serialize())

		// Handle the error
		Handle(err)

		// Set the last hash in the database
		err = txn.Set([]byte(lastHashKey), newBlock.Hash)

		// Set the last hash of the blockchain to the hash of the new block
		blockChain.LastHash = newBlock.Hash

		// return nil or error if any
		return err
	})
}

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
		err   error
	)

	// Read the block from the database
	err = iterator.Database.View(func(txn *badger.Txn) error {
		// Get the block from the database
		item, err = txn.Get(iterator.CurrentHash)

		// Handle the error
		Handle(err)

		// Decode the block from the database
		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
			return nil
		})

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
