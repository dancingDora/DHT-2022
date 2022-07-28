package kademlia

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"time"
)

type network struct {
	serv       *rpc.Server
	lis        net.Listener
	nodePtr    *WrapNode
	QuitSignal chan bool
}

func Accept(ser *rpc.Server, lis net.Listener, ptr *KadNode) {
	for {
		//always run
		conn, tmp_err := lis.Accept()
		select {
		case <-ptr.station.QuitSignal:
			return
		default:
			if tmp_err != nil {
				log.Print("rpc.Serve: accept:", tmp_err.Error())
				return
			}
			go ser.ServeConn(conn)
		}
	}
}

func (this *network) Init(address string, ptr *KadNode) error {
	this.serv = rpc.NewServer()
	this.nodePtr = new(WrapNode)
	this.nodePtr.node = ptr
	this.QuitSignal = make(chan bool, 2)
	//register rpc service
	tmp_err := this.serv.Register(this.nodePtr)
	if tmp_err != nil {
		log.Errorf("[error] register rpc service error!")
		return tmp_err
	}
	//for tcp listen
	this.lis, tmp_err = net.Listen("tcp", address)
	if tmp_err != nil {
		log.Errorf("[error] tcp error!")
		return tmp_err
	}
	go Accept(this.serv, this.lis, this.nodePtr.node)
	return nil
}

func GetClient(addr string) (*rpc.Client, error) {
	var res *rpc.Client
	var err error
	errCh := make(chan error)
	for i := 0; i < 5; i++ {
		go func() {
			res, err = rpc.Dial("tcp", addr)
			errCh <- err
		}()
		select {
		case <-errCh:
			if err == nil {
				return res, err
			} else {
				return nil, err
			}
		case <-time.After(waitTime):
			log.Errorln("In function GetClient time out in" + addr)
			err = errors.New("Time out!")
		}
	}
	return nil, err
}

func CheckOnline(addr string) bool {
	if addr == "" {
		log.Warningln("In checkonline the addr is nil")
		return false
	}
	cli, _ := GetClient(addr)
	if cli != nil {
		defer cli.Close()
		return true
	} else {
		return false
	}
}

func (this *network) ShutDown() error {
	this.QuitSignal <- true
	tmp_err := this.lis.Close()
	if tmp_err != nil {
		log.Errorln("ShutDown error")
	}
	return tmp_err
}
