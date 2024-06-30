package main

import (
	"Build_BlockChain_Using_Go_Advanced/src"
	"fmt"
	"strconv"
)

func main() {
	blockChain := src.InitBlockChain()

	blockChain.AddBlock("First Block after Genesis")
	blockChain.AddBlock("Second Block after Genesis")
	blockChain.AddBlock("Third Block after Genesis")

	for _, block := range blockChain.Blocks {
		fmt.Println("------------------------------------------------------------------")
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := src.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("------------------------------------------------------------------")
		fmt.Println()
	}
}
