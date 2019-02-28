package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
)

func CheckSignature(message string, signature string, address string) bool {
	msg := crypto.Keccak256([]byte(message))

	sig := hexutil.MustDecode(signature)
	if len(sig) != 65 {
		log.Println("signature must be 65 bytes long")
		return false
	}
	if sig[64] != 27 && sig[64] != 28 {
		log.Println("invalid Ethereum signature (V is not 27 or 28)")
		return false
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	pubKey, err := crypto.SigToPub(signHash(msg), sig)
	if err != nil {
		return false
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	addr := common.HexToAddress(address)

	return recoveredAddr == addr
}

func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}
