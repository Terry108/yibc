package main

import (
	"testing"
)

func TestPoW(t *testing.T) {
	b1 := CheckProofofWork([]byte{0, 0, 0, 1, 2, 3}, []byte{0, 0, 0, 1, 2, 3, 4, 5})
	b2 := CheckProofofWork([]byte{0, 0}, []byte("hello"))
	b3 := CheckProofofWork(BLOCK_POW, append(BLOCK_POW, 1))
	b4 := CheckProofofWork(nil, []byte("Hello"))

	if !b1 || b2 || !b3 || !b4 {
		t.Error("PoW测试未通过")
	}
}
