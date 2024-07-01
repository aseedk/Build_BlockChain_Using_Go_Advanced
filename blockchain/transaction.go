package src

import (
	"Build_BlockChain_Using_Go_Advanced/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

// Transaction struct which contains data of a transaction in the blockchain
type Transaction struct {
	ID      []byte     // Hash of the transaction
	Inputs  []TxInput  // Inputs of the transaction
	Outputs []TxOutput // Outputs of the transaction
}

// IsCoinbase function to check if the transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	// Check if the transaction has only one input and the ID of the input is empty
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// Serialize function to serialize the transaction
func (tx *Transaction) Serialize() []byte {
	// Initialize a new bytes buffer
	var encoded bytes.Buffer

	// Create a new gob encoder with the bytes buffer
	encoder := gob.NewEncoder(&encoded)

	// Encode the transaction
	err := encoder.Encode(tx)
	Handle(err)

	// Return the bytes buffer
	return encoded.Bytes()
}

// Hash function to hash the transaction
func (tx *Transaction) Hash() []byte {
	// Set the ID of the transaction to nil
	txCopy := *tx
	txCopy.ID = []byte{}

	// Calculate the hash of the serialized transaction
	hash := sha256.Sum256(txCopy.Serialize())

	// Return the hash
	return hash[:]
}

// Sign function to sign the transaction
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	// Check if the transaction is a coinbase transaction
	if tx.IsCoinbase() {
		return
	}
	fmt.Println(tx.Inputs)
	// Iterate over the inputs of the transaction
	for _, in := range tx.Inputs {
		// Check if the previous transaction is not in the map
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {

			log.Panic("Error: Previous transaction is not correct")
		}
	}
	// Initialize a new transaction copy
	txCopy := tx.TrimmedCopy()

	// Iterate over the inputs of the transaction
	for inID, in := range txCopy.Inputs {
		// Get the previous transaction
		prevTx := prevTXs[hex.EncodeToString(in.ID)]

		// Set the public key of the input to the public key of the previous transaction
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PublicKey = prevTx.Outputs[in.Out].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inID].PublicKey = nil

		// Sign the transaction
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered in f", r)
				}
			}()

			r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
			// Handle the error
			Handle(err)

			// Set the signature of the input to the signature
			signature := append(r.Bytes(), s.Bytes()...)
			tx.Inputs[inID].Signature = signature

		}()

	}
}

// TrimmedCopy function to create a trimmed copy of the transaction
func (tx *Transaction) TrimmedCopy() Transaction {
	// Initialize variables
	var inputs []TxInput
	var outputs []TxOutput

	// Iterate over the inputs of the transaction
	for _, in := range tx.Inputs {
		// Append the input to the inputs
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	// Iterate over the outputs of the transaction
	for _, out := range tx.Outputs {
		// Append the output to the outputs
		outputs = append(outputs, TxOutput{out.Value, out.PublicKeyHash})
	}

	// Create a new transaction with the inputs and outputs
	txCopy := Transaction{tx.ID, inputs, outputs}

	// Return the transaction
	return txCopy

}

// Verify function to verify the transaction
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	// Check if the transaction is a coinbase transaction
	if tx.IsCoinbase() {
		return true
	}

	// Iterate over the inputs of the transaction
	for _, in := range tx.Inputs {
		// Check if the previous transaction is not in the map
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Error: Previous transaction is not correct")
		}
	}

	// Initialize a new transaction copy
	txCopy := tx.TrimmedCopy()

	// Create a new elliptic curve
	curve := elliptic.P256()

	// Iterate over the inputs of the transaction
	for inID, in := range tx.Inputs {
		// Get the previous transaction
		prevTx := prevTXs[hex.EncodeToString(in.ID)]

		// Set the public key of the input to the public key of the previous transaction
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PublicKey = prevTx.Outputs[in.Out].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inID].PublicKey = nil

		// Get the r and s values of the signature
		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		// Get the x and y values of the public key
		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PublicKey)
		x.SetBytes(in.PublicKey[:(keyLen / 2)])
		y.SetBytes(in.PublicKey[(keyLen / 2):])

		// Get the signature
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		// Verify the signature
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	// Return true if the transaction is verified
	return true
}

// String function to return the string representation of the transaction
func (tx *Transaction) String() string {
	// Initialize variables
	var lines []string

	// Add the transaction ID to the lines
	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	// Iterate over the inputs of the transaction
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PublicKey: %x", input.PublicKey))
	}

	// Iterate over the outputs of the transaction
	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PublicKeyHash))
	}

	// Return the lines
	return strings.Join(lines, "\n")

}

// NewTransaction function to create a new transaction
func NewTransaction(from, to string, amount int, UTXO *UTXOSet) *Transaction {
	// Initialize variables
	var (
		inputs  []TxInput
		outputs []TxOutput
	)

	// Create Wallet
	wallets, err := wallet.CreateWallets()

	// Handle the error
	Handle(err)

	// Get the wallet from the wallets
	fromWallet := wallets.GetWallet(from)

	// Get the public key hash of the from wallet
	fromPublicKeyHash := wallet.PublicKeyHash(fromWallet.PublicKey)

	acc, validOutputs := UTXO.FindSpendableOutputs(fromPublicKeyHash, amount)

	if acc < amount {
		log.Panic("Error: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, fromWallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	UTXO.Blockchain.SignTransaction(&tx, fromWallet.PrivateKey)

	return &tx
}

// CoinbaseTransaction function to create a new coinbase transaction
func CoinbaseTransaction(to, data string) *Transaction {
	// If the data is empty, set it to the to address
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		Handle(err)
		data = fmt.Sprintf("%x", randData)
	}

	// Create a new transaction with the provided data
	transactionIn := TxInput{[]byte{}, -1, nil, []byte(data)}
	transactionOut := NewTxOutput(20, to)
	tx := Transaction{nil, []TxInput{transactionIn}, []TxOutput{*transactionOut}}
	tx.ID = tx.Hash()

	// Return the transaction
	return &tx
}
