package Kademlia

import (
	"net"
	"net/rpc"
	"time"
)

const K int = 20
const alpha int32 = 3
const tryTimes int = 4
const localAddress string = "127.0.0.1"
const WaitTime time.Duration = 250 * time.Millisecond
const SleepTime time.Duration = 20 * time.Second
const refreshTimeInterval time.Duration = 30 * time.Second
const expireTimeInterval2 time.Duration = 6 * time.Hour
const expireTimeInterval3 time.Duration = 20 * time.Minute
const republicTimeInterval time.Duration = 5 * time.Hour
const duplicateTimeInterval time.Duration = 15 * time.Minute
const backgroundInterval1 time.Duration = 5 * time.Second
const backgroundInterval2 time.Duration = 10 * time.Minute

type Contact struct {
	SortKey     [20]byte
	ContactInfo Contact
}

type ContactRecord struct {
	SortKey     [20]byte
	ContactInfo Contact
}

type Node struct {
	station   *network
	isRunning bool
	data      database
	table     RoutingTable
	addr      Contact
}

type network struct {
	serv       *rpc.Server
	lis        net.Listener
	nodePtr    *nodeWrapper
	QuitSignal chan bool
}

type nodeWrapper struct {
	node *Node
}

type FindNodeRequest struct {
	Requester Contact
	Target    [20]byte
}

type FindNodeReply struct {
	Requester Contact
	Replier   Contact
	Content   []ContactRecord
}

type StoreRequest struct {
	Key          string
	Value        string
	RequesterPri int
	Requester    Contact
}

type FindValueRequest struct {
	Key       string
	Requester Contact
}

type FindValueReply struct {
	Requester Contact
	Replier   Contact
	Content   []ContactRecord
	IsFind    bool
	Value     string
}

func (this *Node) reset() {
	this.isRunning = false
	this.table.InitRoutingTable(this.addr)
	this.data.init()
}
