package main

import (
	//	"fmt"
	"testing"
)

//测试密钥对生成
func TestKeyGeneration(t *testing.T) {
	keypair := GenerateNewKeypair()
	//	fmt.Println(len(string(keypair.Public)), string(keypair.Public))
	//	fmt.Println(len(string(keypair.Private)), string(keypair.Private))
	if len(keypair.Public) > 80 {
		t.Error("Error generating key")
	}
}

//测试base58，签名和验证签名
func TestKeySign(t *testing.T) {
	for i := 0; i < 5; i++ {
		keypair := GenerateNewKeypair()
		data := ArrayOfBytes(i, 'a')
		hash := SHA256(data)

		signature, err := keypair.Sign(hash)
		if err != nil {
			t.Error("base58 error")
		} else if !SignatureVerify(keypair.Public, signature, hash) {
			t.Error("Signing and verifing error", len(keypair.Public))
		}
	}
}
