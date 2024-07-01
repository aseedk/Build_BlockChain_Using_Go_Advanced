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
		val      []byte
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
		val, err = item.Value()

		// Handle the error
		Handle(err)

		lastHash = append([]byte{}, val...)

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
		val      []byte
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
		val, err = item.Value()

		// Handle the error
		Handle(err)

		lastHash = append([]byte{}, val...)

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
func (blockChain *BlockChain) FindUnspentTransactions(publicKeyHash []byte) []Transaction {
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

				if output.IsLockedWithKey(publicKeyHash) {
					unspentTransactions = append(unspentTransactions, *transaction)
				}
			}

			if !transaction.IsCoinbase() {
				for _, input := range transaction.Inputs {
					if input.UsesKey(publicKeyHash) {
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
func (blockChain *BlockChain) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	// Initialize variables
	var (
		unspentOutputs      = make(map[string][]int)
		unspentTransactions = blockChain.FindUnspentTransactions(publicKeyHash)
		accumulated         = 0
	)

Work:
	for _, transaction := range unspentTransactions {
		transactionId := hex.EncodeToString(transaction.ID)

		for outputIndex, output := range transaction.Outputs {
			if output.IsLockedWithKey(publicKeyHash) && accumulated < amount {
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
func (blockChain *BlockChain) FindUnspentTransactionOutputs(publicKeyHash []byte) []TxOutput {
	var UnspentTransactionsOutputs []TxOutput

	// Find all the unspent transactions in the blockchain
	unspentTransactions := blockChain.FindUnspentTransactions(publicKeyHash)

	// Iterate over the unspent transactions and find the unspent transaction outputs
	for _, transaction := range unspentTransactions {
		// Iterate over the outputs of the transaction
		for _, output := range transaction.Outputs {
			if output.IsLockedWithKey(publicKeyHash) {
				UnspentTransactionsOutputs = append(UnspentTransactionsOutputs, output)
			}
		}
	}

	// Return the unspent transactions outputs
	return UnspentTransactionsOutputs
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
