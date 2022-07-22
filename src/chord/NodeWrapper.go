package chord

import "math/big"

type NodeWrapper struct {
	node *Node
}

func (this *NodeWrapper) FindSuccessor(target *big.Int, result *string) error {
	return this.node.find_successor(target, result)
}

func (this *NodeWrapper) SetSuccessorList(_ int, result *[SuccessorListSize]string) error {
	return this.node.set_successor_list(result)
}

func (this *NodeWrapper) GetPredecessor(_ int, result *string) error {
	return this.node.get_predecessor(result)
}

func (this *NodeWrapper) Notify(instructor string, _ *string) error {
	return this.node.notify(instructor)
}

func (this *NodeWrapper) CheckPredecessor(_ int, _ *string) error {
	return this.node.check_predecessor()
}

func (this *NodeWrapper) Stabilize(_ int, _ *string) error {
	this.node.stabilize()
	return nil
}

func (this *NodeWrapper) StoreData(dataPair Pair, _ *string) error {
	return this.node.store_data(dataPair)
}

func (this *NodeWrapper) GetData(key string, value *string) error {
	return this.node.get_data(key, value)
}

func (this *NodeWrapper) DeleteData(key string, _ *string) error {
	return this.node.delete_data(key)
}

func (this *NodeWrapper) HereditaryData(preAddr string, dataSet *map[string]string) error {
	return this.node.hereditary_data(preAddr, dataSet)
}

func (this *NodeWrapper) InheritData(dataSet *map[string]string, _ *string) error {
	return this.node.inherit_data(dataSet)
}

func (this *NodeWrapper) StoreBackup(dataPair Pair, _ *string) error {
	return this.node.store_backup(dataPair)
}

func (this *NodeWrapper) DeleteBackup(key string, _ *string) error {
	return this.node.delete_backup(key)
}

func (this *NodeWrapper) AddBackup(dataSet map[string]string, _ *string) error {
	return this.node.add_backup(dataSet)
}

func (this *NodeWrapper) SubBackup(dataSet map[string]string, _ *string) error {
	return this.node.sub_backup(dataSet)
}

func (this *NodeWrapper) GenerateBackup(_ *string, dataSet *map[string]string) error {
	return this.node.generate_backup(dataSet)
}
