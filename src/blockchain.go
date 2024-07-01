package src

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"runtime"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
	lastHashKey = "lh"
)

// BlockChain struct which represents a single blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// InitBlockChain function to initialize a new blockchain
func InitBlockChain(address string) *BlockChain {
	// Initialize variables
	var (
		lastHash []byte
		db       *badger.DB
		opts     badger.Options
		err      error
	)

	if FileExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	// Set the options for the badger database
	opts = badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	// Open the badger database
	db, err = badger.Open(opts)

	// Handle the error
	Handle(err)

	// Update the database with the genesis block
	err = db.Update(func(txn *badger.Txn) error {
		// Initialize coinbase transaction
		coinbaseTransaction := CoinbaseTransaction(address, genesisData)

		// Create the genesis block with the coinbase transaction
		genesis := CreateGenesisBlock(coinbaseTransaction)
		fmt.Println("Genesis Created")

		// Set the genesis block in the database
		err = txn.Set(genesis.Hash, genesis.Serialize())

		// Handle the error
		Handle(err)

		// Set the last hash in the database
		err = txn.Set([]byte(lastHashKey), genesis.Hash)

		// Set the last hash to the hash of the genesis block
		lastHash = genesis.Hash

		// Return nil or error if any
		return err
	})

	// Handle the error
	Handle(err)

	// Create a new blockchain with the last hash and the database
	blockchain := BlockChain{lastHash, db}

	// Return the blockchain
	return &blockchain
}

// ContinueBlockChain function to continue an existing blockchain
func ContinueBlockChain(_ string) *BlockChain {
	// Initialize variables
	var (
		lastHash []byte
		db       *badger.DB
		item     *badger.Item
		opts     badger.Options
		err      error
	)

	if !FileExists(dbFile) {
		runtime.Goexit()
	}

	// Set the options for the badger database
	opts = badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	// Open the badger database
	db, err = badger.Open(opts)

	// Handle the error
	Handle(err)

	// Update the database with the genesis block
	err = db.Update(func(txn *badger.Txn) error {
		// Get the last hash from the database
		item, err = txn.Get([]byte(lastHashKey))

		// Handle the error
		Handle(err)

		// Set the last hash to the hash of the last block
		lastHash, err = item.Value()

		// Handle the error
		Handle(err)

		// Return err if any
		return err
	})

	// Handle the error
	Handle(err)

	// Create a new blockchain with the last hash and the database
	blockchain := BlockChain{lastHash, db}

	// Return the blockchain
	return &blockchain
}

// AddBlock method to add a new block to the blockchain
func (blockChain *BlockChain) AddBlock(transactions []*Transaction) *Block {
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
		lastHash, err = item.Value()

		// return err if any
		return err
	})

	// Handle the error
	Handle(err)

	// Create a new block with the data and the last hash
	newBlock := CreateBlock(transactions, lastHash)

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

	return newBlock
}

// FindUTXOs function to find all the unspent transactions in the blockchain
func (blockChain *BlockChain) FindUTXOs() map[string]TxOutputs {
	// Initialize variables
	var (
		UTXOs     = make(map[string]TxOutputs)
		spentTXOs = make(map[string][]int)
		iterator  = blockChain.Iterator()
	)

	for {
		block := iterator.Next()
		for _, transaction := range block.Transactions {
			transactionId := hex.EncodeToString(transaction.ID)
		Outputs:
			for outputIndex, output := range transaction.Outputs {
				if spentTXOs[transactionId] != nil {
					for _, spentOutput := range spentTXOs[transactionId] {
						if spentOutput == outputIndex {
							continue Outputs
						}
					}
				}

				outs := UTXOs[transactionId]
				outs.Outputs = append(outs.Outputs, output)
				UTXOs[transactionId] = outs
			}

			if !transaction.IsCoinbase() {
				for _, input := range transaction.Inputs {
					inputId := hex.EncodeToString(input.ID)
					spentTXOs[inputId] = append(spentTXOs[inputId], input.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	// Return the unspent transactions
	return UTXOs
}

// FindTransaction function to find a transaction in the blockchain
func (blockChain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	// Initialize variables
	var (
		iterator = blockChain.Iterator()
	)

	// Iterate over the blocks in the blockchain
	for {
		block := iterator.Next()

		// Iterate over the transactions in the block
		for _, transaction := range block.Transactions {
			if bytes.Compare(transaction.ID, ID) == 0 {
				return *transaction, nil
			}
		}

		// Break the loop if the previous hash is empty
		if len(block.PrevHash) == 0 {
			break
		}
	}

	// Return an error if the transaction is not found
	return Transaction{}, fmt.Errorf("transaction %x not found", ID)
}

// SignTransaction function to sign a transaction
func (blockChain *BlockChain) SignTransaction(transaction *Transaction, privateKey ecdsa.PrivateKey) {
	// Initialize variables
	var (
		prevTXs = make(map[string]Transaction)
	)

	// Iterate over the inputs of the transaction
	for _, input := range transaction.Inputs {
		// Get the previous transaction from the blockchain
		prevTX, err := blockChain.FindTransaction(input.ID)

		// Handle the error
		Handle(err)

		// Set the previous transaction in the map
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	// Sign the transaction
	transaction.Sign(privateKey, prevTXs)
}

// VerifyTransaction function to verify a transaction
func (blockChain *BlockChain) VerifyTransaction(transaction *Transaction) bool {
	if transaction.IsCoinbase() {
		return true
	}

	// Initialize variables
	var (
		prevTXs = make(map[string]Transaction)
	)

	// Iterate over the inputs of the transaction
	for _, input := range transaction.Inputs {
		// Get the previous transaction from the blockchain
		prevTX, err := blockChain.FindTransaction(input.ID)

		// Handle the error
		Handle(err)

		// Set the previous transaction in the map
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	// Verify the transaction
	return transaction.Verify(prevTXs)
}

// Iterator method to create a new iterator for the blockchain
func (blockChain *BlockChain) Iterator() *BlockChainIterator {
	// Create a new iterator with the last hash and the database
	iterator := &BlockChainIterator{blockChain.LastHash, blockChain.Database}

	// Return the iterator
	return iterator
}
