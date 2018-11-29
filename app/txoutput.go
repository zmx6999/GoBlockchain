package app

import (
	"github.com/btcsuite/btcutil/base58"
	"bytes"
)

type TXOutput struct {
	Value float64
	PublicKeyHash []byte
}

func NewTXOutput(value float64,address string) TXOutput {
	txo:=TXOutput{value,nil}
	x:=base58.Decode(address)
	txo.PublicKeyHash=x[1:len(x)-4]
	return txo
}

func (txo *TXOutput) CanBeUnlockedWith(publicKeyHash []byte) bool {
	return bytes.Compare(txo.PublicKeyHash,publicKeyHash)==0
}