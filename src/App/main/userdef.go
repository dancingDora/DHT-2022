package main

import (
	"chord"
)

/* In this file, you should implement function "NewNode" and
 * a struct which implements the interface "dhtNode".
 */

func NewNode(port string) dhtNode {
	// Todo: create a node and then return it.
	NODE := new(chord.Node)
	NODE.Init(port)
	return NODE

}

// Todo: implement a struct which implements the interface "dhtNode".
