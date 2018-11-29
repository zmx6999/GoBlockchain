package app

import (
	"bolt"
	"errors"
)

type BlockChainIterator struct {
	DB *bolt.DB
	CurrentBlockHash []byte
}

func (bci *BlockChainIterator) GetCurrentBlock() *Block {
	var block *Block
	bci.DB.View(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			return errors.New("bucket "+BucketName+" doesn't exist")
		} else {
			data:=bucket.Get(bci.CurrentBlockHash)
			block=Unserialize(data)
			bci.CurrentBlockHash=block.PrevBlockHash
			return nil
		}
	})
	return block
}

func (bci *BlockChainIterator) Travel(fcn func(block *Block))  {
	for  {
		block:=bci.GetCurrentBlock()
		fcn(block)
		if len(block.PrevBlockHash)==0 {
			break
		}
	}
}