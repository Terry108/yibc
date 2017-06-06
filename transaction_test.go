package main

import (
	"reflect"
	"testing"
)

func TestTrancationMerchalling(t *testing.T) {
	kp := GenerateNewKeypair()
	tr := NewTransaction(kp.Public, nil, []byte(RandomString(RandomInt(0, 1024*1024))))
	tr.Header.Nonce = tr.GenerateNonce(ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX))
	tr.Signature = tr.Sign(kp)

	data, err := tr.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	newT := &Transaction{}
	rem, err := newT.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*newT, *tr) || len(rem) < 0 {
		t.Error("序列化，反序列化失败")
	}
}

func TestTransationVerfication(t *testing.T) {

	pow := ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	kp := GenerateNewKeypair()
	tr := NewTransaction(kp.Public, nil, []byte(RandomString(RandomInt(0, 1024))))
	tr.Header.Nonce = tr.GenerateNonce(pow)
	tr.Signature = tr.Sign(kp)

	if !tr.VerifyTransaction(pow) {
		t.Error("验证交易失败")
	}
}

func TestIncorrectTransationPoWVerfication(t *testing.T) {
	pow := ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	powIncorrect := ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, 'a')

	kp := GenerateNewKeypair()
	tr := NewTransaction(kp.Public, nil, []byte(RandomString(RandomInt(0, 1024))))
	tr.Header.Nonce = tr.GenerateNonce(powIncorrect)
	tr.Signature = tr.Sign(kp)

	if tr.VerifyTransaction(pow) {
		t.Error("未经POW即可验证通过")
	}
}

func TestIncorrectTransationSignatureVerfication(t *testing.T) {
	pow := ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	kp1, kp2 := GenerateNewKeypair(), GenerateNewKeypair()
	tr := NewTransaction(kp2.Public, nil, []byte(RandomString(RandomInt(0, 1024))))
	tr.Header.Nonce = tr.GenerateNonce(pow)
	tr.Signature = tr.Sign(kp1)

	if tr.VerifyTransaction(pow) {
		t.Error("错误的密钥对可以验证通过")
	}
}
