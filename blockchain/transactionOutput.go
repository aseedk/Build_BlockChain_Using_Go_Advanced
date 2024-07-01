package src

import (
	"Build_BlockChain_Using_Go_Advanced/wallet"
	"bytes"
	"encoding/gob"
)

// TxOutput struct which contains the value and the public key of the input
type TxOutput struct {
	Value         int    // Value of the output
	PublicKeyHash []byte // Public key hash of the output
}

// NewTxOutput function to create a new transaction output
func NewTxOutput(value int, address string) *TxOutput {
	// Create a new transaction output with the value
	transactionOutput := &TxOutput{value, nil}

	// Lock the output with the address
	transactionOutput.Lock([]byte(address))

	// Return the output
	return transactionOutput

}

// Lock function to lock the output with the provided address
func (out *TxOutput) Lock(address []byte) {
	// Decode the address
	publicKeyHash := wallet.Base58Decode(address)

	// Remove the version and checksum from the public key hash
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-4]

	// Set the public key hash of the output
	out.PublicKeyHash = publicKeyHash
}

// IsLockedWithKey function to check if the output is locked with the provided public key hash
func (out *TxOutput) IsLockedWithKey(publicKeyHash []byte) bool {
	// Check if the public key hash of the output is equal to the provided public key hash
	return bytes.Compare(out.PublicKeyHash, publicKeyHash) == 0
}

// TxOutputs struct which contains the outputs of the transaction
type TxOutputs struct {
	Outputs []TxOutput // Outputs of the transaction
}

// Serialize function to serialize the transaction outputs
func (outs *TxOutputs) Serialize() []byte {
	// Initialize a new bytes buffer
	var encoded bytes.Buffer

	// Create a new gob encoder with the bytes buffer
	encoder := gob.NewEncoder(&encoded)

	// Encode the outputs
	err := encoder.Encode(outs)
	Handle(err)

	// Return the bytes of the buffer
	return encoded.Bytes()
}

// DeserializeOutputs function to deserialize the transaction outputs
func DeserializeOutputs(data []byte) *TxOutputs {
	// Initialize a new transaction outputs
	var outputs TxOutputs

	// Create a new gob decoder with the data
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// Decode the data
	err := decoder.Decode(&outputs)
	Handle(err)

	// Return the outputs
	return &outputs
}
