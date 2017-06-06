# yibc Blockchain 区块链最小的实现
Blockchain, learn and have a try
最小实现原理参考文章https://www.igvita.com/2014/05/05/minimum-viable-block-chain/
挖矿：
采用PoW共识机制，难度可配置。
密码学：
使用go语言加密包中 ECDSA (224 bits)获取密钥对，然后使用base58进行编码。


