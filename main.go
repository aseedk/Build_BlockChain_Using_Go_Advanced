package main

import (
	"Build_BlockChain_Using_Go_Advanced/src"
	"os"
)

func main() {
	defer os.Exit(0)
	cli := src.CommandLine{}
	cli.Run()
}
