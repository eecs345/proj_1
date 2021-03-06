package kademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
	"container/list"
)

type KademliaCore struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (kc *KademliaCore) Ping(ping PingMessage, pong *PongMessage) error {
	// TODO: Finish implementation
	pong.MsgID = CopyID(ping.MsgID)
	pong.Sender = kc.kademlia.SelfContact
	kc.kademlia.Update(ping.Sender)
    // Specify the sender
	// Update contact, etc
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (kc *KademliaCore) Store(req StoreRequest, res *StoreResult) error {
	// TODO: Implement.

	res.MsgID = CopyID(req.MsgID)
	kc.kademlia.HashTable[req.Key] = req.Value
	res.Err = nil
	kc.kademlia.Update(req.Sender)
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	key:=req.NodeID
	res.MsgID= CopyID(req.MsgID)
	kc.kademlia.Update(req.Sender)
	k:= 20
	res.Nodes=make([]Contact,k)
	res.Err=nil
	distance :=kc.kademlia.NodeID.Xor(key)
	entry:=159-distance.PrefixLen()
	i:=entry
	j:=entry
	judge:=true
	
	temp:=list.new()
	temp.PushBackList(&kc.kademlia.Bucket[entry])
	for counter := 0; counter < kc.kademlia.Bucket[entry].Len(); counter++ {
		res.Nodes[counter]=temp.front().Value.(Contact)
		temp.MoveToBack(temp.front())
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindValue(req FindValueRequest, res *FindValueResult) error {
	// TODO: Implement.
	res.MsgID = CopyID(req.MsgID)
	value,ok := kc.kademlia.HashTable[req.Key]
	if ok == false {
		req_node:=new(FindNodeRequest)
		req_node.Sender=req.Sender
		req_node.MsgID= req.MsgID
		req_node.NodeID = req.Key
		var res_node FindNodeResult

		kc.FindNode(*req_node,&res_node)
		res.Nodes = res_node.Nodes
	} else {
		res.Value = value
	}





	res.Err = nil
	kc.kademlia.Update(req.Sender)
	return nil
}
