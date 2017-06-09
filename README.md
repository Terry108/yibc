# yibc Blockchain 区块链最小的实现
Blockchain, learn and have a try
最小实现原理参考文章https://www.igvita.com/2014/05/05/minimum-viable-block-chain/
## 挖矿：
采用PoW共识机制，难度可配置。
## 密码学：
使用go语言加密包中 ECDSA (224 bits)获取密钥对，然后使用base58进行编码。
## 区块
区块头部
*  Origin：记账者公钥，80字节
*  PreBlock：前区块哈希值，32字节
*  MerkelRoot：Merkel根值，32字节
*  TimeStamp：时间戳，4字节
*  Nonce：随机数，4字节
签名：signed(sha256(header))
区块交易信息
## 交易信息
头部信息
* From          []byte //交易发送方
* To            []byte //交易接受方
* TimeStamp     uint32 //时间戳
* PayloadHash   []byte //sha256(交易数据)
* PayloadLength uint32 //交易数据长度
* Nonce         uint32 //随机数
交易签名
交易详情
## 消息
消息类型：

	const (
	MESSAGE_GET_NODES = iota + 20
	MESSAGE_SEND_NODES

	MESSAGE_GET_TRANSACTION
	MESSAGE_SEND_TRANSACTION

	MESSAGE_GET_BLOCK
	MESSAGE_SEND_BLOCK
	)

Options    []byte //消息类型，包括交易和区块信息
Data       []byte //消息内容



