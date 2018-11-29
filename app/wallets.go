package app

import (
	"os"
	"io/ioutil"
	"bytes"
	"encoding/gob"
	"crypto/elliptic"
)

type Wallets struct {
	WalletMap map[string]*Wallet
}

const WalletFile="181129.wat"

func GetWallets() *Wallets {
	wallets:=LoadWalletsFromFile()
	if wallets==nil {
		wallets=&Wallets{make(map[string]*Wallet)}
	}
	return wallets
}

func (ws *Wallets) CreateWallet() string {
	wallet:=NewWallet()
	address:=wallet.GetAddress()
	ws.WalletMap[address]=wallet
	ws.SaveFile()
	return address
}

func (ws *Wallets) SaveFile()  {
	var buffer bytes.Buffer
	encoder:=gob.NewEncoder(&buffer)
	gob.Register(elliptic.P256())
	err:=encoder.Encode(ws)
	if err!=nil {
		PrintError(err)
	}
	err=ioutil.WriteFile(WalletFile,buffer.Bytes(),0600)
	if err!=nil {
		PrintError(err)
	}
}

func LoadWalletsFromFile() *Wallets {
	if _,err:=os.Stat(WalletFile);os.IsNotExist(err) {
		return nil
	}
	data,err:=ioutil.ReadFile(WalletFile)
	if err!=nil {
		PrintError(err)
	}
	var buffer bytes.Buffer
	buffer.Write(data)
	decoder:=gob.NewDecoder(&buffer)
	gob.Register(elliptic.P256())
	var wallets Wallets
	err=decoder.Decode(&wallets)
	if err!=nil {
		PrintError(err)
	}
	return &wallets
}

func (ws *Wallets) ListAddress() []string {
	var addressList []string
	for address:=range ws.WalletMap{
		addressList=append(addressList,address)
	}
	return addressList
}