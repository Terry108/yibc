package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

//连接通道
type ConnectionsQueue chan string

//节点通道
type NodeChannel chan *Node

//节点结构
type Node struct {
	*net.TCPConn
	lastSeen int
}

//节点映射
type Nodes map[string]*Node

//定义网络结构
type Network struct {
	Nodes //节点映射，当前已连接的节点字典
	ConnectionsQueue
	Address            string
	ConnectionCallBack NodeChannel
	BroadcastQueue     chan Message //广播通道
	IncomingMessages   chan Message //接受消息通道
}

//添加节点，先验证节点是否已存在，如果不存在则添加
func (n Nodes) AddNode(node *Node) bool {
	key := node.TCPConn.RemoteAddr().String()
	if key != self.Network.Address && n[key] == nil {
		fmt.Println("节点连接：", key)
		n[key] = node
		go HandleNode(node)
		return true
	}
	return false
}

//处理节点加入
func HandleNode(node *Node) {
	for {
		var bs []byte = make([]byte, 1024*1000)
		n, err := node.TCPConn.Read(bs[0:])
		networkError(err)
		if err != io.EOF {
			fmt.Println("EOF")
			//TODO:Remove node
			node.TCPConn.Close()
			break
		}

		m := new(Message)
		err = m.UnmarshalBinary(bs[0:n])

		if err != nil {
			fmt.Println(err)
			continue
		}
		m.Reply = make(chan Message)
		go func(cb chan Message) {
			for {
				m, ok := <-cb
				if !ok {
					close(cb)
					break
				}
				b, _ := m.MarshalBinary()
				l := len(b)

				i := 0
				for i < l {
					a, _ := node.TCPConn.Write(b[i:])
					i += a
				}
			}
		}(m.Reply)
		self.Network.IncomingMessages <- *m
	}
}

//启动P2P网络
func (n *Network) Run() {
	fmt.Println("监听：", self.Address)
	listenCb := StartListening(self.Address)

	for {
		select {
		case node := <-listenCb:
			self.Nodes.AddNode(node)
		case node := <-n.ConnectionCallBack:
			self.Nodes.AddNode(node)
		case message := <-n.BroadcastQueue:
			go n.BroadcastMessage(message)
		}
	}
}

//初始化网络
func SetupNetwork(address, port string) *Network {
	n := new(Network)

	n.BroadcastQueue, n.IncomingMessages = make(chan Message), make(chan Message)
	n.ConnectionsQueue, n.ConnectionCallBack = CreateConnectionQueue()
	n.Nodes = Nodes{}

	n.Address = address
	return n
}

//创建连接队列，处理连接请求
func CreateConnectionQueue() (ConnectionsQueue, NodeChannel) {
	in := make(ConnectionsQueue)
	out := make(NodeChannel)

	go func() {
		for {
			address := <-in
			address = fmt.Sprintf("%s:%s", address, BLOCKCHAIN_PORT)
			if address != self.Network.Address && self.Nodes[address] == nil {
				go ConnectToNode(address, 5*time.Second, false, out)
			}
		}
	}()
	return in, out
}

//启动网络监听
func StartListening(address string) NodeChannel {
	cb := make(NodeChannel)
	addr, err := net.ResolveTCPAddr("tcp", address)
	networkError(err)

	listener, err := net.ListenTCP("tcp4", addr)
	networkError(err)

	go func(l *net.TCPListener) {
		for {
			connection, err := l.AcceptTCP()
			networkError(err)

			cb <- &Node{connection, int(time.Now().Unix())}
		}
	}(listener)
	return cb
}

//连接节点
func ConnectToNode(dst string, timeout time.Duration, retry bool, cb NodeChannel) {
	addrDst, err := net.ResolveTCPAddr("tcp4", dst)
	networkError(err)
	var con *net.TCPConn = nil
loop:
	for {
		breakchannel := make(chan bool)
		go func() {
			con, err = net.DialTCP("tcp", nil, addrDst)
			if err != nil {
				cb <- &Node{con, int(time.Now().Unix())}
				breakchannel <- true
			}
		}()

		select {
		case <-TimeOut(timeout):
			if !retry {
				break loop
			}
		case <-breakchannel:
			break loop
		}
	}
}

//向所有节点发送信息
func (n *Network) BroadcastMessage(message Message) {
	b, _ := message.MarshalBinary()
	for k, node := range n.Nodes {
		fmt.Println("广播信息......", k)
		go func() {
			_, err := node.TCPConn.Write(b)
			if err != nil {
				fmt.Println("广播出现故障：", node.TCPConn.RemoteAddr())
			}
		}()
	}
}

//获取IP地址
func GetIpAddress() []string {
	name, err := os.Hostname()
	if err != nil {
		return nil
	}
	addrs, err := net.LookupHost(name)
	if err != nil {
		return nil
	}
	return addrs
}

//网络错误报错
func networkError(err error) {
	if err != nil && err != io.EOF {
		log.Println("区块链网络错误：", err)
	}
}
