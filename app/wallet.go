package app

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

func NewWallet() *Wallet {
	privateKey,err:=ecdsa.GenerateKey(elliptic.P256(),rand.Reader) //must use P256 curve or will result in invalid transaction
	if err!=nil {
		PrintError(err)
	}
	rawPublicKey:=privateKey.PublicKey
	publicKey:=append(rawPublicKey.X.Bytes(),rawPublicKey.Y.Bytes()...)
	return &Wallet{*privateKey,publicKey}
}

func (w *Wallet) GetAddress() string {
	h160:=HashPublicKey(w.PublicKey)
	payload:=append([]byte{0},h160...)
	checkcode:=CheckSum(payload)
	x:=append(payload,checkcode...)
	return base58.Encode(x)
}

func HashPublicKey(publicKey []byte) []byte {
	h:=sha256.Sum256(publicKey)
	ripemd160Hasher:=ripemd160.New()
	ripemd160Hasher.Write(h[:])
	return ripemd160Hasher.Sum(nil)
}

func CheckSum(payload []byte) []byte {
	y1:=sha256.Sum256(payload)
	y2:=sha256.Sum256(y1[:])
	return y2[:4]
}