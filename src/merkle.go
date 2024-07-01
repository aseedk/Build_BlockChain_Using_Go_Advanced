package src

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleNode function to create a new Merkle node
func NewMerkleNode(left *MerkleNode, right *MerkleNode, data []byte) *MerkleNode {
	// Create a new Merkle node with the left and right nodes and the data
	merkleNode := MerkleNode{}

	// Check if the left node is nil
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		merkleNode.Data = hash[:]
	} else {
		// Set the data of the Merkle node
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		merkleNode.Data = hash[:]
	}

	// Set the left and right nodes of the Merkle node
	merkleNode.Left = left
	merkleNode.Right = right

	// Return the Merkle node
	return &merkleNode
}

// NewMerkleTree function to create a new Merkle tree
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// Check if the number of data is odd
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	// Loop through the data
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	// Loop through the nodes
	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		// Loop through the nodes
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		// Set the new level
		nodes = newLevel
	}

	// Create a new Merkle tree with the root node
	merkleTree := MerkleTree{&nodes[0]}

	// Return the Merkle tree
	return &merkleTree
}
