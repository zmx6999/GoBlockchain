package app

import (
	"bolt"
	"errors"
)

type BlockChainIterator struct {
	db *bolt.DB
	currentBlockHash []byte
}

func (bci *BlockChainIterator) GetCurrentBlock() *Block {
	var block *Block
	bci.db.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(BucketName))
		if bucket==nil {
			return errors.New("bucket "+BucketName+" not exists")
		} else {
			data:=bucket.Get(bci.currentBlockHash)
			block=Unserialize(data)
			bci.currentBlockHash=block.PrevBlockHash
			return nil
		}
	})
	return block
}