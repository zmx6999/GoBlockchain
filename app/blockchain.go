package app

import (
	"bolt"
	"os"
	"errors"
	"crypto/ecdsa"
	"bytes"
)

type BlockChain struct {
	db *bolt.DB
	tail []byte
}

const (
	DBName="181125.db"
	BucketName="181125"
	LastKey="last"
	GenesisInfo="The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

func NewBlockChain(address string) *BlockChain {
	if IsDBExists() {
		PrintError(DBName+" exists")
	}
	db,err:=bolt.Open(DBName,0600,nil)
	if err!=nil {
		PrintError(err)
	}
	var tail []byte
	db.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			bucket,err=tx.CreateBucket([]byte(BucketName))
			if err!=nil {
				return err
			}
			coinbase:=NewCoinbase(address,GenesisInfo)
			block:=NewBlock([]*Transaction{coinbase},[]byte{})
			bucket.Put(block.Hash,Serialize(block))
			bucket.Put([]byte(LastKey),block.Hash)
			tail=block.Hash
			return nil
		} else {
			return errors.New("bucket "+BucketName+" exists")
		}
	})
	return &BlockChain{db,tail}
}

func GetBlockChain() *BlockChain {
	if !IsDBExists() {
		PrintError(DBName+" not exists")
	}
	db,err:=bolt.Open(DBName,0600,nil)
	if err!=nil {
		PrintError(err)
	}
	var tail []byte
	db.View(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			return errors.New("bucket "+BucketName+" not exists")
		} else {
			tail=bucket.Get([]byte(LastKey))
			return nil
		}
	})
	return &BlockChain{db,tail}
}

func (bc *BlockChain) AddBlock(transactions []*Transaction)  {
	for _,tx:=range transactions{
		if !bc.VerifyTransaction(tx) {
			PrintError("invalid transaction")
		}
	}
	bc.db.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			return errors.New("bucket "+BucketName+" not exists")
		} else {
			block:=NewBlock(transactions,bc.tail)
			bucket.Put(block.Hash,Serialize(block))
			bucket.Put([]byte(LastKey),block.Hash)
			bc.tail=block.Hash
			return nil
		}
	})
}

func (bc *BlockChain) NewBlockChainIterator() *BlockChainIterator {
	return &BlockChainIterator{bc.db,bc.tail}
}

func IsDBExists() bool {
	if _,err:=os.Stat(DBName);os.IsNotExist(err) {
		return false
	}
	return true
}

func (bc *BlockChain) FindUTXOTransactions(publicKeyHash []byte) []*Transaction {
	var txs []*Transaction
	bci:=bc.NewBlockChainIterator()
	spentUTXO:=make(map[string][]int64)
	for  {
		block:=bci.GetCurrentBlock()
		for _,tx:=range block.Transactions{
			OUTPUT:
			for k,txo:=range tx.TXOutputs{
				indexs:=spentUTXO[string(tx.TXID)]
				if indexs!=nil {
					for _,index:=range indexs{
						if int64(k)==index {
							continue OUTPUT
						}
					}
				}
				if txo.CanBeUnlockedWith(publicKeyHash) {
					txs=append(txs,tx)
				}
			}
			if !tx.IsCoinbase() {
				for _,txi:=range tx.TXInputs{
					if txi.CanUnlockWith(publicKeyHash) {
						spentUTXO[string(txi.TXID)]=append(spentUTXO[string(txi.TXID)],txi.OutputIndex)
					}
				}
			}
		}
		if len(block.PrevBlockHash)==0 {
			break
		}
	}
	return txs
}

func (bc *BlockChain) FindUTXOs(publicKeyHash []byte) []TXOutput {
	var txos []TXOutput
	txs:=bc.FindUTXOTransactions(publicKeyHash)
	for _,tx:=range txs{
		for _,txo:=range tx.TXOutputs{
			if txo.CanBeUnlockedWith(publicKeyHash) {
				txos=append(txos,txo)
			}
		}
	}
	return txos
}

func (bc *BlockChain) FindSuitableUTXOs(publicKeyHash []byte,amount float64) (float64,map[string][]int64) {
	utxos:=make(map[string][]int64)
	txs:=bc.FindUTXOTransactions(publicKeyHash)
	total:=0.0
	FIND:
	for _,tx:=range txs{
		for k,txo:=range tx.TXOutputs{
			if txo.CanBeUnlockedWith(publicKeyHash) {
				if total<amount {
					utxos[string(tx.TXID)]=append(utxos[string(tx.TXID)],int64(k))
					total+=txo.Value
				} else {
					break FIND
				}
			}
		}
	}
	return total,utxos
}

func (bc *BlockChain) NewTransaction(from string,to string,amount float64) *Transaction {
	wallets:=GetWallets()
	wallet:=wallets.WalletMap[from]
	publicKey:=wallet.PublicKey
	publicKeyHash:=HashPublicKey(publicKey)
	total,utxos:=bc.FindSuitableUTXOs(publicKeyHash,amount)
	if total<amount {
		PrintError("insufficient balance")
	}
	var txis []TXInput
	for k,v:=range utxos{
		for _,v2:=range v{
			txis=append(txis,TXInput{[]byte(k),v2,nil,publicKey})
		}
	}
	var txos []TXOutput
	txos=append(txos,NewTXOutput(amount,to))
	if total>amount {
		txos=append(txos,NewTXOutput(total-amount,from))
	}
	tx:=Transaction{nil,txis,txos}
	tx.GetTXID()
	privateKey:=wallet.PrivateKey
	bc.SignTransaction(&tx,&privateKey)
	return &tx
}

func (bc *BlockChain) SignTransaction(transaction *Transaction,privateKey *ecdsa.PrivateKey)  {
	prevTxs:=make(map[string]*Transaction)
	for _,txi:=range transaction.TXInputs{
		prevTx:=prevTxs[string(txi.TXID)]
		if prevTx==nil {
			prevTx=bc.FindTransaction(txi.TXID)
			prevTxs[string(txi.TXID)]=prevTx
		}
	}
	transaction.Sign(privateKey,prevTxs)
}

func (bc *BlockChain) VerifyTransaction(transaction *Transaction) bool {
	if transaction.IsCoinbase() {
		return true
	}
	prevTxs:=make(map[string]*Transaction)
	for _,txi:=range transaction.TXInputs{
		prevTx:=prevTxs[string(txi.TXID)]
		if prevTx==nil {
			prevTx=bc.FindTransaction(txi.TXID)
			prevTxs[string(txi.TXID)]=prevTx
		}
	}
	return transaction.Verify(prevTxs)
}

func (bc *BlockChain) FindTransaction(txid []byte) *Transaction {
	bci:=bc.NewBlockChainIterator()
	for  {
		block:=bci.GetCurrentBlock()
		for _,tx:=range block.Transactions{
			if bytes.Compare(tx.TXID,txid)==0 {
				return tx
			}
		}
		if len(block.PrevBlockHash)==0 {
			break
		}
	}
	return nil
}