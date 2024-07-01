package src

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// Transaction struct which contains data of a transaction in the blockchain
type Transaction struct {
	ID      []byte     // Hash of the transaction
	Inputs  []TxInput  // Inputs of the transaction
	Outputs []TxOutput // Outputs of the transaction
}

// CoinbaseTransaction function to create a new coinbase transaction
func CoinbaseTransaction(to, data string) *Transaction {
	// If the data is empty, set it to the to address
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// Create a new transaction with the provided data
	transactionIn := TxInput{[]byte{}, -1, data}
	transactionOut := TxOutput{100, to}
	tx := Transaction{nil, []TxInput{transactionIn}, []TxOutput{transactionOut}}
	tx.SetID()

	// Return the transaction
	return &tx
}

// SetID function to set the ID of the transaction
func (tx *Transaction) SetID() {
	// Initialize a new bytes buffer and a hash
	var encoded bytes.Buffer
	var hash [32]byte

	// Create a new gob encoder with the bytes buffer
	encoder := gob.NewEncoder(&encoded)

	// Encode the transaction
	err := encoder.Encode(tx)
	Handle(err)

	// Calculate the hash of the encoded transaction
	hash = sha256.Sum256(encoded.Bytes())

	// Set the ID of the transaction to the hash
	tx.ID = hash[:]
}

// NewTransaction function to create a new transaction
func NewTransaction(from, to string, amount int, blockchain *BlockChain) *Transaction {
	// Initialize variables
	var (
		inputs  []TxInput
		outputs []TxOutput
	)

	acc, validOutputs := blockchain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TxOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// IsCoinbase function to check if the transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	// Check if the transaction has only one input and the ID of the input is empty
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}
