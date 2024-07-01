package src

// TxInput struct which contains the ID of the transaction, the output and the signature
type TxInput struct {
	ID        []byte // ID of the transaction
	Out       int    // Output of the transaction
	Signature string // Signature of the transaction
}

// CanUnlock function to check if the transaction can be unlocked with the provided data
func (in *TxInput) CanUnlock(data string) bool {
	// Check if the signature of the input is equal to the data
	return in.Signature == data
}
