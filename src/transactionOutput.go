package src

// TxOutput struct which contains the value and the public key of the input
type TxOutput struct {
	Value     int    // Value of the output
	PublicKey string // Public key of the output
}

// CanBeUnlocked function to check if the transaction can be unlocked with the provided data
func (out *TxOutput) CanBeUnlocked(data string) bool {
	// Check if the public key of the output is equal to the data
	return out.PublicKey == data
}
