package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
    "strconv"
	"container/list"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID ID
    SelfContact Contact
	//Buckets []Bucket
	Buckets []*list.List
	Storage map[ID][]byte
}


func (ka *Kademlia)UpdateBuckets(contact Contact){
	distance := ka.NodeID.Xor(contact.NodeID)
	index := distance.PrefixLen()
	index = 159 - index
	if (index == -1){
		fmt.Println("it is myself")
		return
	}
	if (ka.Buckets[index] == nil){
		ka.Buckets[index] = list.New()
		ele := ka.Buckets[index].PushBack(contact)
		fmt.Println(*ele)
		return
	}
	for e := ka.Buckets[index].Front(); e != nil; e = e.Next(){
		if (e.Value.(Contact).NodeID.Compare(contact.NodeID) == 0){
			if (ka.Buckets[index].Len() < k){
				ka.Buckets[index].MoveToBack(e)
				return
			}else {
				// ping and update
				nodead := ka.Buckets[index].Front().Value.(Contact).Host
				nodept := ka.Buckets[index].Front().Value.(Contact).Port
				err := ka.DoPing(nodead, nodept)
				if string(err[0]) == "O"{
					ka.Buckets[index].MoveToBack(e)
				}else {
					ka.Buckets[index].Remove(e)
					ka.Buckets[index].PushBack(contact)
				}
				return
			}
		}
	}
	ka.Buckets[index].PushBack(contact)
	return
}

//type Bucket *list.List

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	k.Buckets = make([]*list.List, IDBits)
	k.Storage = make(map[ID][]byte)
	fmt.Println(k.Buckets[0])
	// Set up RPC server
	// NOTE: KademliaCore is just a wrapper around Kademlia. This type includes
	// the RPC functions.
	rpc.Register(&KademliaCore{k})
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	// Run RPC server forever.
	go http.Serve(l, nil)

    // Add self contact
    hostname, port, _ := net.SplitHostPort(l.Addr().String())
    port_int, _ := strconv.Atoi(port)
    ipAddrStrings, err := net.LookupHost(hostname)
    var host net.IP
    for i := 0; i < len(ipAddrStrings); i++ {
        host = net.ParseIP(ipAddrStrings[i])
        if host.To4() != nil {
            break
        }
    }
    k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}
	return k
}

type NotFoundError struct {
	id  ID
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId.Compare(k.SelfContact.NodeID)==0 {
			return &k.SelfContact, nil
	}	else{
		distance :=k.NodeID.Xor(nodeId)
		entry:=159-distance.PrefixLen()
	if k.Buckets[entry]==nil{
		return nil, &NotFoundError{nodeId, "Not found"}
	}else{
		for e := k.Buckets[entry].Front(); e != nil; e = e.Next() {
			if e.Value.(Contact).NodeID.Compare(nodeId)==0{
				return e.Value.(*Contact), nil
			}
		}
	}
	}
return nil, &NotFoundError{nodeId, "Not found"}
}

// This is the function to perform the RPC
func (k *Kademlia) DoPing(host net.IP, port uint16) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	dest := ContactToDest(host,port)
	client, err := rpc.DialHTTP("tcp", dest)
	if (err != nil){
		return "ERR: HTTP Dial failed!"
	}
	//ping := new(kademlia.PingMessage)
	ping := new(PingMessage)
	ping.Sender = k.SelfContact
	ping.MsgID = NewRandomID()
	var pong PongMessage
	err = client.Call("KademliaCore.Ping", ping, &pong)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR: rpc failed!!"
	}else {
		if(!pong.MsgID.Equals(ping.MsgID)){
			return "ERR: MsgID does not Match!"
		}
		k.UpdateBuckets(pong.Sender)
		return "OK: "
	}
}

func ContactToDest(host net.IP, port uint16) string{
	addr := host.String()
	portnum := strconv.Itoa(int(port))
	dest := addr + ":" +portnum
	return dest
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//which node should store this file
	dest := ContactToDest(contact.Host, contact.Port)
	client, err := rpc.DialHTTP("tcp", dest)
	if (err != nil){
		log.Fatal("Dial:",err)
		return "ERR: HTTP Dial failed!"
	}
	request := new(StoreRequest)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.Key = key
	request.Value = value
	var result StoreResult
	err = client.Call("KademliaCore.Store",request,&result)
	if (err != nil){
		log.Fatal("Call:",err)
		return "ERR: rcp failed "
	}else{
		if (!request.MsgID.Equals(result.MsgID)){
			return "ERR: MsgID does Match"
		}
		k.UpdateBuckets(*contact)
		return "OK:"
	}
	//return "ERR: Not implemented"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	dest := ContactToDest(contact.Host, contact.Port)
	client, err := rpc.DialHTTP("tcp", dest)
	if (err != nil){
		log.Fatal("Dial:",err)
		return "ERR: HTTP Dial failed!"
	}
	request := new(FindNodeRequest)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.NodeID = searchKey
	var result FindNodeResult
	err = client.Call("KademliaCore.FindNode",request,&result)
	if (err != nil){
		log.Fatal("Call:",err)
		return "ERR: rcp failed "
	}else{
		if (!request.MsgID.Equals(result.MsgID)) {
			return "ERR: MsgID does Match"
		}
		k.UpdateBuckets(*contact)
		return "OK:"
	}
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) DoIterativeFindNode(id ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeFindValue(key ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
