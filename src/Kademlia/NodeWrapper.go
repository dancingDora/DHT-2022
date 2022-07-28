package kademlia

type WrapNode struct {
	node *KadNode
}

func (this *WrapNode) FindNode(input *FindNodeArg, res *ClosestList) error {
	*res = this.node.FindNode(&input.TarID)
	this.node.kBucketUpdate(input.Sender)
	return nil
}

func (this *WrapNode) AddPair(input *StoreArg, _ *string) error {
	this.node.data.AddPair(input.Key, input.Value)
	this.node.kBucketUpdate(input.Sender)
	return nil
}

func (this *WrapNode) FindValue(input *FindValueArg, res *FindValueRet) error {
	*res = this.node.FindValue(input.Key, &input.Hash)
	this.node.kBucketUpdate(input.Sender)
	return nil
}
