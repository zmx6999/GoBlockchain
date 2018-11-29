package app

import (
	"time"
	"crypto/sha256"
	"bytes"
	"encoding/binary"
	"math/big"
	"encoding/gob"
	"fmt"
	"strings"
)

type Block struct {
	Version int64
	PrevBlockHash []byte
	MerkleRoot []byte
	Timestamp int64
	Difficulty int64
	Nonce int64
	Transactions []*Transaction
	Hash []byte
}

func NewBlock(txs []*Transaction,prevBlockHash []byte) *Block {
	block:=Block{
		0,
		prevBlockHash,
		[]byte{},
		time.Now().Unix(),
		20,
		0,
		txs,
		[]byte{},
	}
	block.GetMerkleRoot()
	pow:=block.NewPOW()
	block.Hash,block.Nonce=pow.Mining()
	return &block
}

func (b *Block) GetMerkleRoot()  {
	var t [][]byte
	for _,tx:=range b.Transactions{
		t=append(t,tx.TXID)
	}
	h:=sha256.Sum256(bytes.Join(t,[]byte{}))
	b.MerkleRoot=h[:]
}

func (b *Block) GetBytes(nonce int64) []byte {
	t:=[][]byte{
		int64ToBytes(b.Version),
		b.PrevBlockHash,
		b.MerkleRoot,
		int64ToBytes(b.Timestamp),
		int64ToBytes(b.Difficulty),
		int64ToBytes(nonce),
	}
	return bytes.Join(t,[]byte{})
}

func int64ToBytes(x int64) []byte {
	var buffer bytes.Buffer
	binary.Write(&buffer,binary.BigEndian,x)
	return buffer.Bytes()
}

func (b *Block) NewPOW() *POW {
	bt:=big.NewInt(1)
	bt.Lsh(bt,256-uint(b.Difficulty))
	return &POW{b,*bt}
}

func Serialize(block *Block) []byte {
	var buffer bytes.Buffer
	encoder:=gob.NewEncoder(&buffer)
	encoder.Encode(block)
	return buffer.Bytes()
}

func Unserialize(data []byte) *Block {
	var buffer bytes.Buffer
	buffer.Write(data)
	decoder:=gob.NewDecoder(&buffer)
	var block Block
	decoder.Decode(&block)
	return &block
}

func (b *Block) String() string {
	lines:=[]string{fmt.Sprintf("          Block %x",b.Hash)}
	lines=append(lines,fmt.Sprintf("Version:%d",b.Version))
	lines=append(lines,fmt.Sprintf("PrevBlockHash:%x",b.PrevBlockHash))
	lines=append(lines,fmt.Sprintf("MerkleRoot:%x",b.MerkleRoot))
	lines=append(lines,fmt.Sprintf("Timestamp:%d",b.Timestamp))
	lines=append(lines,fmt.Sprintf("Difficulty:%d",b.Difficulty))
	lines=append(lines,fmt.Sprintf("Nonce:%d",b.Nonce))
	pow:=b.NewPOW()
	lines=append(lines,fmt.Sprintf("Valid:%t",pow.IsValid()))
	return strings.Join(lines,"\n")
}