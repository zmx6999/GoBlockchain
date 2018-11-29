package app

import (
	"bolt"
	"os"
	"errors"
	"bytes"
	"crypto/ecdsa"
)

type BlockChain struct {
	DB *bolt.DB
	Tail []byte
}

const (
	DBName="181129.db"
	BucketName="181129"
	LastKey="last"
)

func NewBlockChain(address string) *BlockChain {
	if DBExists() {
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
			coinbase:=NewCoinbase(address,"Chancellor on brink of second bailout")
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
	if !DBExists() {
		PrintError(DBName+" doesn't exist")
	}
	db,err:=bolt.Open(DBName,0600,nil)
	if err!=nil {
		PrintError(err)
	}
	var tail []byte
	db.View(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			return errors.New("bucket "+BucketName+" doesn't exist")
		} else {
			tail=bucket.Get([]byte(LastKey))
			return nil
		}
	})
	return &BlockChain{db,tail}
}

func DBExists() bool {
	if _,err:=os.Stat(DBName);os.IsNotExist(err) {
		return false
	}
	return true
}

func (bc *BlockChain) NewBlockChainIterator() *BlockChainIterator {
	return &BlockChainIterator{bc.DB,bc.Tail}
}

func (bc *BlockChain) FindUTXOTransactions(publicKeyHash []byte) []*Transaction {
	var txs []*Transaction
	spentUTXO:=make(map[string][]int64)
	bci:=bc.NewBlockChainIterator()
	bci.Travel(func(block *Block) {
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
						spentUTXO[string(txi.TXID)]=append(spentUTXO[string(txi.TXID)],txi.OutIndex)
					}
				}
			}
		}
	})
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
	txos:=[]TXOutput{NewTXOutput(amount,to)}
	if total>amount {
		txos=append(txos,NewTXOutput(total-amount,from))
	}
	tx:=Transaction{nil,txis,txos}
	tx.GetTXID()
	bc.SignTransaction(&tx,&wallet.PrivateKey)
	return &tx
}

func (bc *BlockChain) AddBlock(txs []*Transaction)  {
	for _,tx:=range txs{
		if !bc.VerifyTransaction(tx) {
			PrintError("invalid transaction")
		}
	}
	bc.DB.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			return errors.New("bucket "+BucketName+" doesn't exist")
		} else {
			block:=NewBlock(txs,bc.Tail)
			bucket.Put(block.Hash,Serialize(block))
			bucket.Put([]byte(LastKey),block.Hash)
			bc.Tail=block.Hash
			return nil
		}
	})
}

func (bc *BlockChain) FindTransaction(id []byte) *Transaction {
	var tx *Transaction
	bci:=bc.NewBlockChainIterator()
	bci.Travel(func(block *Block) {
		for _,_tx:=range block.Transactions{
			if bytes.Compare(id,_tx.TXID)==0 {
				tx=_tx
				return
			}
		}
	})
	return tx
}

func (bc *BlockChain) SignTransaction(tx *Transaction,privateKey *ecdsa.PrivateKey)  {
	prevTxs:=make(map[string]*Transaction)
	for _,txi:=range tx.TXInputs{
		if prevTxs[string(txi.TXID)]==nil {
			prevTx:=bc.FindTransaction(txi.TXID)
			prevTxs[string(txi.TXID)]=prevTx
		}
	}
	tx.Sign(privateKey,prevTxs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTxs:=make(map[string]*Transaction)
	for _,txi:=range tx.TXInputs{
		if prevTxs[string(txi.TXID)]==nil {
			prevTx:=bc.FindTransaction(txi.TXID)
			prevTxs[string(txi.TXID)]=prevTx
		}
	}
	return tx.Verify(prevTxs)
}