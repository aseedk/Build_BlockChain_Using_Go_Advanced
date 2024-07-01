package wallet

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "./tmp/wallets.data"

// Wallets struct which contains the wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// CreateWallets function to create a new Wallets struct
func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile()

	return &wallets, err
}

// AddWallet function to add a wallet to the Wallets struct
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := string(wallet.Address())
	ws.Wallets[address] = wallet
	return address
}

// GetAllAddresses function to get all the addresses of the wallets
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet function to get a wallet by its address
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// SaveFile function to save the wallets to a file
func (ws *Wallets) SaveFile() {
	jsonData, err := json.Marshal(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, jsonData, 0644)
	if err != nil {
		log.Panic(err)
	}
}

// LoadFile function to load the wallets from a file
func (ws *Wallets) LoadFile() error {
	if _, err := ioutil.ReadFile(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets

	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// wallet.go
