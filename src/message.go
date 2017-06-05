package main

import (
	"bytes"
	"errors"
)

//消息结构
type Message struct {
	Identifier byte   //标识符
	Options    []byte //消息类型，包括交易和区块信息
	Data       []byte //消息内容

	Reply chan Message
}

//新建消息
func NewMessage(id byte) *Message {
	return &Message{Identifier: id}
}

//将消息序列化
func (m *Message) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.WriteByte(m.Identifier)
	buf.Write(FitBytesInto(m.Options, MESSAGE_OPTIONS_SIZE))
	buf.Write(m.Data)

	return buf.Bytes(), nil
}

//反序列化消息
func (m *Message) UnmarshalBinary(d []byte) error {

	if len(d) < MESSAGE_OPTIONS_SIZE+MESSAGE_TYPE_SIZE {
		return errors.New("Insuficient message size")
	}

	buf := bytes.NewBuffer(d)
	m.Identifier = buf.Next(1)[0]
	m.Options = StripByte(buf.Next(MESSAGE_OPTIONS_SIZE), 0)
	m.Data = buf.Next(MaxInt)

	return nil
}
