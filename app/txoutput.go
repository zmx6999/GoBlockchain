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
	txo.Lock(address)
	return txo
}

func (txo *TXOutput) Lock(address string)  {
	x:=base58.Decode(address)
	h:=x[1:len(x)-4]
	txo.PublicKeyHash=h
}

func (txo *TXOutput) CanBeUnlockedWith(publicKeyHash []byte) bool {
	return bytes.Compare(txo.PublicKeyHash,publicKeyHash)==0
}