package app

import "bytes"

type TXInput struct {
	TXID []byte
	OutIndex int64
	Signature []byte
	PublicKey []byte
}

func (txi *TXInput) CanUnlockWith(publicKeyHash []byte) bool {
	return bytes.Compare(HashPublicKey(txi.PublicKey),publicKeyHash)==0
}