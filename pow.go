package main

import (
	"reflect"
)

var (
	//交易信息计算难度值
	TRANSACTION_POW = ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	//区块计算难度值
	BLOCK_POW = ArrayOfBytes(BLOCK_POW_COMPLEXITY, POW_PREFIX)
)

//验证计算的难度值是否符合要求
//验证方法：比较前端的0是否相等。0越多，代表难度值越大
func CheckProofofWork(predix, hash []byte) bool {
	if len(predix) > 0 {
		return reflect.DeepEqual(predix, hash[:len(predix)])
	}
	return true
}
