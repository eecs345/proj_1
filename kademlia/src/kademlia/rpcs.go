package kademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
	"fmt"
	"sort"
	"log"
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
	//fmt.Println("Ping from : ", ping.Sender.NodeID.AsString())
	pong.MsgID = CopyID(ping.MsgID)
    // Specify the sender
	pong.Sender = kc.kademlia.SelfContact
	// Update contact, etc
	kc.kademlia.UpdateBuckets(ping.Sender)
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
	kc.kademlia.Lock.Lock()
	kc.kademlia.Storage[req.Key] = req.Value
	kc.kademlia.Lock.Unlock()
	kc.kademlia.UpdateBuckets(req.Sender)
	fmt.Println(kc.kademlia.NodeID.AsString()," : ",kc.kademlia.Storage)
	kc.kademlia.UpdateBuckets(req.Sender)
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

type Closest_Node struct{
	distance ID
	contact Contact
}

type NodeSlice []Closest_Node

func (kc *KademliaCore) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	res.MsgID = CopyID(req.MsgID)
	dis := kc.kademlia.NodeID.Xor(req.NodeID)
	BucketIndex := dis.PrefixLen()
	BucketIndex = IDBits - 1 - BucketIndex
	NodeList := make(NodeSlice,0)
	if BucketIndex != IDBits - 1 {
		BucketIndex += 1
	}
	kc.kademlia.Lock.RLock()
	i := 0
	for ; i <= BucketIndex; i++ {
		if kc.kademlia.Buckets[i] != nil{
			for j := kc.kademlia.Buckets[i].Front(); j != nil; j = j.Next() {
				if !j.Value.(Contact).NodeID.Equals(req.Sender.NodeID){
					var tmp Closest_Node
					tmp.distance = j.Value.(Contact).NodeID.Xor(req.NodeID)
					tmp.contact = CopyContact(j.Value.(Contact))
					NodeList = append(NodeList, tmp)
				}
			}
		}
	}
	if (len(NodeList) < k){
		for ; len(NodeList) <=  k && i <= IDBits - 1; i++ {
			if kc.kademlia.Buckets[i] != nil{
				for j := kc.kademlia.Buckets[i].Front(); j != nil; j = j.Next() {
					if !j.Value.(Contact).NodeID.Equals(req.Sender.NodeID){
						var tmp Closest_Node
						tmp.distance = j.Value.(Contact).NodeID.Xor(req.NodeID)
						tmp.contact = CopyContact(j.Value.(Contact))
						NodeList = append(NodeList, tmp)
					}
				}
			}
		}
	}
	kc.kademlia.Lock.RUnlock()
	l := len(NodeList)
	if l <= k {
		//return the contacts
		for i := 0; i < l; i++ {
			res.Nodes = append(res.Nodes, CopyContact(NodeList[i].contact))
		} 
	}else{
		//sort the contacts and return
		sort.Sort(NodeSlice(NodeList))
		for i := 0; i < k; i++ {
			res.Nodes[i] = CopyContact(NodeList[i].contact)
		}
	}
	kc.kademlia.UpdateBuckets(req.Sender)
	return nil
}

func (a NodeSlice) Len() int {    // Overwrite  Len() 
	return len(a)
}
func (a NodeSlice) Swap(i, j int){     // Overwrite  Swap() 
	a[i], a[j] = a[j], a[i]
}
func (a NodeSlice) Less(i, j int) bool {    // Overwrite  Less() 
	return a[i].distance.Less(a[j].distance)
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
	data := kc.kademlia.LocalFindValue(req.Key)
	if  string(data[0]) == "O" {
		res.Value = []byte(data)
		kc.kademlia.UpdateBuckets(req.Sender)
	}else {
		res.Value = nil
		request := new(FindNodeRequest)
		request.Sender = req.Sender
		request.NodeID = CopyID(req.Key)
		var result FindNodeResult
		err := kc.FindNode(*request, &result)
		if err != nil{
			log.Fatal("Find Node : ",err)
		}
		for i := 0; i < len(result.Nodes); i++{
			res.Nodes = append(res.Nodes,result.Nodes[i])
		}
	}
	return nil
}
