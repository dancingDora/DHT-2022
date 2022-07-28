package main

import (
	"Kademlia"
)

/* In this file, you should implement function "NewNode" and
 * a struct which implements the interface "dhtNode".
 */

func NewNode(port int) dhtNode {
	// Todo: create a node and then return it.
	NODE := new(kademlia.KadNode)
	NODE.Init(port)
	return NODE

}

// Todo: implement a struct which implements the interface "dhtNode".
