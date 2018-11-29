package app

import (
	"math/big"
	"crypto/sha256"
)

type POW struct {
	Block *Block
	Target big.Int
}

func (p *POW) Mining() ([]byte,int64) {
	var h [32]byte
	var nonce int64
	for  {
		h=sha256.Sum256(p.Block.GetBytes(nonce))
		bt:=big.Int{}
		bt.SetBytes(h[:])
		if bt.Cmp(&p.Target)==-1 {
			break
		}
		nonce++
	}
	return h[:],nonce
}

func (p *POW) IsValid() bool {
	h:=sha256.Sum256(p.Block.GetBytes(p.Block.Nonce))
	bt:=big.Int{}
	bt.SetBytes(h[:])
	return bt.Cmp(&p.Target)==-1
}