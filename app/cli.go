package app

import (
	"fmt"
	"os"
	"github.com/btcsuite/btcutil/base58"
	"bytes"
	"strconv"
)

type CLI struct {}

const Usage=`
createw "create wallet"
laddress "list address"
createbc --address ADDRESS "create blockchain"
balance --address ADDRESS "get balance"
send --from FROM_ADDRESS --to TO_ADDRESS --amount AMOUNT --miner MINER_ADDRESS "transfer"
print "print blockchain"
`

func PrintError(e interface{})  {
	fmt.Println(e)
	os.Exit(-1)
}

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
				PrintError("address cannot be empty")
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
				PrintError("address cannot be empty")
			}
			cli.Balance(address)
		} else {
			PrintError(Usage)
		}
	case "print":
		cli.Print()
	default:
		PrintError(Usage)
	}
}

func (cli *CLI) CreateWallet()  {
	wallets:=GetWallets()
	address:=wallets.CreateWallet()
	fmt.Println(address)
}

func (cli *CLI) ListAddress()  {
	wallets:=GetWallets()
	addressList:=wallets.ListAddress()
	for _,address:=range addressList{
		fmt.Println(address)
	}
}

func (cli *CLI) CreateBlockChain(address string)  {
	if !isValidAddress(address) {
		PrintError("invalid address")
	}
	bc:=NewBlockChain(address)
	defer bc.DB.Close()
}

func (cli *CLI) Send(from string,to string,amount float64,miner string)  {
	if !isValidAddress(from) {
		PrintError("invalid from address")
	}
	if !isValidAddress(to) {
		PrintError("invalid to address")
	}
	if !isValidAddress(miner) {
		PrintError("invalid miner address")
	}
	bc:=GetBlockChain()
	defer bc.DB.Close()
	coinbase:=NewCoinbase(miner,"")
	tx:=bc.NewTransaction(from,to,amount)
	bc.AddBlock([]*Transaction{coinbase,tx})
}

func (cli *CLI) Balance(address string)  {
	if !isValidAddress(address) {
		PrintError("invalid address")
	}
	bc:=GetBlockChain()
	defer bc.DB.Close()
	x:=base58.Decode(address)
	h:=x[1:len(x)-4]
	txos:=bc.FindUTXOs(h)
	total:=0.0
	for _,txo:=range txos{
		total+=txo.Value
	}
	fmt.Println(total)
}

func (cli *CLI) Print()  {
	bc:=GetBlockChain()
	defer bc.DB.Close()
	bci:=bc.NewBlockChainIterator()
	bci.Travel(func(block *Block) {
		fmt.Println(block)
		for _,tx:=range block.Transactions{
			fmt.Println(tx)
		}
	})
}

func isValidAddress(address string) bool {
	x:=base58.Decode(address)
	if len(x)<4 {
		return false
	}
	payload:=x[:len(x)-4]
	checkcode1:=x[len(x)-4:]
	checkcode2:=CheckSum(payload)
	return bytes.Compare(checkcode1,checkcode2)==0
}