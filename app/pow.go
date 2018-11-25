package app

import (
	"math/big"
	"crypto/sha256"
)

type ProofOfWork struct {
	block *Block
	target big.Int
}

func (pow *ProofOfWork) Mining() (int64,[]byte) {
	var nonce int64
	var h [32]byte
	for  {
		h=sha256.Sum256(pow.block.GetHash(nonce))
		bt:=big.Int{}
		bt.SetBytes(h[:])
		if bt.Cmp(&pow.target)==-1 {
			break
		}
		nonce++
	}
	return nonce,h[:]
}

func (pow *ProofOfWork) IsValid() bool {
	h:=sha256.Sum256(pow.block.GetHash(pow.block.Nonce))
	bt:=big.Int{}
	bt.SetBytes(h[:])
	return bt.Cmp(&pow.target)==-1
}