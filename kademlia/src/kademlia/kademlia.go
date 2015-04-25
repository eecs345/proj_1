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
    bucket []*list.List
}


func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	fmt.Println("hello")
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	fmt.Println("....")

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
    k.bucket= make([]*list.List,160)
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

func (k *Kademlia) Update(cc Contact){
	flag:=0
	distance :=k.NodeID.Xor(cc.NodeID)
	entry:=159-distance.PrefixLen()
	if entry == -1 {
		fmt.Println("This is myself")
		return
	}
	if k.bucket[entry]==nil{
		k.bucket[entry]=list.New()
		k.bucket[entry].PushBack(cc)
	} else{
			for e := k.bucket[entry].Front(); e != nil; e = e.Next() {
	// do something with e.Value
				if e.Value.(Contact).NodeID.Compare(cc.NodeID)==0 {
					k.bucket[entry].MoveToBack(e)
					flag=1
				}

			}
		if flag==0{

			if k.bucket[entry].Len()<20 {
				k.bucket[entry].PushBack(cc)
			} else {
				target := k.bucket[entry].Front()
				err :=k.DoPing(target.Value.(Contact).Host, target.Value.(Contact).Port)
				if err[0] != 'E' {
					k.bucket[entry].MoveToBack(target)
				} else {
					k.bucket[entry].Remove(target)
					k.bucket[entry].PushBack(cc)
				}
			}
		}
		}
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
    if nodeId == k.SelfContact.NodeID {
        return &k.SelfContact, nil
    }	else{
    	distance :=k.NodeID.Xor(nodeId)
			entry:=159-distance.PrefixLen()
		if k.bucket[entry]==nil{
			return nil, &NotFoundError{nodeId, "Not found"}
		}else{
			for e := k.bucket[entry].Front(); e != nil; e = e.Next() {
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

	firstPeerStr := host.String()+":"+ strconv.Itoa(int(port))
	client, err := rpc.DialHTTP("tcp", firstPeerStr)
	if err != nil {
		log.Fatal("DialHTTP: ", err)
		return "ERR: Not implemented"
	}
	ping := new(PingMessage)
	ping.MsgID = NewRandomID()
	ping.Sender = k.SelfContact
	var pong PongMessage
	err = client.Call("KademliaCore.Ping", ping, &pong)
	if err != nil {
		log.Fatal("Call: ", err)
		return "ERR: Not implemented"
	}else{
		k.Update(pong.Sender)
		return "OK: It's good"
	}
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
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
