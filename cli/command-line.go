package cli

import (
	"Build_BlockChain_Using_Go_Advanced/blockchain"
	"Build_BlockChain_Using_Go_Advanced/network"
	"Build_BlockChain_Using_Go_Advanced/wallet"
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
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM to TO")
	fmt.Println("createwallet - Create a new wallet")
	fmt.Println("listaddresses - List the addresses in our wallet file")
	fmt.Println("reindexutxo - Rebuild the UTXO set")
	fmt.Println("startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. - Enable mining and send reward to ADDRESS")
}

// ValidateArgs function to validate the arguments provided by the user
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

// GetBalance function to get the balance of an address
func (cli *CommandLine) GetBalance(address, nodeId string) {
	if !wallet.ValidateAddress(address) {
		fmt.Println("Address is not valid")
		runtime.Goexit()
	}

	// Initialize the blockchain
	blockchain := src.ContinueBlockChain(nodeId)
	UTXOSet := src.UTXOSet{Blockchain: blockchain}

	// Defer closing the database
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)

	// Initialize the balance
	balance := 0

	// Get the public key hash from the address
	publicKeyHash := wallet.Base58Decode([]byte(address))
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-4]

	// Find the unspent transaction outputs for the address
	UTXOs := UTXOSet.FindUnspentTransactionOutputs(publicKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)

}

// CreateBlockChain function to create a new blockchain
func (cli *CommandLine) CreateBlockChain(address string, nodeId string) {
	if !wallet.ValidateAddress(address) {
		fmt.Println("Address is not valid")
		runtime.Goexit()
	}

	blockchain := src.InitBlockChain(address, nodeId)
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)

	UTXOSet := src.UTXOSet{Blockchain: blockchain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")

}

// PrintChain function to print the blocks in the blockchain
func (cli *CommandLine) PrintChain(nodeId string) {
	blockchain := src.ContinueBlockChain(nodeId)
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

		pow := src.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println("------------------------------------------------------------------")
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// Send function to send coins from one address to another
func (cli *CommandLine) Send(from, to string, amount int, nodeId string, mineNow bool) {
	if !wallet.ValidateAddress(from) {
		fmt.Println("Address is not valid")
		runtime.Goexit()
	}

	if !wallet.ValidateAddress(to) {
		fmt.Println("Address is not valid")
		runtime.Goexit()
	}

	blockchain := src.ContinueBlockChain(nodeId)
	UTXOSet := src.UTXOSet{Blockchain: blockchain}

	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)
	wallets, err := wallet.CreateWallets(nodeId)

	if err != nil {
		log.Panic(err)
	}

	w := wallets.GetWallet(from)

	tx := src.NewTransaction(&w, to, amount, &UTXOSet)

	if mineNow {
		cbTx := src.CoinbaseTransaction(from, "")
		block := blockchain.MineBlock([]*src.Transaction{cbTx, tx})
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("Sent transaction")
	}

	fmt.Println("Success!")
}

// CreateWallet function to create a new wallet
func (cli *CommandLine) CreateWallet(nodeId string) {
	wallets, _ := wallet.CreateWallets(nodeId)
	address := wallets.AddWallet()
	wallets.SaveFile(nodeId)

	fmt.Printf("New address is: %s\n", address)

}

// ListAddresses function to list the addresses in the wallet file
func (cli *CommandLine) ListAddresses(nodeId string) {
	wallets, err := wallet.CreateWallets(nodeId)
	src.Handle(err)

	addresses := wallets.GetAllAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

// ReindexUTXO function to reindex the UTXO set
func (cli *CommandLine) ReindexUTXO(nodeId string) {
	blockchain := src.ContinueBlockChain(nodeId)
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockchain.Database)
	UTXOSet := src.UTXOSet{Blockchain: blockchain}
	UTXOSet.Reindex()

	transactions := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", transactions)
}

// StartNode function to start a node
func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	network.StartServer(nodeID, minerAddress)
}

// Run function to run the command line interface
func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	nodeId := os.Getenv("NODE_ID")
	if nodeId == "" {
		fmt.Printf("NODE_ID env is not set!")
		runtime.Goexit()
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockChainAddress := createBlockChainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "The address to send coins from")
	sendTo := sendCmd.String("to", "", "The address to send coins to")
	sendAmount := sendCmd.Int("amount", 0, "The amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

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
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
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
		cli.GetBalance(*getBalanceAddress, nodeId)
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			createBlockChainCmd.Usage()
			runtime.Goexit()
		}
		cli.CreateBlockChain(*createBlockChainAddress, nodeId)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain(nodeId)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.Send(*sendFrom, *sendTo, *sendAmount, nodeId, *sendMine)
	}

	if createWalletCmd.Parsed() {
		cli.CreateWallet(nodeId)
	}

	if listAddressesCmd.Parsed() {
		cli.ListAddresses(nodeId)
	}

	if reindexUTXOCmd.Parsed() {
		cli.ReindexUTXO(nodeId)
	}

	if startNodeCmd.Parsed() {
		cli.StartNode(nodeId, *startNodeMiner)
	}

	fmt.Println()
}
