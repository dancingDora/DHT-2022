package chord

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
	"time"
)

const (
	SuccessorListSize = 5
	HashBitsSize      = 160
)

type Node struct {
	address        string
	ID             *big.Int
	conRoutineFlag bool
	QuitSignal     chan bool
	SuccessorList  [SuccessorListSize]string
	predecessor    string
	fingerTable    [HashBitsSize]string

	rwLock     sync.RWMutex
	dataSet    map[string]string
	dataLock   sync.RWMutex
	backupSet  map[string]string
	backupLock sync.RWMutex

	station *network

	next int
}

func (this *Node) Init(port string) {
	this.address = fmt.Sprintf("%s:%d", localAddress, port)
	this.ID = ConsistentHash(this.address)
	this.conRoutineFlag = false
	this.reset()
}

func (this *Node) reset() {
	this.dataLock.Lock()
	this.dataSet = make(map[string]string)
	this.dataLock.Unlock()
	this.QuitSignal = make(chan bool, 2)
	this.backupLock.Lock()
	this.backupLock.Unlock()
	this.rwLock.Lock()
	this.next = 1
	this.rwLock.Unlock()
}

func (this *Node) Run() {
	this.station = new(network)
	err := this.station.Init(this.address, this)
	if err != nil {
		log.Errorln("<Run> failed ", err)
		return
	}
	this.conRoutineFlag = true
	this.next = 1
}

func (this *Node) Create() {
	this.predecessor = ""
	this.SuccessorList[0] = this.address
	this.fingerTable[0] = this.address
	this.background()
}

func (this *Node) background() {
	go func() {
		for this.conRoutineFlag {
			this.stabilize()
			time.Sleep(timeCut)
		}
	}()
	go func() {
		for this.conRoutineFlag {
			this.check_predecessor()
			time.Sleep(timeCut)
		}
	}()
	go func() {
		for this.conRoutineFlag {
			this.fix_finger()
			time.Sleep(timeCut)
		}
	}()
}

func (this *Node) first_online_successor(result *string) error {
	for i := 0; i < SuccessorListSize; i++ {
		if CheckOnline(this.SuccessorList[i]) {
			*result = this.SuccessorList[i]
			return nil
		}
	}
	log.Errorln("<first_online_successor> List Break in ", this.address)
	return errors.New("List Break")
}

func (this *Node) stabilize() {
	var sucPreAddr string
	var newSucAddr string
	this.first_online_successor(&newSucAddr)
	err := RemoteCall(newSucAddr, "NodeWrapper.GetPredecessor", 2022, &sucPreAddr)

	if err != nil {
		log.Errorln("<stabilize> fail to get predecessor in ", newSucAddr)
		return
	}
	if sucPreAddr != "" && contain(ConsistentHash(sucPreAddr), this.ID, ConsistentHash(newSucAddr), false) {
		newSucAddr = sucPreAddr
	}
	var temp [SuccessorListSize]string
	err = RemoteCall(newSucAddr, "NodeWrapper.SetSuccessorList", 2022, &temp)
	if err != nil {
		log.Errorln("<stabilize> Fail to stabilize ,Can't get successorList: ", err)
		return
	}
	this.rwLock.Lock()
	this.SuccessorList[0] = newSucAddr
	this.fingerTable[0] = newSucAddr
	for i := 1; i < SuccessorListSize; i++ {
		this.SuccessorList[i] = temp[i-1]
	}
	this.rwLock.Unlock()
	var occupy string
	err = RemoteCall(newSucAddr, "NodeWrapper.Notify", this.address, &occupy)
	if err != nil {
		log.Errorln("<stabilize> Fail to notify ,Can't get call successor : ", err)
		return
	}
	log.Infoln("<stabilize> successfully in", this.address)
}

func (this *Node) apply_backup() error {
	this.backupLock.RLock()
	this.dataLock.Lock()
	for k, v := range this.backupSet {
		this.dataSet[k] = v
	}
	this.dataLock.Unlock()
	this.backupLock.RUnlock()
	var sucAddr string
	err := this.first_online_successor(&sucAddr)
	if err != nil {
		log.Errorln("<apply_backup> Fail to Find Successor in ", this.address)
		return err
	}
	var occupy string
	err = RemoteCall(sucAddr, "NodeWrapper.AddBackup", this.backupSet, &occupy)
	if err != nil {
		log.Errorln("<apply_backup> Fail to Update ", this.address, "'s SUccessor's Backup becaise ", err)
		return err
	}
	this.backupLock.Lock()
	this.backupSet = make(map[string]string)
	this.backupLock.Unlock()
	log.Infoln("<apply_backup> Successfully apply backup in ", this.address)
	return nil
}

func (this *Node) check_predecessor() error {
	if this.predecessor != "" && !CheckOnline(this.predecessor) {
		this.rwLock.Lock()
		this.predecessor = ""
		this.rwLock.Unlock()
		this.apply_backup()
		log.Infoln("<check_predecessor> Find failed predecessor")
	}
	return nil
}

func (this *Node) closest_preceding_node(target *big.Int) string {
	for i := HashBitsSize - 1; i >= 0; i-- {
		if this.fingerTable[i] == "" {
			continue
		}
		if !CheckOnline(this.fingerTable[i]) {
			continue
		}
		if contain(ConsistentHash(this.fingerTable[i]), this.ID, target, false) {
			log.Infoln("<closest_preceding_node>")
		}
	}
	var preaddr string
	err := this.first_online_successor(&preaddr)
	if err != nil {
		log.Errorln("<closest_preceding_node> List Break")
		return ""
	}
	return preaddr
}

func (this *Node) find_successor(target *big.Int, result *string) error {
	var sucAddr string
	this.first_online_successor((&sucAddr))
	if contain(target, this.ID, ConsistentHash(sucAddr), true) {
		*result = sucAddr
		return nil
	}
	closestPre := this.closest_preceding_node(target)
	return RemoteCall(closestPre, "NodeWrapper.FindSuccessor", target, result)
}

func (this *Node) fix_finger() {
	var ans string
	err := this.find_successor(calculateID(this.ID, this.next), &ans)
	if err != nil {
		//	log.Errorln("<fix_finger> error occurs")
		return
	}
	this.rwLock.Lock()
	this.fingerTable[0] = this.SuccessorList[0]
	this.fingerTable[this.next] = ans
	this.next++
	if this.next >= HashBitsSize {
		this.next = 1
	}
	this.rwLock.Unlock()
}

func (this *Node) Join(addr string) bool {
	if isOnline := CheckOnline(addr); !isOnline {
		//log.Warningln("<Join> Node in ", this.address, " fail to join network in ", addr, " for the network is failed")
		return false
	}
	var sucAddr string
	err := RemoteCall(addr, "NodeWrapper.FindSuccessor", this.ID, &sucAddr)
	if err != nil {
		log.Errorln("<Join> Fail to Join ,Error msg: ", err)
		return false
	}
	var temp [SuccessorListSize]string
	err = RemoteCall(sucAddr, "NodeWrapper.SetSuccessorList", 2022, &temp)
	if err != nil {
		log.Errorln("<Join> Fail to Join ,Can't get successorList: ", err)
		return false
	}
	this.rwLock.Lock()
	this.predecessor = ""
	this.SuccessorList[0] = sucAddr
	this.fingerTable[0] = sucAddr
	for i := 1; i < SuccessorListSize; i++ {
		this.SuccessorList[i] = temp[i-1]
	}
	this.rwLock.Unlock()
	err = RemoteCall(sucAddr, "NodeWrapper.HereditaryData", this.address, &this.dataSet)
	if err != nil {
		log.Errorln("<Join> Fail to Join ,Can't share Data and Backup: ", err)
		return false
	}
	this.background()
	return true
}

func (this *Node) Quit() {
	if !this.conRoutineFlag {
		return
	}
	log.Warningln("<Quit> fail to quit in ", this.address)
	err := this.station.ShutDown()
	if err != nil {
		log.Errorln("<Quit> fail to quit in ", this.address)
	}

	this.rwLock.Lock()
	this.conRoutineFlag = false
	this.rwLock.Unlock()

	var sucAddr string
	var occupy string
	this.first_online_successor(&sucAddr)
	err = RemoteCall(sucAddr, "NodeWrapper.CheckPredecessor", 2022, &occupy)
	if err != nil {
		log.Errorln("<Quit.CheckPredecessor> Error : ", err)
	}
	err = RemoteCall(this.predecessor, "NodeWrapper.Stabilize", 2021, &occupy)
	if err != nil {
		log.Errorln("<Quit.Stabilize> Error : ", err)
	}
	this.reset()
}

func (this *Node) ForceQuit() {
	err := this.station.ShutDown()
	if err != nil {
		log.Errorln("<Quit> fail to quit in ", this.address)
	}
	this.rwLock.Lock()
	this.conRoutineFlag = false
	this.rwLock.Unlock()
	this.reset()
}

func (this *Node) Ping(addr string) bool {
	return CheckOnline(addr)
}

func (this *Node) Put(key string, value string) bool {
	if !this.conRoutineFlag {
		log.Errorln("<Put> The Node is sleeping : ", this.address)
		return false
	}
	var targetAddr string
	err := this.find_successor(ConsistentHash(key), &targetAddr)
	if err != nil {
		log.Errorln("<Put> Fail to Find Key's Successor")
		return false
	}
	var occupy string
	var dataPair Pair = Pair{key, value}
	err = RemoteCall(targetAddr, "NodeWrapper.StoreData", dataPair, &occupy)
	if err != nil {
		log.Errorln("<Put> Put Data into ", targetAddr, " successfully")
		return false
	}
	return true
}

func (this *Node) Get(key string) (bool, string) {
	if !this.conRoutineFlag {
		log.Errorln("<Get> The node is sleeping : ", this.address)
		return false, ""
	}
	var value string
	var targetAddr string
	err := this.find_successor(ConsistentHash(key), &targetAddr)
	if err != nil {
		log.Errorln("<Get> Fail to Get Data from Target Node in ", targetAddr, " error : ", err)
		return false, ""
	}
	err = RemoteCall(targetAddr, "NodeWrapper.GetData", key, &value)
	return true, value
}

func (this *Node) Delete(key string) bool {
	if !this.conRoutineFlag {
		log.Errorln("<Delete> The node is sleeping : ", this.address)
		return false
	}
	var targetAddr string
	err := this.find_successor(ConsistentHash(key), &targetAddr)
	if err != nil {
		log.Errorln("<Delete> Fail to Find Key's Successor")
		return false
	}
	var occupy string
	err = RemoteCall(targetAddr, "NodeWrapper.DeleteData", key, &occupy)
	if err != nil {
		log.Errorln("<Delete> Fail to Delete Data from Target Node in ", targetAddr, " error : ", err)
		return false
	}
	return true
}

func (this *Node) set_successor_list(result *[SuccessorListSize]string) error {
	this.rwLock.RLock()
	*result = this.SuccessorList
	this.rwLock.RUnlock()
	return nil
}

func (this *Node) get_predecessor(result *string) error {
	this.rwLock.RLock()
	*result = this.predecessor
	this.rwLock.RUnlock()
	return nil
}

func (this *Node) notify(instructor string) error {
	if this.predecessor == instructor {
		return nil
	}
	if this.predecessor == "" || contain(ConsistentHash(instructor), ConsistentHash(this.predecessor), this.ID, false) {
		this.rwLock.Lock()
		this.predecessor = instructor
		this.rwLock.Unlock()
		var occupy string
		err := RemoteCall(instructor, "NodeWrapper.GenerateBackup", &occupy, &this.backupSet)
		if err != nil {
			log.Errorln("<notify> Fail to get backup from predecessor ", instructor, " because ", err)
		}
	}
	return nil
}

func (this *Node) store_data(dataPair Pair) error {
	this.dataLock.Lock()
	this.dataSet[dataPair.Key] = dataPair.Value
	this.dataLock.Unlock()
	var sucAddr string
	this.first_online_successor(&sucAddr)
	var occupy string
	err := RemoteCall(sucAddr, "NodeWrapper.StoreBackup", dataPair, &occupy)
	if err != nil {
		log.Errorln("<store_data>Fail to store backup in ", this.address)
	}
	return err
}

func (this *Node) delete_data(key string) error {
	this.dataLock.Lock()
	_, ok := this.dataSet[key]
	if ok {
		delete(this.dataSet, key)
	}
	this.dataLock.Unlock()
	if ok {
		var sucAddr string
		this.first_online_successor(&sucAddr)
		var occupy string
		err := RemoteCall(sucAddr, "NodeWrapper.DeleteBackup", key, &occupy)
		if err != nil {
			log.Errorln("<delete_data> Fail to delete backup in ", this.address)
		}
		return nil
	} else {
		return errors.New("<delete_data> Unreachable Data")
	}
}

func (this *Node) hereditary_data(predeAddr string, dataSet *map[string]string) error {
	this.backupLock.Lock()
	this.dataLock.Lock()
	this.backupSet = make(map[string]string)
	for k, v := range this.dataSet {
		if !contain(ConsistentHash(k), ConsistentHash(predeAddr), this.ID, true) {
			(*dataSet)[k] = v
			this.backupSet[k] = v
			delete(this.dataSet, k)
		}
	}
	this.backupLock.Unlock()
	this.dataLock.Unlock()
	var occupy string
	var sucAddr string
	this.first_online_successor(&sucAddr)
	err := RemoteCall(sucAddr, "NodeWrapper.SubBackup", *dataSet, &occupy)
	if err != nil {
		log.Errorln("<hereditary_data> fail to reduce successors' backup because : ", err)
	}
	this.rwLock.Lock()
	this.predecessor = predeAddr
	this.rwLock.Unlock()
	return nil
}

func (this *Node) inherit_data(dataSet *map[string]string) error {
	this.dataLock.Lock()
	for k, v := range *dataSet {
		this.dataSet[k] = v
	}
	this.dataLock.Unlock()
	return nil
}

func (this *Node) store_backup(dataPair Pair) error {
	this.backupLock.Lock()
	this.backupSet[dataPair.Key] = dataPair.Value
	this.backupLock.Unlock()
	return nil
}

func (this *Node) delete_backup(key string) error {
	this.backupLock.Lock()
	_, ok := this.backupSet[key]
	if ok {
		delete(this.backupSet, key)
		this.backupLock.Unlock()
	}
	if ok {
		return nil
	} else {
		return errors.New("<delete_backup> Unreachable Data")
	}
}

func (this *Node) add_backup(dataSet map[string]string) error {
	this.backupLock.Lock()
	for k, v := range dataSet {
		this.backupSet[k] = v
	}
	this.backupLock.Unlock()
	return nil
}

func (this *Node) sub_backup(dataSet map[string]string) error {
	this.backupLock.Lock()
	for k, _ := range dataSet {
		delete(this.backupSet, k)
	}
	this.backupLock.Unlock()
	return nil
}

func (this *Node) generate_backup(sucBackup *map[string]string) error {
	this.dataLock.RLock()
	*sucBackup = make(map[string]string)
	for k, v := range this.dataSet {
		(*sucBackup)[k] = v
	}
	this.dataLock.RUnlock()
	return nil
}

func (this *Node) get_data(key string, value *string) error {
	this.dataLock.RLock()
	tmp, ok := this.dataSet[key]
	this.dataLock.RUnlock()
	if ok {
		*value = tmp
		return nil
	} else {
		*value = ""
		return errors.New("<get_data> Unreachable Data")
	}
}
