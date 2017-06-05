package main

import (
	"fmt"
	"reflect"
	"time"
)

//交易队列通道
type TransactionQuene chan *Transaction

//区块队列通道
type BlockQuene chan Block

//区块结构
type BlockChain struct {
	CurrentBlock Block //当前区块
	BlockSlice         //区块切片

	TransactionQuene
	BlockQuene
}

//初始化区块链
func SetupBlockChain() *BlockChain {
	bc := new(BlockChain)
	//TransactionQuene, BlockQuene := make(TransactionQuene), make(BlockQuene)

	//TODO:从区块文件中读取
	bc.CurrentBlock = bc.CreateNewBlock()
	return bc
}

//创建新区块
func (bc *BlockChain) CreateNewBlock() Block {
	prev := bc.BlockSlice.PreviousBlock()

	prevBlockHash := []byte{}
	if prev != nil {
		prevBlockHash = prev.Hash()
	}
	nb := NewBlock(prevBlockHash)
	nb.BlockHeader.Origin = self.Keypair.Public

	return nb
}

//向区块链中添加区块
func (bc *BlockChain) AddBlock(b Block) {
	bc.BlockSlice = append(bc.BlockSlice, b)
}

//启动区块链
func (bc *BlockChain) Run() {
	interruptBlockGen := bc.GenerateBlock()
	for {
		select {
		//处理交易
		case tr := <-bc.TransactionQuene:
			if bc.CurrentBlock.TransactionSlice.Exists(*tr) {
				continue
			}
			if !tr.VerifyTransaction(TRANSCATION_POW) {
				fmt.Println("收到未经验证的交易信息:", tr)
				continue
			}
			bc.CurrentBlock.AddTransaction(*tr)
			interruptBlockGen <- bc.CurrentBlock
			//将交易广播到网络
			mes := NewMessage(MESSAGE_SEND_TRANSACTION)
			mes.Data, _ = tr.MarshalBinary()

			time.Sleep(300 * time.Millisecond)
			self.Network.BroadcastQueue <- *mes

		//区块处理
		case b := <-bc.BlockQuene:
			if bc.Exists(b) {
				continue
			}
			if !b.VerifyBlock(BLOCK_POW) {
				fmt.Println("区块未验证通过，不符合难度要求。")
				continue
			}
			if !reflect.DeepEqual(b.PreBlock, bc.CurrentBlock.Hash()) {
				//TODO：区块孤儿池的实现
				fmt.Println("缺失区块")
			} else {
				fmt.Println("新区块", b.Hash())
				transDiff := TransactionSlice{}
				if !reflect.DeepEqual(b.BlockHeader.MerkelRoot, bc.CurrentBlock.MerkelRoot) {
					//被打包的交易信息有差别
					transDiff = DiffTransactionSlices(*b.TransactionSlice, *bc.CurrentBlock.TransactionSlice)
				}
				bc.AddBlock(b)

				//广播区块
				mes := NewMessage(MESSAGE_SEND_BLOCK)
				mes.Data, _ = b.MarshalBinary()
				self.Network.BroadcastQueue <- *mes

				//新区块
				bc.CurrentBlock = bc.CreateNewBlock()
				bc.CurrentBlock.TransactionSlice = &transDiff

				interruptBlockGen <- bc.CurrentBlock
			}
		}
	}
}

//比对交易切片，返回不同交易
func DiffTransactionSlices(a, b TransactionSlice) (diff TransactionSlice) {
	//假设交易队列是有序的
	lastj := 0
	for _, t := range a {
		found := false
		for j := lastj; j < len(b); j++ {
			if reflect.DeepEqual(t.Signature, b[j].Signature) {
				found = true
				lastj = j
				break
			}
		}
		if !found {
			diff = append(diff, t)
		}
	}
	return diff
}

//生成区块
//当收到新的区块或交易时，打断挖矿，重新开始挖矿
func (bc *BlockChain) GenerateBlock() chan Block {
	//当收到新的区块时，打断当前挖矿计算
	interrupt := make(chan Block)

	go func() {
		block := <-interrupt
	Loop:
		fmt.Println("开始挖矿啦！")
		block.BlockHeader.MerkelRoot = block.GenerateMerkelRoot()
		block.BlockHeader.Nonce = 0
		block.BlockHeader.TimeStamp = uint32(time.Now().Unix())
		for {
			sleepTime := time.Nanosecond
			if block.TransactionSlice.Len() > 0 {
				if CheckProofofWork(BLOCK_POW, block.Hash()) {

					block.Signture = block.Sign(self.Keypair)
					bc.BlockQuene <- block
					sleepTime = time.Hour * 24
					fmt.Println("哇！挖矿成功，生成区块！")
				} else {
					block.BlockHeader.Nonce++
				}
			} else {
				sleepTime = time.Hour * 24
				fmt.Println("无交易信息，休息会~")
			}

			select {
			case block = <-interrupt:
				goto Loop
			case <-TimeOut(sleepTime):
				continue
			}
		}
	}()
	return interrupt
}
