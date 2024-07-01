package src

import (
	"Build_BlockChain_Using_Go_Advanced/wallet"
	"bytes"
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
