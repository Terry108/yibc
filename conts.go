package main

import "fmt"

const (
	BLOCKCHAIN_PORT  = "9119"
	NETWORK_KEY_SIZE = 80
	KEY_SIZE         = 28

	TRANSACTION_POW_COMPLEXITY = 1 //交易计算难度
	TRANSCATION_HEADER_SIZE    = NETWORK_KEY_SIZE /*From key*/ + NETWORK_KEY_SIZE /*To key*/ +
		4 /*int32 TimeStamp*/ + 32 /*sha256 payload hash*/ + 4 /*int32 payload length*/ + 4 /*int32 nonce*/
	BLOCK_HEADER_SIZE = NETWORK_KEY_SIZE /*orgin key*/ + 4 /*int32 timeStamp*/ +
		32 /*prev block hash*/ + 32 /*merkel hash*/ + 4 /*int32 nonce*/

	BLOCK_POW_COMPLEXITY = 3 //区块计算难度

	POW_PREFIX = 0 //复杂度前缀

	MESSAGE_TYPE_SIZE    = 1
	MESSAGE_OPTIONS_SIZE = 4
)

const (
	MESSAGE_GET_NODES = iota + 20
	MESSAGE_SEND_NODES

	MESSAGE_GET_TRANSACTION
	MESSAGE_SEND_TRANSACTION

	MESSAGE_GET_BLOCK
	MESSAGE_SEND_BLOCK
)

func SEED_NODE() []string {
	nodes := []string{"10.0.5.33"}
	for i := 0; i < 100; i++ {
		nodes = append(nodes, fmt.Sprintf("172.17.0.%d", i))
	}

	return nodes
}
