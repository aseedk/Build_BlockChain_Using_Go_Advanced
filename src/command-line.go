package src

import (
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"runtime"
	"strconv"
)

// CommandLine interface to interact with the blockchain
type CommandLine struct{}

// PrintUsage function to print the usage of the command line
func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println(" createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println(" printchain - Print all the blocks of the blockchain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM to TO")
}

// ValidateArgs function to validate the arguments provided by the user
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

// GetBalance function to get the balance of an address
func (cli *CommandLine) GetBalance(address string) {
	// Initialize the blockchain
	blockchain := ContinueBlockChain(address)

	// Defer closing the database
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)

	// Initialize the balance
	balance := 0

	// Find the unspent transaction outputs for the address
	UTXOs := blockchain.FindUnspentTransactionOutputs(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)

}

// CreateBlockChain function to create a new blockchain
func (cli *CommandLine) CreateBlockChain(address string) {
	blockchain := InitBlockChain(address)
	err := blockchain.Database.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Finished!")

}

// PrintChain function to print the blocks in the blockchain
func (cli *CommandLine) PrintChain() {
	blockchain := ContinueBlockChain("")
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)

	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()
		fmt.Println("------------------------------------------------------------------")
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
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

// Send function to send coins from one address to another
func (cli *CommandLine) Send(from, to string, amount int) {
	blockchain := ContinueBlockChain(from)
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)

	tx := NewTransaction(from, to, amount, blockchain)
	blockchain.AddBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

// Run function to run the command line interface
func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockChainAddress := createBlockChainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "The address to send coins from")
	sendTo := sendCmd.String("to", "", "The address to send coins to")
	sendAmount := sendCmd.Int("amount", 0, "The amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.GetBalance(*getBalanceAddress)
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			createBlockChainCmd.Usage()
			runtime.Goexit()
		}
		cli.CreateBlockChain(*createBlockChainAddress)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.Send(*sendFrom, *sendTo, *sendAmount)
	}

	fmt.Println()
}
