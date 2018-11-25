package app

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"crypto/elliptic"
	"strings"
)

type Transaction struct {
	TXID []byte
	TXInputs []TXInput
	TXOutputs []TXOutput
}

const Reward=50.0

func NewCoinbase(address string,data string) *Transaction {
	if data=="" {
		data=fmt.Sprintf("reward %s %f",address,Reward)
	}
	txi:=TXInput{nil,-1,nil,[]byte(data)}
	txo:=NewTXOutput(Reward,address)
	tx:=Transaction{nil,[]TXInput{txi},[]TXOutput{txo}}
	tx.GetTXID()
	return &tx
}

func (tx *Transaction) GetTXID()  {
	var buffer bytes.Buffer
	encoder:=gob.NewEncoder(&buffer)
	encoder.Encode(tx)
	h:=sha256.Sum256(buffer.Bytes())
	tx.TXID=h[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.TXInputs)==1 && tx.TXInputs[0].TXID==nil && tx.TXInputs[0].OutputIndex==-1
}

func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey,prevTxs map[string]*Transaction)  {
	if tx.IsCoinbase() {
		return
	}
	_tx:=tx.TrimmedCopy()
	for k,txi:=range _tx.TXInputs{
		prevTx:=prevTxs[string(txi.TXID)]
		if prevTx==nil {
			panic("transaction not found")
		}
		_tx.TXInputs[k].PublicKey=prevTx.TXOutputs[txi.OutputIndex].PublicKeyHash
		_tx.GetTXID()
		_tx.TXInputs[k].PublicKey=nil
		r,s,err:=ecdsa.Sign(rand.Reader,privateKey,_tx.TXID)
		if err!=nil {
			PrintError(err)
		}
		tx.TXInputs[k].Signature=append(r.Bytes(),s.Bytes()...)
	}
}

func (tx *Transaction) Verify(prevTxs map[string]*Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	_tx:=tx.TrimmedCopy()
	for k,txi:=range tx.TXInputs{
		prevTx:=prevTxs[string(txi.TXID)]
		if prevTx==nil {
			panic("transaction not found")
		}
		_tx.TXInputs[k].PublicKey=prevTx.TXOutputs[txi.OutputIndex].PublicKeyHash
		_tx.GetTXID()
		_tx.TXInputs[k].PublicKey=nil
		r:=big.Int{}
		r.SetBytes(txi.Signature[:len(txi.Signature)/2])
		s:=big.Int{}
		s.SetBytes(txi.Signature[len(txi.Signature)/2:])
		x:=big.Int{}
		x.SetBytes(txi.PublicKey[:len(txi.PublicKey)/2])
		y:=big.Int{}
		y.SetBytes(txi.PublicKey[len(txi.PublicKey)/2:])
		rawPublicKey:=ecdsa.PublicKey{elliptic.P256(),&x,&y}
		if !ecdsa.Verify(&rawPublicKey,_tx.TXID,&r,&s) {
			return false
		}
	}
	return true
}

func (tx *Transaction) TrimmedCopy() *Transaction {
	var txis []TXInput
	for _,txi:=range tx.TXInputs{
		txis=append(txis,TXInput{txi.TXID,txi.OutputIndex,nil,nil})
	}
	var txos []TXOutput
	for _,txo:=range tx.TXOutputs{
		txos=append(txos,TXOutput{txo.Value,txo.PublicKeyHash})
	}
	_tx:=Transaction{tx.TXID,txis,txos}
	return &_tx
}

func (tx *Transaction) String() string {
	var lines []string
	lines=append(lines,fmt.Sprintf("   Transaction %x",tx.TXID))
	for k,txi:=range tx.TXInputs{
		lines=append(lines,fmt.Sprintf("    Input %d",k))
		lines=append(lines,fmt.Sprintf("      TXID:%x",txi.TXID))
		lines=append(lines,fmt.Sprintf("      OutputIndex:%d",txi.OutputIndex))
		lines=append(lines,fmt.Sprintf("      Signature:%x",txi.Signature))
		lines=append(lines,fmt.Sprintf("      PublicKey:%x",txi.PublicKey))
	}
	for k,txo:=range tx.TXOutputs{
		lines=append(lines,fmt.Sprintf("    Output %d",k))
		lines=append(lines,fmt.Sprintf("      Value:%f",txo.Value))
		lines=append(lines,fmt.Sprintf("      PublicKeyHash:%x",txo.PublicKeyHash))
	}
	return strings.Join(lines,"\n")
}