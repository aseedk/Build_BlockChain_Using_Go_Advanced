package src

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

// CommandLine interface to interact with the blockchain
type CommandLine struct {
	Blockchain *BlockChain
}

// PrintUsage function to print the usage of the command line
func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the blockchain")
	fmt.Println(" print - prints the blocks in the blockchain")
}

// ValidateArgs function to validate the arguments provided by the user
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

// AddBlock function to add a block to the blockchain
func (cli *CommandLine) AddBlock(data string) {
	cli.Blockchain.AddBlock(data)
	fmt.Println("Block added!")
}

// PrintChain function to print the blocks in the blockchain
func (cli *CommandLine) PrintChain() {
	iterator := cli.Blockchain.Iterator()

	for {
		block := iterator.Next()
		fmt.Println("------------------------------------------------------------------")
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("------------------------------------------------------------------")
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// Run function to run the command line interface
func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)

	addBlockData := addBlockCmd.String("block", "", "The data of the block")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		Handle(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		Handle(err)
	default:
		cli.PrintUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.AddBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	fmt.Println()
}
