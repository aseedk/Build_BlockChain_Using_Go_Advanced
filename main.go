package main

import (
	"Build_BlockChain_Using_Go_Advanced/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	commandLine := cli.CommandLine{}
	commandLine.Run()
}
