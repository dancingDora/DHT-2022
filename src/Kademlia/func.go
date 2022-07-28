package kademlia

import (
	"crypto/sha1"
	"errors"
	"math/big"
	"net"
	"net/rpc"
	"time"
)

const M int = 160
const K int = 15
const Alpha int = 3
const ExpiredTime = 960 * time.Second
const NeedRepublicTime = 120 * time.Second
const waitTime = 250 * time.Millisecond
const UpdateInterval = 25 * time.Millisecond
const RepublishINterval = 100 * time.Millisecond
const RemoteTryTime = 3
const RemoteTryInterval = 25 * time.Millisecond

var Mod = big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(M)), nil)
var localAddress string

func init() {
	localAddress = GetLocalAddress()
}

func GetLocalAddress() string {
	var localaddress string

	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}

	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}

	return localaddress
}

func Hash(str string) big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	var ret big.Int
	ret.SetBytes(hasher.Sum(nil))
	ret.Mod(&ret, Mod)
	return ret
}

//for kBucketTYpe
func (this *KBucketType) Reflesh() {
	for i := 0; i < this.size; i++ {
		if Ping(this.bucket[i].Ip) != nil {
			for j := i + 1; j < this.size; j++ {
				this.bucket[j-1] = this.bucket[j]
			}
			this.size--
			return
		}
	}
}

func (this *KBucketType) Update(addr AddrType) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if addr.Ip == "" {
		return
	}
	pos := -1
	for i := 0; i < this.size; i++ {
		if this.bucket[i].Ip == addr.Ip {
			pos = i
			break
		}
	}
	if pos == -1 {
		if this.size < K {
			this.size++
			this.bucket[this.size-1] = addr
			return
		} else {
			if Ping(this.bucket[0].Ip) == nil {
				//bucket[0] is online
				head := this.bucket[0]
				for i := 1; i < K; i++ {
					this.bucket[i-1] = this.bucket[i]
				}
				this.bucket[K-1] = head
				return
			} else {
				//bucket[0] is offline
				for i := 1; i < K; i++ {
					this.bucket[i-1] = this.bucket[i]
				}
				this.bucket[K-1] = addr
				return
			}
		}
	} else {
		for i := pos + 1; i < this.size; i++ {
			this.bucket[i-1] = this.bucket[i]
		}
		this.bucket[this.size-1] = addr
	}
}

//for closetlist:
func (this *ClosestList) Insert(addr AddrType) bool {
	res := false
	if Ping(addr.Ip) != nil {
		return res
	}
	for i := 0; i < this.Size; i++ {
		if this.List[i].Ip == addr.Ip {
			return res
		}
	}
	newDis := dis(&this.Standard, &addr.Id)
	if this.Size < K {
		res = true
		for i := 0; i < this.Size; i++ {
			tmpDis := dis(&this.List[i].Id, &this.Standard)
			if newDis.Cmp(&tmpDis) < 0 {
				this.Size++
				for j := this.Size - 1; j > i; j-- {
					this.List[j] = this.List[j-1]
				}
				this.List[i] = addr
				return res
			}
		}
		this.Size++
		this.List[this.Size-1] = addr
		return res
	} else {
		for i := 0; i < K; i++ {
			tmpDis := dis(&this.List[i].Id, &this.Standard)
			if newDis.Cmp(&tmpDis) < 0 {
				res = true
				for j := K - 1; j > i; j-- {
					this.List[j] = this.List[j-1]
				}
				this.List[i] = addr
				return res
			}
		}
		return res
	}
}

func (this *ClosestList) Remove(addr AddrType) bool {
	for i := 0; i < this.Size; i++ {
		if this.List[i].Ip == addr.Ip {
			this.Size--
			for j := i; j < this.Size; j++ {
				this.List[j] = this.List[j+1]
			}
			return true
		}
	}
	return false
}

func Ping(addr string) error {
	var ret *rpc.Client
	var err error
	if addr == "" {
		return errors.New("[error] Empty IP addr")
	}
	for i := 0; i < RemoteTryTime; i++ {
		ret, err = rpc.Dial("tcp", addr)
		if err == nil {
			ret.Close()
			return err
		}
		time.Sleep(RemoteTryInterval)
	}
	return err
}

func Diag(addr string) (*rpc.Client, error) {
	var ret *rpc.Client
	var err error
	if addr == "" {
		return nil, errors.New("ERROR: empty IP addr")
	}
	for i := 0; i < RemoteTryTime; i++ {
		ret, err = rpc.Dial("tcp", addr)
		if err == nil {
			return ret, err
		}
		time.Sleep(RemoteTryInterval)
	}
	return nil, err
}

//for big.Int
func dis(x, y *big.Int) big.Int {
	var res big.Int
	res.Xor(x, y)
	return res
}
func cpl(x, y *big.Int) int {
	Xdis := dis(x, y)
	return Xdis.BitLen() - 1
}

//for dataType
func (this *DataType) GetRePublishList() (republishList []string) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	for key, tim := range this.republishTime {
		if time.Now().After(tim) {
			republishList = append(republishList, key)
			this.republishTime[key] = time.Now().Add(NeedRepublicTime)
		}
	}
	return republishList
}

func (this *DataType) DeleteExpiredData() {
	var expiredList []string
	this.lock.RLock()
	for key, tim := range this.validTime {
		if time.Now().After(tim) {
			expiredList = append(expiredList, key)
		}
	}
	this.lock.RUnlock()
	this.lock.Lock()
	for _, key := range expiredList {
		delete(this.hashMap, key)
		delete(this.validTime, key)
		delete(this.republishTime, key)
	}
	this.lock.Unlock()
}

func (this *DataType) AddPair(key string, value string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.hashMap[key] = value
	this.validTime[key] = time.Now().Add(ExpiredTime)
	this.republishTime[key] = time.Now().Add(NeedRepublicTime)
}

func (this *DataType) GetValue(key string) (founded bool, res string) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	res, founded = this.hashMap[key]
	return founded, res
}

func (this *DataType) SubPair(key string) (founded bool) {
	this.lock.Lock()
	this.lock.Unlock()
	_, founded = this.hashMap[key]
	if founded {
		delete(this.hashMap, key)
		delete(this.republishTime, key)
		delete(this.validTime, key)
	}
	return founded
}

func (this *DataType) CopyData() map[string]string {
	res := make(map[string]string)
	this.lock.Lock()
	defer this.lock.Unlock()
	for key, value := range this.hashMap {
		res[key] = value
	}
	return res
}
