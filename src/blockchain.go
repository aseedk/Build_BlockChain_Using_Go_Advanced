package src

import (
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
	opts = badger.DefaultOptions(dbPath)
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
	opts = badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	opts.Logger = nil

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
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

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

// CreateGenesisBlock function to create the first block in the blockchain
func CreateGenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// AddBlock method to add a new block to the blockchain
func (blockChain *BlockChain) AddBlock(transactions []*Transaction) {
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
}

// FindUnspentTransactions function to find all the unspent transactions in the blockchain
func (blockChain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	// Initialize variables
	var (
		unspentTransactions     []Transaction
		spentTransactionOutputs = make(map[string][]int)
		iterator                = blockChain.Iterator()
	)

	for {
		block := iterator.Next()
		for _, transaction := range block.Transactions {
			transactionId := hex.EncodeToString(transaction.ID)
		Outputs:
			for outputIndex, output := range transaction.Outputs {
				if spentTransactionOutputs[transactionId] != nil {
					for _, spentOutput := range spentTransactionOutputs[transactionId] {
						if spentOutput == outputIndex {
							continue Outputs
						}
					}
				}

				if output.CanBeUnlocked(address) {
					unspentTransactions = append(unspentTransactions, *transaction)
				}
			}

			if !transaction.IsCoinbase() {
				for _, input := range transaction.Inputs {
					if input.CanUnlock(address) {
						inputId := hex.EncodeToString(input.ID)
						spentTransactionOutputs[inputId] = append(spentTransactionOutputs[inputId], input.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	// Return the unspent transactions
	return unspentTransactions
}

// FindSpendableOutputs function to find the spendable outputs in the blockchain
func (blockChain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	// Initialize variables
	var (
		unspentOutputs      = make(map[string][]int)
		unspentTransactions = blockChain.FindUnspentTransactions(address)
		accumulated         = 0
	)

Work:
	for _, transaction := range unspentTransactions {
		transactionId := hex.EncodeToString(transaction.ID)

		for outputIndex, output := range transaction.Outputs {
			if output.CanBeUnlocked(address) && accumulated < amount {
				accumulated += output.Value
				unspentOutputs[transactionId] = append(unspentOutputs[transactionId], outputIndex)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	// Iterate over the unspent transactions and find the spendable outputs
	return accumulated, unspentOutputs
}

// FindUnspentTransactionOutputs function to find all the unspent transaction outputs in the blockchain
func (blockChain *BlockChain) FindUnspentTransactionOutputs(address string) []TxOutput {
	var UnspentTransactionsOutputs []TxOutput

	// Find all the unspent transactions in the blockchain
	unspentTransactions := blockChain.FindUnspentTransactions(address)

	// Iterate over the unspent transactions and find the unspent transaction outputs
	for _, transaction := range unspentTransactions {
		// Iterate over the outputs of the transaction
		for _, output := range transaction.Outputs {
			if output.CanBeUnlocked(address) {
				UnspentTransactionsOutputs = append(UnspentTransactionsOutputs, output)
			}
		}
	}

	// Return the unspent transactions outputs
	return UnspentTransactionsOutputs
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
