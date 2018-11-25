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

const WalletName="181125.wat"

func GetWallets() *Wallets {
	wallets:=LoadWalletsFromFile()
	if wallets==nil {
		wallets=&Wallets{make(map[string]*Wallet)}
	}
	return wallets
}

func (ws *Wallets) CreateWallet() string {
	wallets:=GetWallets()
	wallet:=NewWallet()
	address:=wallet.GetAddress()
	wallets.WalletMap[address]=wallet
	wallets.SaveFile()
	return address
}

func LoadWalletsFromFile() *Wallets {
	if _,err:=os.Stat(WalletName);os.IsNotExist(err) {
		return nil
	}
	data,err:=ioutil.ReadFile(WalletName)
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

func (ws *Wallets) SaveFile()  {
	var buffer bytes.Buffer
	encoder:=gob.NewEncoder(&buffer)
	gob.Register(elliptic.P256())
	err:=encoder.Encode(ws)
	if err!=nil {
		PrintError(err)
	}
	err=ioutil.WriteFile(WalletName,buffer.Bytes(),0600)
}

func (ws *Wallets) ListAddress() []string {
	var addressList []string
	for address:=range ws.WalletMap{
		addressList=append(addressList,address)
	}
	return addressList
}