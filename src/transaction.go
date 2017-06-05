package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"time"
)

//交易信息结构
type Transaction struct {
	Header    TranscationHeader
	Signature []byte //签名
	Payload   []byte //交易详情
}

//交易信息头部结构
type TranscationHeader struct {
	From          []byte //交易发送方
	To            []byte //交易接受方
	TimeStamp     uint32 //时间戳
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32 //随机数
}

//创建交易
func NewTransaction(from, to, payload []byte) *Transaction {
	t := Transaction{Header: TranscationHeader{From: from, To: to}, Payload: payload}
	t.Header.TimeStamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(payload)
	t.Header.PayloadLength = uint32(len(payload))

	return &t
}

//获取交易头部哈希值
func (t *Transaction) Hash() []byte {
	txhb, _ := t.Header.MarshalBinary()
	return SHA256(txhb)
}

//生成交易信息签名
func (t *Transaction) Sign(keypair *Keypair) []byte {
	s, _ := keypair.Sign(t.Hash())
	return s
}

//验证交易信息
//验证签名，payloadHash和Pow
func (t *Transaction) VerifyTransaction(pow []byte) bool {
	headHash := t.Hash()
	payloadHash := SHA256(t.Payload)

	return reflect.DeepEqual(payloadHash, t.Header.PayloadHash) &&
		CheckProofofWork(pow, headHash) &&
		SignatureVerify(t.Header.From, t.Signature, headHash)
}

//获取满足难度的随机值
func (t *Transaction) GenerateNonce(prefix []byte) uint32 {
	newT := t
	for {
		if CheckProofofWork(prefix, newT.Hash()) {
			break
		}
		newT.Header.Nonce++
	}
	return newT.Header.Nonce
}

//序列化交易信息
func (t *Transaction) MarshalBinary() ([]byte, error) {
	headerByters, _ := t.Header.MarshalBinary()

	if len(headerByters) != TRANSCATION_HEADER_SIZE {
		return nil, errors.New("序列化交易头部信息失败")
	}

	return append(append(headerByters, FitBytesInto(t.Signature, NETWORK_KEY_SIZE)...), t.Payload...), nil
}

//反序列化交易信息
func (t *Transaction) UnmarshalBinary(d []byte) ([]byte, error) {
	buf := bytes.NewBuffer(d)
	if len(d) < TRANSCATION_HEADER_SIZE+NETWORK_KEY_SIZE {
		return nil, errors.New("交易字节长度小于反序列化要求的长度")
	}
	header := &TranscationHeader{}
	if err := header.UnmarshalBinary(buf.Next(TRANSCATION_HEADER_SIZE)); err != nil {
		return nil, err
	}
	t.Header = *header
	t.Signature = StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	t.Payload = buf.Next(int(t.Header.PayloadLength))

	return buf.Next(MaxInt), nil
}

//序列化交易头部信息
func (th *TranscationHeader) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.Write(FitBytesInto(th.From, NETWORK_KEY_SIZE))
	buf.Write(FitBytesInto(th.To, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, th.TimeStamp)
	buf.Write(FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil
}

//反序列化交易头部信息
func (th *TranscationHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	th.From = StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	th.To = StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.TimeStamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}

//交易队列切片信息
type TransactionSlice []Transaction

//获取交易队列切片长度
func (ts TransactionSlice) Len() int {
	return len(ts)
}

//检查交易是否在交易队列切片中存在
func (ts TransactionSlice) Exists(tr Transaction) bool {
	for _, t := range ts {
		if reflect.DeepEqual(t.Hash(), tr.Hash()) {
			return true
		}
	}
	return false
}

//向切片中添加交易信息
func (ts TransactionSlice) AddTransaction(tr Transaction) TransactionSlice {
	//交易按时间排序
	for i, t := range ts {
		if t.Header.TimeStamp >= tr.Header.TimeStamp {
			return append(append(ts[:i], tr), ts[i:]...)
		}
	}
	return append(ts, tr)
}

//将交易队列序列化
func (ts TransactionSlice) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, t := range ts {
		bs, err := t.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(bs)
	}
	return buf.Bytes(), nil
}

//反序列化交易队列
func (ts *TransactionSlice) UnmarshalBinary(d []byte) error {
	remaining := d
	for len(remaining) > TRANSCATION_HEADER_SIZE+NETWORK_KEY_SIZE {
		t := new(Transaction)
		rem, err := t.UnmarshalBinary(remaining)
		if err != nil {
			return err
		}
		(*ts) = append((*ts), *t)
		remaining = rem
	}
	return nil
}
