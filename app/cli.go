package app

import (
	"fmt"
	"os"
	"strconv"
	"github.com/btcsuite/btcutil/base58"
		"bytes"
)

type CLI struct {}

const Usage=`
	Usage:
	createw "create a wallet"
	laddress "list addresses"
	createbc --address ADDRESS "create a blockchain"
	send --from FROM_ADDRESS --to TO_ADDRESS --amount AMOUNT --miner MINER_ADDRESS
	balance --address ADDRESS "get balance"
	print "print blockchain"
`

func (cli *CLI) Run()  {
	if len(os.Args)<2 {
		PrintError(Usage)
	}
	switch os.Args[1] {
	case "createw":
		cli.CreateWallet()
	case "laddress":
		cli.ListAddress()
	case "createbc":
		if len(os.Args)>3 && os.Args[2]=="--address" {
			address:=os.Args[3]
			if address=="" {
				PrintError("invalid address")
			}
			cli.CreateBlockChain(address)
		} else {
			PrintError(Usage)
		}
	case "send":
		if len(os.Args)<10 {
			PrintError(Usage)
		}
		from:=os.Args[3]
		to:=os.Args[5]
		amount,_:=strconv.ParseFloat(os.Args[7],64)
		miner:=os.Args[9]
		cli.Send(from,to,amount,miner)
	case "balance":
		if len(os.Args)>3 && os.Args[2]=="--address" {
			address:=os.Args[3]
			if address=="" {
				PrintError("invalid address")
			}
			cli.GetBalance(address)
		} else {
			PrintError(Usage)
		}
	case "print":
		cli.PrintBlockChain()
	default:
		PrintError(Usage)
	}
}

func (cli *CLI) CreateWallet()  {
	wallets:=GetWallets()
	address:=wallets.CreateWallet()
	fmt.Println("address ",address)
}

func (cli *CLI) ListAddress()  {
	wallets:=GetWallets()
	addressList:=wallets.ListAddress()
	for _,address:=range addressList{
		fmt.Println(address)
	}
}

func (cli *CLI) CreateBlockChain(address string)  {
	if !IsValidAddress(address) {
		PrintError("invalid address")
	}
	bc:=NewBlockChain(address)
	defer bc.db.Close()
}

func (cli *CLI) Send(from string,to string,amount float64,miner string)  {
	if !IsValidAddress(from) || !IsValidAddress(to) || !IsValidAddress(miner) {
		PrintError("invalid address")
	}
	bc:=GetBlockChain()
	defer bc.db.Close()
	coinbase:=NewCoinbase(miner,"")
	tx:=bc.NewTransaction(from,to,amount)
	bc.AddBlock([]*Transaction{coinbase,tx})
}

func (cli *CLI) GetBalance(address string)  {
	if !IsValidAddress(address) {
		PrintError("invalid address")
	}
	bc:=GetBlockChain()
	defer bc.db.Close()
	x:=base58.Decode(address)
	h:=x[1:len(x)-4]
	txos:=bc.FindUTXOs(h)
	balance:=0.0
	for _,txo:=range txos{
		balance+=txo.Value
	}
	fmt.Println(address," balance ",balance)
}

func (cli *CLI) PrintBlockChain()  {
	bc:=GetBlockChain()
	defer bc.db.Close()
	bci:=bc.NewBlockChainIterator()
	for  {
		block:=bci.GetCurrentBlock()
		fmt.Println(block)
		for _,tx:=range block.Transactions{
			fmt.Println(tx)
		}
		if len(block.PrevBlockHash)==0 {
			break
		}
	}
}

func IsValidAddress(address string) bool {
	x:=base58.Decode(address)
	if len(x)<4 {
		return false
	}
	payload:=x[:len(x)-4]
	checkcode1:=x[len(x)-4:]
	checkcode2:=CheckSum(payload)
	return bytes.Compare(checkcode1,checkcode2)==0
}

func PrintError(e interface{})  {
	fmt.Println(e)
	os.Exit(-1)
}