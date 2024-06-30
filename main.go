package main

import (
	"Build_BlockChain_Using_Go_Advanced/src"
	"fmt"
	"github.com/dgraph-io/badger"
	"os"
)

func main() {
	defer os.Exit(0)
	blockChain := src.InitBlockChain()
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(blockChain.Database)

	commandLine := src.CommandLine{Blockchain: blockChain}
	commandLine.Run()
}
