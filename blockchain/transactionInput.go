package src

import (
	"Build_BlockChain_Using_Go_Advanced/wallet"
	"bytes"
)

// TxInput struct which contains the ID of the transaction, the output and the signature
type TxInput struct {
	ID        []byte // ID of the transaction
	Out       int    // Output of the transaction
	Signature []byte // Signature of the transaction
	PublicKey []byte // Public key of the transaction
}

// UsesKey function to check if the transaction input uses the provided public key
func (in *TxInput) UsesKey(publicKeyHash []byte) bool {
	// Get the hash of the public key
	lockingHash := wallet.PublicKeyHash(in.PublicKey)

	// Check if the hash of the public key is equal to the provided public key hash
	return bytes.Compare(lockingHash, publicKeyHash) == 0
}
