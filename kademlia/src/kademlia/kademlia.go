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
	"sync"
	"math"
	"strings"
	"encoding/hex"
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
	Lock sync.RWMutex
}


func (ka *Kademlia)UpdateBuckets(contact Contact){
	distance := ka.NodeID.Xor(contact.NodeID)
	index := distance.PrefixLen()
	index = IDBits - 1 - index
	if (index == -1){
		fmt.Println("it is myself")
		return
	}
	ka.Lock.Lock()
	defer ka.Lock.Unlock()
	if (ka.Buckets[index] == nil){
		ka.Buckets[index] = list.New()
		ele := ka.Buckets[index].PushBack(contact)
		fmt.Println(ele.Value.(Contact).NodeID.AsString())
		return
	}
	for e := ka.Buckets[index].Front(); e != nil; e = e.Next(){
		if e.Value.(Contact).NodeID.Compare(contact.NodeID) == 0 {
				ka.Buckets[index].MoveToBack(e)
				return
			}
	}
	if ka.Buckets[index].Len() < k {
		ele := ka.Buckets[index].PushBack(contact)
		fmt.Println(ele.Value.(Contact).NodeID.AsString())
		return
	}else{
			// ping and update
			nodead := ka.Buckets[index].Front().Value.(Contact).Host
			nodept := ka.Buckets[index].Front().Value.(Contact).Port
			//err := ka.DoPing(nodead, nodept)
			dest := ContactToDest(nodead, nodept)
			client, err := rpc.DialHTTP("tcp", dest)
			if (err != nil){
				f := ka.Buckets[index].Front()
				ka.Buckets[index].Remove(f)
				ka.Buckets[index].PushBack(contact)
				return
			}
			ping := new(PingMessage)
			ping.Sender = ka.SelfContact
			ping.MsgID = NewRandomID()
			var pong PongMessage
			err = client.Call("KademliaCore.Ping", ping, &pong)
			if err != nil {
				f := ka.Buckets[index].Front()
				ka.Buckets[index].Remove(f)
				ka.Buckets[index].PushBack(contact)
				return
			}else {
				if !pong.MsgID.Equals(ping.MsgID) {
					f := ka.Buckets[index].Front()
					ka.Buckets[index].Remove(f)
					ka.Buckets[index].PushBack(contact)
					return
				}
				f := ka.Buckets[index].Front()
				ka.Buckets[index].MoveToBack(f)
			}
			return
		}
	return
}

//type Bucket *list.List

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	fmt.Println("NodeID : ",k.NodeID.AsString())
	k.Buckets = make([]*list.List, IDBits)
	k.Storage = make(map[ID][]byte)
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
	if nodeId.Compare(k.SelfContact.NodeID) == 0 {
		return &k.SelfContact, nil
	}else {
		distance := k.NodeID.Xor(nodeId)
		k.Lock.RLock()
		defer k.Lock.RUnlock()
		entry := IDBits - 1 - distance.PrefixLen()
		if k.Buckets[entry] == nil{
			return nil, &NotFoundError{nodeId, "Not found"}
		}else {
			for e := k.Buckets[entry].Front(); e != nil; e = e.Next() {
				if e.Value.(Contact).NodeID.Compare(nodeId) == 0 {
					tmp := e.Value.(Contact)
					return &tmp, nil
				}
			}
		}
	}
	return nil, &NotFoundError{nodeId, "Not found"}
}


//func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
//	// TODO: Search through contacts, find specified ID
//	// Find contact with provided ID
//    if nodeId == k.SelfContact.NodeID {
//        return &k.SelfContact, nil
//   }
//	return nil, &NotFoundError{nodeId, "Not found"}
//}

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
		if !pong.MsgID.Equals(ping.MsgID) {
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
		if !request.MsgID.Equals(result.MsgID) {
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
	err = client.Call("KademliaCore.FindNode", request, &result)
	if (err != nil){
		log.Fatal("Call:", err)
		return "ERR: rcp failed "
	}else{
		if !request.MsgID.Equals(result.MsgID) {
			return "ERR: MsgID does Match"
		}
		for i := 0; i < len(result.Nodes); i++ {
			fmt.Println("Return NodeID : ", result.Nodes[i].NodeID.AsString())
			fmt.Println("       Host : ", result.Nodes[i].Host)
			fmt.Println("       Port : ", result.Nodes[i].Port)
		}
		k.UpdateBuckets(*contact)
		return "OK:"
	}
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	dest := ContactToDest(contact.Host,contact.Port)
	client, err := rpc.DialHTTP("tcp", dest)
	if (err != nil){
		log.Fatal("Dial:",err)
		return "ERR: HTTP Dial failed!"
	}
	request := new(FindValueRequest)
	request.MsgID = NewRandomID()
	request.Key = CopyID(searchKey)
	request.Sender = k.SelfContact
	var result FindValueResult
	err = client.Call("KademliaCore.FindValue", request, &result)
	if (err != nil){
		log.Fatal("Call:", err)
		return "ERR: rcp failed "
	}
		if !request.MsgID.Equals(result.MsgID) {
			return "ERR: MsgID does Match"
		}
		k.UpdateBuckets(*contact)
		if result.Value == nil {
			for i := 0; i < len(result.Nodes); i++ {
				fmt.Println("Return NodeID : ", result.Nodes[i].NodeID.AsString())
				fmt.Println("       Host : ", result.Nodes[i].Host)
				fmt.Println("       Port : ", result.Nodes[i].Port)
				return "OK : k-Contacts returned!"
			}
		}
	return string(result.Value)
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	k.Lock.RLock()
	defer k.Lock.RUnlock()
	value, ok := k.Storage[searchKey]
	if ok {
		resp := "OK:" + string(value)
		return resp
	} else {
		return "ERR: No Such Key Stored!"
	}
}
func (k *Kademlia) parseResult(result string)[]Contact
{
	con := make([]Contact, 1)
	if result[0]=="O" {
		A :=strings.Split(result,"\n")
		B := strings.Split(A[1],",")
		var id []byte
		var NID ID
		var ip net.IP
		var port int
		for index,item := range A {
			if index>0{
				B=strings.Split(item,",")
				id=hex.DecodeString(B[0])
				NID=CopyID(id)
				ip=ip.ParseIP(B[1])
				port, _ = strconv.Atoi(B[2])
				con=append(con,new(Contact{NID,ip,uint16(port)}))
			}
		}
		
	}
	return con
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




// Yuxuan He's work



type Shortlist []Con

type Con struct {
	contact Contact
	distance  ID
	active bool
}


func UniqueSlice(slice *Shortlist) {
    found := make(map[Shortlist]bool)
    total := 0
    for i, val := range *slice {
        if _, ok := found[val]; !ok {
            found[val] = true
            (*slice)[total] = (*slice)[i]
            total++
        }
    }
    *slice = (*slice)[:total]
}


func (a Shortlist) Len() int {    // Overwrite  Len()
	return len(a)
}
func (a Shortlist) Swap(i, j int){     // Overwrite  Swap()
	a[i], a[j] = a[j], a[i]
}
func (a Shortlist) Less(i, j int) bool {    // Overwrite  Less()
	return a[i].distance.Less(a[j].distance)
}


func UpdateShortList(contact_list <-chan string, shortlist *Shortlist,id ID) string {
		counter := 0							// 用一个计数器来判断， 是不是三次都接收完毕了
		flag := false
		for {
			select {
			 	case contact_string := <-contactlist:
					counter = counter + 1
					new_contact := parse_string(contact_string)   // 将string 类型， 转化为 contact slice 类型
					templist := make(Shortlist,0)
					if new_contact != nil {
						for i=new_contact.Front();i!=nil;i=i.Next() {
							var temp Con
							if (i==new_contact.Front())   //是自身的contact
							{
								temp.active = true
							} else {
								temp.active = false				//是返回的contact
							}
							temp.contact = CopyContact(i)
							temp.distance = CopyID(id.Xor(i.NodeID))
							templist.append(templist, temp)
						}
						closest_distance = (*shortlist)[0].distance
						(*shortlist).append(*(shortlist), templist)
						// Slice 去重
						UniqueSlice(shortlist)
						sort.Sort(Shortlist(*shortlist)) //需要import sort




						if ((*shortlist)[0].distance < closest_distance) {
							// closest Node Updated
							flag = true
						}
						if math.Mod(counter,3)==0 {   //使用mod 需要import math
							// 接受到了3 次
							break;
						}

					}
			}

		}

		// break出来 进行判断处理
		var num_active = 0
		for i=(*shorlist).Front();i!=nil;i=i.Next() {
			if i.active == true {
				num_active = num_active + 1
			}
		}

		if num_active >= 20 {
			return "Full"
		} else if flag = true {
			return "Continue"
		} else {
			return "Another"
		}


}
