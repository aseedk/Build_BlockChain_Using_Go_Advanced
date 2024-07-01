package src

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
	lastHashKey = "lh"
)

// BlockChain struct which represents a single blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// InitBlockChain function to initialize a new blockchain
func InitBlockChain(address, nodeId string) *BlockChain {
	// Initialize variables
	var (
		lastHash []byte
		db       *badger.DB
		opts     badger.Options
		err      error
	)
	fmt.Println("NodeId", nodeId)
	dbFile := fmt.Sprintf(dbPath, nodeId)
	fmt.Println("DB File: ", dbFile)
	if DBexists(dbFile) {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	// Set the options for the badger database
	opts = badger.DefaultOptions
	opts.Dir = dbFile
	opts.ValueDir = dbFile

	// Open the badger database
	db, err = openDb(dbFile, opts)

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
func ContinueBlockChain(nodeId string) *BlockChain {
	// Initialize variables
	var (
		lastHash []byte
		db       *badger.DB
		item     *badger.Item
		opts     badger.Options
		err      error
	)
	path := fmt.Sprintf(dbPath, nodeId)
	if !DBexists(path) {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	// Set the options for the badger database
	opts = badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	// Open the badger database
	db, err = openDb(path, opts)

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

// AddBlock function to add a new block to the blockchain
func (blockChain *BlockChain) AddBlock(block *Block) {
	err := blockChain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte(lastHashKey))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)

		lastBlockData, _ := item.Value()
		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte(lastHashKey), block.Hash)
			Handle(err)
			blockChain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

// GetBlock function to get a block from the blockchain
func (blockChain *BlockChain) GetBlock(hash []byte) (Block, error) {
	var block Block

	err := blockChain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err != nil {
			return fmt.Errorf("block not found: %s", err)
		}

		encodedBlock, err := item.Value()
		block = *Deserialize(encodedBlock)
		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes function to get the hashes of all the blocks in the blockchain
func (blockChain *BlockChain) GetBlockHashes() [][]byte {
	var (
		iterator = blockChain.Iterator()
		hashes   [][]byte
	)

	for {
		block := iterator.Next()
		hashes = append(hashes, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return hashes
}

// GetBestHeight function to get the height of the last block in the blockchain
func (blockChain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := blockChain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHashKey))
		Handle(err)

		lastHash, _ := item.Value()
		item, err = txn.Get(lastHash)
		Handle(err)

		lastBlockData, _ := item.Value()

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})

	Handle(err)

	return lastBlock.Height
}

// MineBlock method to add a new block to the blockchain
func (blockChain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var (
		lastHash   []byte
		lastHeight int
		item       *badger.Item
		err        error
	)

	for _, tx := range transactions {
		if !blockChain.VerifyTransaction(tx) {
			fmt.Println("Invalid transaction")
		}
	}

	// Read the last hash from the database
	err = blockChain.Database.View(func(txn *badger.Txn) error {
		// Get the last hash from the database
		item, err = txn.Get([]byte(lastHashKey))

		// Handle the error
		Handle(err)

		// set the last hash to the hash of the last block
		lastHash, err = item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)

		lastBlockData, _ := item.Value()
		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height

		// return err if any
		return err
	})

	// Handle the error
	Handle(err)

	// Create a new block with the data and the last hash
	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

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

// retry function to retry a function if it fails
func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf("removing lock: %s", err)
	}

	// Set the options for the badger database
	retryOpts := originalOpts
	retryOpts.Truncate = true

	// Open the badger database
	db, err := badger.Open(retryOpts)

	// Handle the error
	Handle(err)

	// Return the database
	return db, err
}

// openDb function to open the database
func openDb(dir string, opts badger.Options) (*badger.DB, error) {
	db, err := badger.Open(opts)
	if err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err = retry(dir, opts); err == nil {
				log.Println("database unlocked, retrying open")
				return db, nil
			}
		}
		return nil, err
	}
	return db, nil
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}
