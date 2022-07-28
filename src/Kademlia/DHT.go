package kademlia

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
	"time"
)

var localAddres string

type Pair struct {
	Key   string
	Value string
}

type Addr struct {
	Ip string
	Id big.Int
}

func (this *Addr) addr_init(port int) {
	this.Ip = fmt.Sprintf("%s:%d", localAddress, port)
	this.Id = Hash(this.Ip)
}

type Data struct {
	hashMap       map[string]string
	validTime     map[string]time.Time
	republishTime map[string]time.Time
	lock          sync.RWMutex
}

func (this *Data) Init() {
	this.hashMap = make(map[string]string)
	this.validTime = make(map[string]time.Time)
	this.republishTime = make(map[string]time.Time)
}

type ClosestList struct {
	Size     int
	Standard big.Int
	List     [K]Addr
}

type KBucket struct {
	size   int
	bucket [K]Addr
	mux    sync.Mutex
}

type Node struct {
	address        Addr
	data           Data
	station        *network
	conRoutineFlag bool
	routeTable     [M]KBucket
	mux            sync.RWMutex
}

type FindNodeArg struct {
	TarID  big.Int
	Sender Addr
}

type FindValueArg struct {
	Key    string
	Hash   big.Int
	Sender Addr
}

type StoreArg struct {
	Key    string
	Value  string
	Sender Addr
}

type FindValueRet struct {
	First  ClosestList
	Second string
}

func (this *Node) Init(port int) {
	this.address.addr_init(port)
	this.reset()
}

func (this *Node) Run() {
	this.station = new(network)
	tmp_err := this.station.Init(this.address.Ip, this)
	if tmp_err != nil {
		log.Errorln("[Run error] can not init station, the node IP is : ", this.address.Ip)
		return
	} else {
		log.Infoln("[Run success] in : ", this.address.Ip)
		this.conRoutineFlag = true
		go this.RePublish()
	}
}

func (this *Node) Join(ip string) bool {
	tmpAddr := Addr{ip, Hash(ip)}
	this.kBucketUpdate(tmpAddr)
	client, tmp_err := Diag(ip)
	if tmp_err != nil {
		log.Errorln("[Diag error] in ", ip)
	} else {
		var res ClosestList
		tmp_err = client.Call("WrapNode.FindNode", &FindNodeArg{this.address.Id, this.address}, &res)
		for i := 0; i < res.Size; i++ {
			this.kBucketUpdate(res.List[i])
		}
		client.Close()
	}

	closestlist := this.NodeLookup(&this.address.Id)
	for i := 0; i < closestlist.Size; i++ {
		this.kBucketUpdate(closestlist.List[i])
		client, tmp_err = Diag(closestlist.List[i].Ip)
		if tmp_err != nil {
			log.Errorln("[Diag error] in function Join diag error in", closestlist.List[i].Ip)
		} else {
			var res ClosestList
			tmp_err = client.Call("WrapNode.FindNode", &FindNodeArg{this.address.Id, this.address}, &res)
			if tmp_err != nil {
				log.Errorln("[Error] remotecall FindNode in Join error", this.address.Ip, "because", tmp_err)
			}
			for j := 0; j < res.Size; j++ {
				this.kBucketUpdate(res.List[j])
			}
			client.Close()
		}
	}
	return true
}

func (this *Node) Ping(addr string) bool {
	isOnline := CheckOnline(addr)
	return isOnline
}

func (this *Node) Put(key string, value string) bool {
	keyID := Hash(key)
	closestList := this.NodeLookup(&keyID)
	closestList.Insert(this.address)
	for i := 0; i < closestList.Size; i++ {
		client, tmp_err := Diag(closestList.List[i].Ip)
		if tmp_err != nil {
			log.Errorln("[Error] in function Put can not diag node aimIp", closestList.List[i].Ip)
		} else {
			var o string
			tmp_err = client.Call("WrapNode.AddPair", &StoreArg{key, value, this.address}, &o)
			if tmp_err != nil {
				log.Errorln("[Error] can not call addpair because", tmp_err)
			}
			client.Close()
		}
	}
	return true
}

func (this *Node) Get(key string) (bool, string) {
	keyID := Hash(key)
	isDiaged := make(map[string]bool)
	finfValueRes := this.FindValue(key, &keyID)
	if finfValueRes.Second != "" {
		return true, finfValueRes.Second
	}
	closestlist := finfValueRes.First
	isUpdated := true
	for isUpdated {
		isUpdated = false
		var tmp ClosestList
		var removeList []Addr
		for i := 0; i < closestlist.Size; i++ {
			if isDiaged[closestlist.List[i].Ip] == true {
				continue
			}
			client, tmp_err := Diag(closestlist.List[i].Ip)
			isDiaged[closestlist.List[i].Ip] = true
			var res FindValueRet
			if tmp_err != nil {
				log.Errorln("[Error] in function get can not diag", closestlist.List[i].Ip, "because", tmp_err)
				removeList = append(removeList, closestlist.List[i])
			} else {
				tmp_err = client.Call("WrapNode.FindValue", &FindValueArg{Key: key, Sender: this.address}, &res)
				if res.Second != "" {
					return true, res.Second
				} else {
					for j := 0; j < res.First.Size; j++ {
						tmp.Insert(res.First.List[j])
					}
				}
				client.Close()
			}
		}
		for _, rmKey := range removeList {
			closestlist.Remove(rmKey)
		}
		for i := 0; i < tmp.Size; i++ {
			isUpdated = isUpdated || closestlist.Insert(tmp.List[i])
		}
	}
	secList := this.NodeLookup(&keyID)
	for _, aimAddr := range secList.List {
		client, tmp_err := Diag(aimAddr.Ip)
		if tmp_err != nil {
			log.Errorln("[Error] in function Get can not diag", aimAddr.Ip, "because", tmp_err)
		}
		defer client.Close()
		var res FindValueRet
		client.Call("WrapNode.FindValue", &FindValueArg{Key: key, Sender: this.address}, &res)
		if res.Second != "" {
			return true, res.Second
		}
	}
	return false, ""
}

func (this *Node) FindNode(tarID *big.Int) (closestList ClosestList) {
	this.mux.RLock()
	defer this.mux.RUnlock()
	closestList.Standard = *tarID
	for i := 0; i < M; i++ {
		for j := 0; j < this.routeTable[i].size; j++ {
			if Ping(this.routeTable[i].bucket[j].Ip) == nil { // if online
				closestList.Insert(this.routeTable[i].bucket[j])
			}
		}
	}
	return
}

func (this *Node) FindValue(key string, hash *big.Int) FindValueRet {
	this.mux.RLock()
	defer this.mux.RUnlock()
	founded, value := this.data.GetValue(key)
	if founded {
		return FindValueRet{ClosestList{}, value}
	}
	retClosest := ClosestList{Standard: *hash}
	for i := 0; i < M; i++ {
		for j := 0; j < this.routeTable[i].size; j++ {
			if Ping(this.routeTable[i].bucket[j].Ip) == nil { //if online
				retClosest.Insert(this.routeTable[i].bucket[j])
			}
		}
	}
	return FindValueRet{retClosest, ""}
}

func (this *Node) NodeLookup(tarID *big.Int) (closestList ClosestList) {
	closestList = this.FindNode(tarID)
	closestList.Insert(this.address)
	isUpdate := true
	diaged := make(map[string]bool)
	for isUpdate {
		isUpdate = false
		var tmp ClosestList
		var removeList []Addr
		for i := 0; i < closestList.Size; i++ {
			if diaged[closestList.List[i].Ip] == true {
				continue
			}
			this.kBucketUpdate(closestList.List[i])
			client, tmp_err := Diag(closestList.List[i].Ip)
			diaged[closestList.List[i].Ip] = true
			var res ClosestList
			//remove the offline node
			if tmp_err != nil {
				removeList = append(removeList, closestList.List[i])
			} else {
				tmp_err = client.Call("WrapNode.FindNode", &FindNodeArg{TarID: *tarID, Sender: this.address}, &res)
				for j := 0; j < res.Size; j++ {
					tmp.Insert(res.List[j])
				}
				client.Close()
			}
		}
		for _, key := range removeList {
			closestList.Remove(key)
		}
		for i := 0; i < tmp.Size; i++ {
			//todo:can be simplified
			isUpdate = isUpdate || closestList.Insert(tmp.List[i])
		}
	}
	return
}

func (this *Node) RePublish() {
	for this.conRoutineFlag {
		for i := 0; i < M; i++ {
			this.routeTable[i].Reflesh()
		}
		this.mux.Lock()
		copyData := this.data.CopyData()
		republishList := this.data.GetRePublishList()
		this.mux.Unlock()
		for _, key := range republishList {
			this.Put(key, copyData[key])
		}
		this.data.DeleteExpiredData()
		time.Sleep(RepublishINterval)
	}
}

func (this *Node) reset() {
	this.conRoutineFlag = false
	this.data.Init()
}

func (this *Node) kBucketUpdate(addr Addr) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if addr.Ip == "" || addr.Ip == this.address.Ip {
		return
	}
	this.routeTable[cpl(&this.address.Id, &addr.Id)].Update(addr)
}

func (this *Node) Create() {
	return
}
func (this *Node) Quit() {
	this.station.ShutDown()
	this.reset()
}

func (this *Node) ForceQuit() {
	this.station.ShutDown()
	this.reset()
}

func (this *Node) Delete(key string) bool {
	return true
}
