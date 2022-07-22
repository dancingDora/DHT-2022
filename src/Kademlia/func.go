package Kademlia

import (
	"container/list"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const ()

type database struct {
	rwLock        sync.RWMutex
	dataset       map[string]string
	expireTime    map[string]time.Time
	duplicateTime map[string]time.Time
	republicTime  map[string]time.Time
	privilege     map[string]int
}

func (this *database) init() {
	this.dataset = make(map[string]string)
	this.expireTime = make(map[string]time.Time)
	this.duplicateTime = make(map[string]time.Time)
	this.republicTime = make(map[string]time.Time)
	this.privilege = make(map[string]int)
}



func (this *database) store(request StoreRequest) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	if _, ok := this.dataset[request.Key]; !ok {
		requestPri := request.RequesterPri
		this.privilege[request.Key] = request.RequesterPri
		this.dataset[request.Key] = request.Value
		if requestPri == 0 {
			this.republicTime[request.Key] = time.Now().Add(republicTimeInterval)
			return
		}
		if requestPri == 1 {
			this.duplicateTime[request.Key] = time.Now().Add(duplicateTimeInterval)
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			return
		}
		if requestPri == 2 {
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			return
		}
		log.Errorf("<database store> Wrong privilege")
	} else {
		originPri := this.privilege[request.Key]
		requestPri := request.RequesterPri
		if originPri == 1 || requestPri >= originPri {
			return
		}
		if requestPri == 0 {
			this.privilege[request.Key] = 1
			this.dataset[request.Key] = request.Value
			this.republicTime[request.Key] = time.Now().Add(republicTimeInterval)
			delete(this.expireTime, request.Key)
			delete(this.duplicateTime, request.Key)
			return
		} else if requestPri == 1 {
			this.privilege[request.Key] = 2
			this.dataset[request.Key] = request.Value
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			this.duplicateTime[request.Key] = time.Now().Add(duplicateTimeInterval)
			delete(this.republicTime, request.Key)
			return
		} else if requestPri == 2 {
			this.privilege[request.Key] = 3
			this.dataset[request.Key] = request.Value
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			delete(this.duplicateTime, request.Key)
			delete(this.republicTime, request.Key)
			return
		}
		log.Errorln("<database store> Wrong privilege2")
	}
}

func (this *database) expire() {
	tmp := make(map[string]bool)
	this.rwLock.RLock()
	for k, v := range this.expireTime {
		if !v.After(time.Now()) {
			tmp[k] = true
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range tmp {
		delete(this.dataset, k)
		delete(this.expireTime, k)
		delete(this.duplicateTime, k)
		delete(this.republicTime, k)
		delete(this.privilege, k)
	}
	this.rwLock.Unlock()
}

func (this *database) duplicate() (result map[string]string) {
	result = make(map[string]string)
	this.rwLock.RLock()
	for k, v := range this.duplicateTime {
		if !v.After(time.Now()) {
			result[k] = this.dataset[k]
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range result {
		this.duplicateTime[k] = time.Now().Add(duplicateTimeInterval)
	}
	this.rwLock.Unlock()
	return
}

func (this *database) republic() (result map[string]string){
	result = make(map[string]string)
	this.rwLock.RLock()
	for k, v := range this.republicTime {
		if !v.After(time.Now()) {
			result[k] = this.dataset[k]
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range result {
		this.republicTime[k] = time.Now().Add(republicTimeInterval)
	}
	this.rwLock.Unlock()
	return
}

func (this *database) get(key string) (bool, string) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	if v, ok := this.dataset[key]; ok {
		if _, ok2 := this.expireTime[key]; ok2 {
			if !this.expireTime[key].After(time.Now().Add(expireTimeInterval3)) {
				this.expireTime[key] = time.Now().Add(expireTimeInterval3)
			}
		}
		return true, v
	} else {
		return false, ""
	}
}

type RoutingTable struct {
	nodeAddr       Contact
	rwLock         sync.RWMutex
	buckets        [160]*list.List
	refreshIndex   int
	refreshTimeSet [160]time.Time
}

func (this *RoutingTable) InitRoutingTable(nodeAddr Contact) {
	this.nodeAddr = nodeAddr
	this.rwLock.Lock()
	for i := 0; i < 160; i++ {
		this.buckets[i] = list.New()
		this.refreshTimeSet[i] = time.Now()
	}
	this.refreshIndex = 0
	this.rwLock.Unlock()
}

func (this *RoutingTable) Update(contact *Contact) {
	this.rwLock.RLock()
	bucket := this.buckets[PrefixLen(Xor(this.nodeAddr.NodeID, contact.NodeID))]
	target := bucket.Front()
	target = nil
	for i := bucket.Front(); ; i = i.Next() {
		if i == nil {
			target = nil
			break
		}
		if i.Value.(*Contact).NodeID.Equals(contact.NodeID) {
			target = i
			break
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	if target

}