package chord

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"time"
)

type network struct {
	serv    *rpc.Server
	lis     net.Listener
	nodePtr *NodeWrapper
}

func MyAccept(server *rpc.Server, lis net.Listener, ptr *Node) {
	for {
		conn, err := lis.Accept()
		select {
		case <-ptr.QuitSignal:
			return
		default:
			if err != nil {
				fmt.Println("rpc.Serve: accept:", err.Error())
				return
			}
			go server.ServeConn(conn)
		}
	}
}

func (this *network) Init(address string, ptr *Node) error {
	this.serv = rpc.NewServer()
	this.nodePtr = new(NodeWrapper)
	this.nodePtr.node = ptr

	err1 := this.serv.Register(this.nodePtr)
	if err1 != nil {
		log.Errorf("<RPC Init>fail to register in address : %s\n", address)
		return err1
	}

	this.lis, err1 = net.Listen("tcp", address)
	if err1 != nil {
		log.Errorf("<RPC init> service start success in %s\n", address)
		return err1
	}
	//log.Infof("<RPC Init> service start success in %s\n", address)
	go MyAccept(this.serv, this.lis, this.nodePtr.node)
	return nil
}

func GetClient(address string) (*rpc.Client, error) {
	if address == "" {
		log.Warningln("<GetClient> IP address is nil")
		return nil, errors.New("<GetClient> IP address is nil")
	}
	var client *rpc.Client
	var err error
	ch := make(chan error)
	for i := 0; i < 5; i++ {
		go func() {
			client, err = rpc.Dial("tcp", address)
			ch <- err
		}()
		select {
		case <-ch:
			if err == nil {
				return client, nil
			} else {
				return nil, err
			}
		case <-time.After(waitTime):
			err = errors.New("Timeout")
			//log.Infoln("<GetClient> Timeout to ", address)
		}
	}
	return nil, err
}

func CheckOnline(address string) bool {
	client, err := GetClient(address)
	if err != nil {
		log.Infoln("<CheckOnline> Ping Fail in ", address, "because : ", err)
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}
	//log.Infoln("<CheckOnline Ping Online in ", address)
	return true
}

func RemoteCall(targetNode string, funcClass string, input interface{}, result interface{}) error {
	if targetNode == "" {
		log.Warningln("<RemoteCall> IP address is nil")
		return errors.New("Null address for RemoteCall")
	}
	client, err := GetClient(targetNode)
	if err != nil {
		log.Warningln("<RemoteCall> Fail to dial in ", targetNode, " and error is ", err)
		return err
	}
	if client != nil {
		defer client.Close()
	}
	err2 := client.Call(funcClass, input, result)
	if err2 == nil {
		//log.Infoln("<RemoteCall> in ", targetNode, " with ", funcClass, " success!")
		return nil
	} else {
		log.Errorln("<RemoteCall> in ", targetNode, " with ", funcClass, " fail because : ", err2)
		return err2
	}
}

func (this *network) ShutDown() error {
	this.nodePtr.node.QuitSignal <- true
	err := this.lis.Close()
	if err != nil {
		log.Errorln("<ShutDown> Fail to close the network in ", this.nodePtr.node.address)
		return err
	}
	//log.Infoln("<> -", this.nodePtr.node.address, "- network close successfully")
	return nil
}
