package network

import (
	"Build_BlockChain_Using_Go_Advanced/blockchain"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/vrecan/death/v3"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string
	minerAddress    string
	knownNodes      = []string{"localhost:3000"}
	blocksInTransit [][]byte
	memoryPool      = make(map[string]src.Transaction)
)

// Address struct which contains the address of the node
type Address struct {
	AddressList []string
}

// Block struct which contains the block of the blockchain
type Block struct {
	AddressFrom string
	Block       []byte
}

// GetBlocks struct which contains the command to get the blocks
type GetBlocks struct {
	AddressFrom string
}

// GetData struct which contains the command to get the data
type GetData struct {
	AddressFrom string
	Type        string
	ID          []byte
}

// Inventory struct which contains the inventory of the blockchain
type Inventory struct {
	AddressFrom string
	Type        string
	Items       [][]byte
}

// Tx struct which contains the transaction of the blockchain
type Tx struct {
	AddressFrom string
	Transaction []byte
}

// Version struct which contains the version of the blockchain
type Version struct {
	Version     int
	BestHeight  int
	AddressFrom string
}

// StartServer function to start the server
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)

	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer func(ln net.Listener) {
		err = ln.Close()
		if err != nil {
			log.Panic(err)
		}
	}(ln)

	chain := src.ContinueBlockChain(nodeID)
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			log.Panic(err)
		}
	}(chain.Database)

	go CloseDB(chain)

	if nodeAddress != knownNodes[0] {
		SendVersion(knownNodes[0], chain)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)
	}
}

// HandleConnection function to handle the connection
func HandleConnection(conn net.Conn, chain *src.BlockChain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Panic(err)
		}
	}(conn)

	command := BytesToCommand(req[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		HandleAddress(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInventory(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		fmt.Println("Unknown command!")
	}
}

// SendAddress function to send the address
func SendAddress(address string) {
	nodes := Address{knownNodes}
	nodes.AddressList = append(nodes.AddressList, nodeAddress)
	payload := GobEncode(nodes)
	request := append(CommandToBytes("addr"), payload...)
	SendData(address, request)
}

// SendBlock function to send the block
func SendBlock(address string, block []byte) {
	data := Block{nodeAddress, block}
	payload := GobEncode(data)
	request := append(CommandToBytes("block"), payload...)
	SendData(address, request)
}

// SendInventory function to send the inventory
func SendInventory(address, kind string, items [][]byte) {
	data := Inventory{nodeAddress, kind, items}
	payload := GobEncode(data)
	request := append(CommandToBytes("inv"), payload...)
	SendData(address, request)
}

// SendTx function to send the transaction
func SendTx(address string, tnx *src.Transaction) {
	data := Tx{nodeAddress, tnx.Serialize()}
	payload := GobEncode(data)
	request := append(CommandToBytes("tx"), payload...)
	SendData(address, request)
}

// SendVersion function to send the version
func SendVersion(addr string, chain *src.BlockChain) {
	bestHeight := chain.GetBestHeight()
	payload := GobEncode(Version{version, bestHeight, nodeAddress})
	request := append(CommandToBytes("version"), payload...)
	SendData(addr, request)
}

// SendGetData function to send the data
func SendGetData(address, kind string, id []byte) {
	payload := GobEncode(GetData{nodeAddress, kind, id})
	request := append(CommandToBytes("getdata"), payload...)
	SendData(address, request)
}

// HandleAddress function to handle the address
func HandleAddress(request []byte) {
	var buff bytes.Buffer
	var payload Address

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	knownNodes = append(knownNodes, payload.AddressList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	RequestBlocks()
}

// HandleBlock function to handle the block
func HandleBlock(request []byte, chain *src.BlockChain) {
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	block := src.Deserialize(blockData)

	fmt.Println("Received a new block!")
	chain.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddressFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := src.UTXOSet{Blockchain: chain}
		UTXOSet.Reindex()
	}
}

// HandleGetBlocks function to handle the get blocks
func HandleGetBlocks(request []byte, chain *src.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlocks

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := chain.GetBlockHashes()
	SendInventory(payload.AddressFrom, "block", blocks)
}

// HandleGetData function to handle the data
func HandleGetData(request []byte, chain *src.BlockChain) {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		SendBlock(payload.AddressFrom, &block)
	}
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]
		SendTx(payload.AddressFrom, &tx)
	}
}

// HandleVersion function to handle the version
func HandleVersion(request []byte, chain *src.BlockChain) {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	bestHeight := chain.GetBestHeight()
	otherHeight := payload.BestHeight

	if bestHeight < otherHeight {
		SendGetBlocks(payload.AddressFrom)
	} else if bestHeight > otherHeight {
		SendVersion(payload.AddressFrom, chain)
	}

	if !NodeIsKnown(payload.AddressFrom) {
		knownNodes = append(knownNodes, payload.AddressFrom)
	}
}

// HandleTx function to handle the transaction
func HandleTx(request []byte, chain *src.BlockChain) {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	txData := payload.Transaction
	tx := src.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddressFrom {
				SendInventory(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(memoryPool) >= 2 && len(minerAddress) > 0 {
			MineTransactions(chain)
		}
	}
}

// HandleInventory function to handle the inventory
func HandleInventory(request []byte, chain *src.BlockChain) {
	var buff bytes.Buffer
	var payload Inventory

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0]
		SendGetData(payload.AddressFrom, "block", blockHash)

		var newInTransit [][]byte
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]
		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddressFrom, "tx", txID)
		}
	}
}

// MineTransactions function to mine the transactions
func MineTransactions(chain *src.BlockChain) {
	var transactions []*src.Transaction

	for id := range memoryPool {
		fmt.Printf("ID: %x\n", id)
		tx := memoryPool[id]
		if chain.VerifyTransaction(&tx) {
			transactions = append(transactions, &tx)
		}
	}

	if len(transactions) == 0 {
		fmt.Println("All transactions are invalid!")
		return
	}

	cbTx := src.CoinbaseTransaction(minerAddress, "")
	transactions = append(transactions, cbTx)

	newBlock := chain.MineBlock(transactions)
	UTXOSet := src.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	fmt.Println("New block is mined!")

	for _, tx := range transactions {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range knownNodes {
		if node != nodeAddress {
			SendInventory(node, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(memoryPool) > 0 {
		MineTransactions(chain)
	}
}

// NodeIsKnown function to check if the node is known
func NodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}

// SendData function to send the data
func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updateNodes []string
		for _, node := range knownNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knownNodes = updateNodes
	}

	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Panic(err)
		}
	}(conn)

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

// CommandToBytes function to convert the command to bytes
func CommandToBytes(command string) []byte {
	var b [commandLength]byte
	for i, c := range command {
		b[i] = byte(c)
	}
	return b[:]
}

// BytesToCommand function to convert the bytes to command
func BytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

// GobEncode function to encode the data
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// RequestBlocks function to request the blocks
func RequestBlocks() {
	for _, node := range knownNodes {
		SendGetBlocks(node)
	}
}

// CloseDB function to close the database
func CloseDB(chain *src.BlockChain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		err := chain.Database.Close()
		if err != nil {
			log.Panic(err)
		}
	})
}
