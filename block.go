package main

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

//区块结构
type Block struct {
	*BlockHeader
	Signture []byte
	*TransactionSlice
}

//区块头部结构
type BlockHeader struct {
	Origin     []byte //记账者公钥
	PreBlock   []byte //前区块哈希值
	MerkelRoot []byte //Merkel根值
	TimeStamp  uint32 //时间戳
	Nonce      uint32 //随机数
}

//新建区块
func NewBlock(previousBlock []byte) Block {
	header := &BlockHeader{PreBlock: previousBlock}
	return Block{header, nil, new(TransactionSlice)}
}

//添加交易信息
func (b *Block) AddTransaction(t *Transaction) {
	newSlice := b.TransactionSlice.AddTransaction(*t)
	b.TransactionSlice = &newSlice
}

//区块签名
func (b *Block) Sign(keypair *Keypair) []byte {
	s, _ := keypair.Sign(b.Hash())
	return s
}

//验证区块
func (b *Block) VerifyBlock(prefix []byte) bool {
	headerHash := b.Hash()
	merkel := b.GenerateMerkelRoot()

	return reflect.DeepEqual(merkel, b.BlockHeader.MerkelRoot) &&
		CheckProofofWork(prefix, headerHash) &&
		SignatureVerify(b.BlockHeader.Origin, b.Signture, headerHash)
}

//获取区块的哈希值
func (b *Block) Hash() []byte {
	headHash, _ := b.BlockHeader.MarshalBinary()
	return SHA256(headHash)
}

//生成Merkel根值
func (b *Block) GenerateMerkelRoot() []byte {

	var merkel func(hashes [][]byte) []byte
	merkel = func(hashes [][]byte) []byte {

		l := len(hashes)
		if l == 0 {
			return nil
		}
		if l == 1 {
			return hashes[0]
		} else {
			if l%2 == 1 {
				return merkel([][]byte{merkel(hashes[:l-1]), hashes[l-1]})
			}

			bs := make([][]byte, l/2)
			for i, _ := range bs {
				j, k := i*2, (i*2)+1
				bs[i] = SHA256(append(hashes[j], hashes[k]...))
			}
			return merkel(bs)
		}
	}
	ts := Map(func(t Transaction) []byte { return t.Hash() },
		[]Transaction(*b.TransactionSlice)).([][]byte)
	return merkel(ts)
}

//序列化区块信息
func (b *Block) MarshalBinary() ([]byte, error) {
	bhb, err := b.BlockHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sign := FitBytesInto(b.Signture, NETWORK_KEY_SIZE)
	tsb, err := b.TransactionSlice.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append(append(bhb, sign...), tsb...), nil
}

//反序列化区块信息
func (b *Block) UnmarshalBinary(d []byte) error {
	buf := bytes.NewBuffer(d)
	header := new(BlockHeader)
	err := header.UnmarshalBinary(buf.Next(BLOCK_HEADER_SIZE))
	if err != nil {
		return err
	}
	b.BlockHeader = header
	b.Signture = StripByte(buf.Next(NETWORK_KEY_SIZE), 0)

	ts := new(TransactionSlice)
	err = ts.UnmarshalBinary(buf.Next(MaxInt))
	if err != nil {
		return err
	}
	b.TransactionSlice = ts

	return nil
}

//序列化区块链头部
func (bh *BlockHeader) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.Write(FitBytesInto(bh.Origin, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, bh.TimeStamp)
	buf.Write(FitBytesInto(bh.PreBlock, 32))
	buf.Write(FitBytesInto(bh.MerkelRoot, 32))
	binary.Write(buf, binary.LittleEndian, bh.Nonce)

	return buf.Bytes(), nil
}

//反序列化区块链头部
func (bh *BlockHeader) UnmarshalBinary(d []byte) error {
	buf := bytes.NewBuffer(d)
	bh.Origin = StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &bh.TimeStamp)
	bh.PreBlock = buf.Next(32)
	bh.MerkelRoot = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &bh.Nonce)

	return nil
}

//区块切片
type BlockSlice []Block

//检查区块是否在区块切片中存在
func (bs BlockSlice) Exists(b Block) bool {
	//	for _, bi := range bs {
	//		if reflect.DeepEqual(bi.Hash(), b.Hash()) {
	//			return true
	//		}
	//	}

	//优化：新区块最有可能在顶部，故倒叙比对
	l := len(bs)
	for i := l - 1; i >= 0; i-- {
		bb := bs[i]
		if reflect.DeepEqual(bb.Signture, b.Signture) {
			return true
		}
	}

	return false
}

//获取前一区块
func (bs BlockSlice) PreviousBlock() *Block {
	l := len(bs)
	if l > 0 {
		return &bs[l-1]
	}
	return nil
}
