package src

import (
	"bytes"
	"encoding/hex"
	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *BlockChain
}

// DeleteByPrefix function to delete the UTXOs with the provided prefix
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.Blockchain.Database.Update(func(tx *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := tx.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	// Set collect size
	collectSize := 100000
	u.Blockchain.Database.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := tx.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected >= collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					panic(err)
				}
				keysForDelete = make([][]byte, 0, collectSize)
			}
		}
		if len(keysForDelete) > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				panic(err)
			}
		}
		return nil
	})
}

// Reindex function to reindex the UTXOs
func (u *UTXOSet) Reindex() {
	db := u.Blockchain.Database
	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.Blockchain.FindUTXOs()

	err := db.Update(func(tx *badger.Txn) error {
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			Handle(err)
			key = append(utxoPrefix, key...)

			err = tx.Set(key, outs.Serialize())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

// Update function to update the UTXOs
func (u *UTXOSet) Update(block *Block) {
	err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inId := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inId)
					Handle(err)

					v, err := item.Value()
					Handle(err)

					outs := DeserializeOutputs(v)

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err = txn.Delete(inId)
						Handle(err)
					} else {
						err = txn.Set(inId, updatedOuts.Serialize())
						Handle(err)
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			key := append(utxoPrefix, tx.ID...)
			err := txn.Set(key, newOutputs.Serialize())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

// CountTransactions function to count the transactions in the UTXOs
func (u *UTXOSet) CountTransactions() int {
	counter := 0
	err := u.Blockchain.Database.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := tx.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}
		return nil
	})
	Handle(err)
	return counter
}

// FindUnspentTransactionOutputs function to find all the unspent transaction outputs in the blockchain
func (u *UTXOSet) FindUnspentTransactionOutputs(publicKeyHash []byte) []TxOutput {
	var UnspentTransactionsOutputs []TxOutput

	//
	err := u.Blockchain.Database.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := tx.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			Handle(err)

			k = k[prefixLength:]
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					UnspentTransactionsOutputs = append(UnspentTransactionsOutputs, out)
				}
			}
		}
		return nil
	})
	// Handle the error
	Handle(err)

	// Return the unspent transactions outputs
	return UnspentTransactionsOutputs
}

// FindSpendableOutputs function to find the spendable outputs in the blockchain
func (u *UTXOSet) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	// Initialize variables
	var (
		unspentOutputs = make(map[string][]int)
		accumulated    = 0
	)

	err := u.Blockchain.Database.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := tx.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			Handle(err)

			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})

	Handle(err)

	return accumulated, unspentOutputs
}
