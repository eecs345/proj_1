package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	//Buckets []Bucket
	Buckets []*list.List
	Storage map[ID][]byte
	VDOs 	map[ID]VanashingDataObject
	Lock    sync.RWMutex
}

func (ka *Kademlia) UpdateBuckets(contact Contact) {
	distance := ka.NodeID.Xor(contact.NodeID)
	index := distance.PrefixLen()
	index = IDBits - 1 - index
	if index == -1 {
		fmt.Println("it is myself")
		return
	}
	ka.Lock.Lock()
	defer ka.Lock.Unlock()
	if ka.Buckets[index] == nil {
		ka.Buckets[index] = list.New()
		ele := ka.Buckets[index].PushBack(contact)
		fmt.Println(ele.Value.(Contact).NodeID.AsString())
		return
	}
	for e := ka.Buckets[index].Front(); e != nil; e = e.Next() {
		if e.Value.(Contact).NodeID.Compare(contact.NodeID) == 0 {
			ka.Buckets[index].MoveToBack(e)
			return
		}
	}
	if ka.Buckets[index].Len() < k {
		ele := ka.Buckets[index].PushBack(contact)
		fmt.Println(ele.Value.(Contact).NodeID.AsString())
		return
	} else {
		// ping and update
		nodead := ka.Buckets[index].Front().Value.(Contact).Host
		nodept := ka.Buckets[index].Front().Value.(Contact).Port
		//err := ka.DoPing(nodead, nodept)
		dest := ContactToDest(nodead, nodept)
		client, err := rpc.DialHTTP("tcp", dest)
		if err != nil {
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
		} else {
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
	k.VDOs = make(map[ID]VanashingDataObject)
	fmt.Println("NodeID : ", k.NodeID.AsString())
	k.Buckets = make([]*list.List, IDBits)
	k.Storage = make(map[ID][]byte)
	// Set up RPC server
	// NOTE: KademliaCore is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	// rpc.Register(&KademliaCore{k})
	// rpc.HandleHTTP()
	s := rpc.NewServer() // Create a new RPC server
  s.Register(&KademliaCore{k})
  _, port, _ := net.SplitHostPort(laddr) // extract just the port number
  s.HandleHTTP(rpc.DefaultRPCPath+port, rpc.DefaultDebugPath+port) // I'm making a unique RPC path for this instance of Kademlia

	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Print("Listen: ", err)
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
	} else {
		distance := k.NodeID.Xor(nodeId)
		k.Lock.RLock()
		defer k.Lock.RUnlock()
		entry := IDBits - 1 - distance.PrefixLen()
		if k.Buckets[entry] == nil {
			return nil, &NotFoundError{nodeId, "Not found"}
		} else {
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

	// dest := ContactToDest(host, port)
	// client, err := rpc.DialHTTP("tcp", dest)

	port_str := strconv.Itoa(int(port))
	client, err := rpc.DialHTTPPath("tcp", host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)

	if err != nil {
		return "ERR: HTTP Dial failed!"
	}
	//ping := new(kademlia.PingMessage)
	ping := new(PingMessage)
	ping.Sender = k.SelfContact
	ping.MsgID = NewRandomID()
	var pong PongMessage
	err = client.Call("KademliaCore.Ping", ping, &pong)
	if err != nil {
		log.Print("Call: ", err)
		return "ERR: rpc failed!!"
	} else {
		if !pong.MsgID.Equals(ping.MsgID) {
			return "ERR: MsgID does not Match!"
		}
		k.UpdateBuckets(pong.Sender)
		return "OK: " + pong.Sender.NodeID.AsString()
	}
}

func ContactToDest(host net.IP, port uint16) string {
	addr := host.String()
	portnum := strconv.Itoa(int(port))
	dest := addr + ":" + portnum
	return dest
}

func ContactToString(contact Contact) string {
	ret := contact.NodeID.AsString() + "," + contact.Host.String() + "," + strconv.Itoa(int(contact.Port)) + "\n"
	return ret
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//which node should store this file

	// dest := ContactToDest(contact.Host, contact.Port)
	// client, err := rpc.DialHTTP("tcp", dest)
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)

	if err != nil {
		log.Print("Dial:", err)
		return "ERR: HTTP Dial failed!"
	}
	request := new(StoreRequest)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.Key = key
	request.Value = value
	var result StoreResult
	err = client.Call("KademliaCore.Store", request, &result)
	if err != nil {
		log.Print("Call:", err)
		return "ERR: rcp failed "
	} else {
		if !request.MsgID.Equals(result.MsgID) {
			return "ERR: MsgID does Match"
		}
		k.UpdateBuckets(*contact)
		return "OK: "+string(request.Value)
		}
	//return "ERR: Not implemented"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	// dest := ContactToDest(contact.Host, contact.Port)
	// client, err := rpc.DialHTTP("tcp", dest)

	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)

	if err != nil {
		log.Print("Dial:", err)
		return "ERR: HTTP Dial failed!"
	}
	request := new(FindNodeRequest)
	request.MsgID = NewRandomID()
	request.Sender = k.SelfContact
	request.NodeID = searchKey
	var result FindNodeResult
	err = client.Call("KademliaCore.FindNode", request, &result)
	if err != nil {
		log.Print("Call:", err)
		return "ERR: rcp failed "
	} else {
		if !request.MsgID.Equals(result.MsgID) {
			return "ERR: MsgID does Match"
		}
		ret := ""
		for i := 0; i < len(result.Nodes); i++ {
			k.UpdateBuckets(result.Nodes[i])
			ret = ret + ContactToString(result.Nodes[i])
			fmt.Println("Return NodeID : ", result.Nodes[i].NodeID.AsString())
			fmt.Println("       Host : ", result.Nodes[i].Host)
			fmt.Println("       Port : ", result.Nodes[i].Port)
		}
		k.UpdateBuckets(*contact)
		return "OK:\n" + ret
	}
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	k.Lock.RLock()
	defer k.Lock.RUnlock()
	value, ok := k.Storage[searchKey]
	if ok {
		resp := "OK: " + string(value)
		return resp
	} else {
		return "ERR: No Such Key Stored!"
	}
}

// Proj 2
type Shortlist []Con

type Con struct {
	contact  Contact
	distance ID
	active   bool
}

func (k *Kademlia) InitShortlist(id ID, shortlist *Shortlist) {
	distance := k.NodeID.Xor(id)
	index := distance.PrefixLen()
	index = IDBits - 1 - index
	if index == -1 {
		fmt.Println("This is myself!")
		return
	}
	if k.Buckets[index] != nil {
		if k.Buckets[index].Len() >= alpha {
			for e := k.Buckets[index].Front(); len(*shortlist) < alpha; e = e.Next() {
				*shortlist = append(*shortlist, Con{e.Value.(Contact), e.Value.(Contact).NodeID.Xor(id), false})
			}
			return
		}
	}
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	for i := index; len(*shortlist) < alpha && i >= 0; i -= 1 {
		if k.Buckets[i] != nil {
			fmt.Println(k.Buckets[i])
			fmt.Println("length of shortlist = ", len(*shortlist))
			for e := k.Buckets[i].Front(); len(*shortlist) < alpha && e != nil; e = e.Next() {
				*shortlist = append(*shortlist, Con{e.Value.(Contact), e.Value.(Contact).NodeID.Xor(id), false})
			}
			fmt.Println(*shortlist)
		}
	}
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	if len(*shortlist) < alpha {
		for i := index; len(*shortlist) < alpha && i < IDBits; i += 1 {
			if k.Buckets[i] != nil {
				for e := k.Buckets[i].Front(); len(*shortlist) < alpha && e != nil; e = e.Next() {
					*shortlist = append(*shortlist, Con{e.Value.(Contact), e.Value.(Contact).NodeID.Xor(id), false})
				}
			}
		}
		return
	} else {
		return
	}
}

func (k *Kademlia) GetCons(shortlist *Shortlist, num int) []Contact {
	count := 0
	var ret []Contact
	for i := 0; i < len(*shortlist); {
		if !(*shortlist)[i].active {
			ret = append(ret, (*shortlist)[i].contact)
			count += 1
			(*shortlist) = append((*shortlist)[:i], (*shortlist)[i+1:]...)
			if count == num {
				break
			}
		} else {
			i += 1
		}
	}
	return ret
}

func (k *Kademlia) IterFindNode(id ID, contact Contact, retch chan string) {
	res := k.DoFindNode(&contact, id)
	if string(res[0]) == "O" {
		tmp := strings.SplitN(res, "\n", 2)
		active := contact.NodeID.AsString() + "," + contact.Host.String() + "," + strconv.Itoa(int(contact.Port)) + "\n"
		ret := tmp[0] + "\n" + active + tmp[1]
		fmt.Println("the result of one findnode\n", ret)
		retch <- ret
	} else {
		retch <- res
	}
}
func (k *Kademlia) IterFindValue(id ID, contact Contact, retch chan string) {
	res := k.DoFindValue(&contact, id)
	if string(res[0]) == "O" {
		tmp := strings.SplitN(res, "\n", 2)
		active := contact.NodeID.AsString() + "," + contact.Host.String() + "," + strconv.Itoa(int(contact.Port)) + "\n"
		ret := tmp[0] + "\n" + active + tmp[1]
		fmt.Println("the added nodes of one findvalue\n", ret)
		retch <- ret
	} else {
		retch <- res
	}
}
func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	// dest := ContactToDest(contact.Host, contact.Port)
	// client, err := rpc.DialHTTP("tcp", dest)
	port_str := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)

	if err != nil {
		log.Print("Dial:", err)
		return "ERR: HTTP Dial failed!"
	}
	request := new(FindValueRequest)
	request.MsgID = NewRandomID()
	request.Key = CopyID(searchKey)
	request.Sender = k.SelfContact
	var result FindValueResult
	err = client.Call("KademliaCore.FindValue", request, &result)
	if err != nil {
		log.Print("Call:", err)
		return "ERR: rcp failed "
	}
	if !request.MsgID.Equals(result.MsgID) {
		return "ERR: MsgID does Match"
	}
	k.UpdateBuckets(*contact)

	if result.Value == nil {
		ret := ""
		for i := 0; i < len(result.Nodes); i++ {
			k.UpdateBuckets(result.Nodes[i])
			ret = ret + result.Nodes[i].NodeID.AsString() + "," + result.Nodes[i].Host.String() + "," + strconv.Itoa(int(result.Nodes[i].Port)) + "\n"
			fmt.Println("Return NodeID : ", result.Nodes[i].NodeID.AsString())
			fmt.Println("       Host : ", result.Nodes[i].Host)
			fmt.Println("       Port : ", result.Nodes[i].Port)
		}
		return "OK : \n" + ret
	}
	return "Perfect:" + contact.NodeID.AsString() + "," + string(result.Value[3:])
}

func parseResult(result string) []Contact {
	con := make([]Contact, 0)
	if p := strings.Index(result, "OK"); p == 0 {
		A := strings.Split(result, "\n")
		B := strings.Split(A[1], ",")
		var ip net.IP
		var IDd ID
		var port int
		for index, item := range A {
			if index != len(A)-1 {
				if index > 0 {
					B = strings.Split(item, ",")
					for _, it := range B {
						fmt.Println(it)
					}
					IDd, _ = IDFromString(B[0])
					ip = net.ParseIP(B[1])
					port, _ = strconv.Atoi(B[2])
					con = append(con, Contact{IDd, ip, uint16(port)})
				}
			}
		}
	}
	return con
}

// Yuxuan He's work

func UniqueSlice(slice *Shortlist) {
	found := make(map[ID]bool)
	total := 0
	for i, val := range *slice {
		if _, ok := found[val.distance]; !ok {
			found[val.distance] = true
			(*slice)[total] = (*slice)[i]
			total++
		}
	}
	*slice = (*slice)[:total]
}

func (a Shortlist) Len() int { // Overwrite  Len()
	return len(a)
}
func (a Shortlist) Swap(i, j int) { // Overwrite  Swap()
	a[i], a[j] = a[j], a[i]
}
func (a Shortlist) Less(i, j int) bool { // Overwrite  Less()
	return a[i].distance.Less(a[j].distance)
}

func (k *Kademlia) UpdateShortList(contact_list <-chan string, shortlist *Shortlist, id ID, times int) string {
	counter := 0 // 用一个计数器来判断， 是不是三次都接收完毕了
	flag := false
	if times == 0 {
		return "Full"
	}
	var closest_distance ID
	fmt.Println("enter updateshortlist")
Loop:
	for {
		select {
		case contact_string := <-contact_list:
			fmt.Println("receive string from channel")
			if string(contact_string[0]) == "P" {
				return "OK :"+contact_string[8:]
			}
			counter = counter + 1
			new_contact := parseResult(contact_string) // 将string 类型， 转化为 contact slice 类型
			templist := make(Shortlist, 0)
			if new_contact != nil {
				for index, i := range new_contact {
					var temp Con
					if index == 0 { //是自身的contact
						temp.active = true
					} else {
						temp.active = false //是返回的contact
					}
					temp.contact = CopyContact(i)
					temp.distance = CopyID(id.Xor(i.NodeID))
					templist = append(templist, temp)
				}
				if len(*shortlist) == 0 {
					closest_distance = MaxID()
					fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$\nclosest_distance = ", closest_distance.AsString(), "\n$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$\n")
				} else {
					closest_distance = (*shortlist)[0].distance
				}
				(*shortlist) = append(*(shortlist), templist...)
				// Slice 去重
				UniqueSlice(shortlist)
				sort.Sort(Shortlist(*shortlist)) //需要import sort
				if (*shortlist)[0].distance.Less(closest_distance) {
					// closest Node Updated
					closest_distance = (*shortlist)[0].distance
					flag = true
				}
				fmt.Println("counter = ", counter)
			}
			if counter == times { //使用mod 需要import math
				fmt.Println("receive enough data")
				// 接受到了 times 次
				break Loop
			}
		}
	}
	fmt.Println("out for loop")
	// break出来 进行判断处理
	var num_active = 0
	for _, i := range *shortlist {
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		if i.active == true {
			num_active = num_active + 1
		}
	}
	fmt.Println("updateshorlist out")
	if num_active >= 20 {
		return "Full"
	} else if flag == true {
		return "Continue"
	} else {
		return "Another"
	}
}

func (ka *Kademlia) CollectFromShortlist(shortlist *Shortlist) string {
	sllen := len(*shortlist)
	if sllen == 0 {
		return "Find Nothing!"
	}
	ret := ""
	for count, i := 0, 0; i < sllen && count < k; i += 1 {
		if (*shortlist)[i].active == true {
			tmp := (*shortlist)[i].contact
			ret = ret + ContactToString(tmp)
			count += 1
		}
	}
	return ret
}

func (ka *Kademlia) DoIterativeFindNode(id ID) string {
	// For project 2!
	// Initialize the shortlist
	var shortlist Shortlist
	ka.InitShortlist(id, &shortlist)
	stop := false
	ch := make(chan string)
	// While loop
	for !stop {
		alphacons := ka.GetCons(&shortlist, alpha)
		times := len(alphacons)
		fmt.Println(times, " parallel RPCs")
		for i := 0; i < times; i += 1 {
			go ka.IterFindNode(id, alphacons[i], ch)
		}
		fmt.Println("before update shortlist")
		// Update shortlist
		signal := ka.UpdateShortList(ch, &shortlist, id, times)

		//continue or stop
		fmt.Println(signal)

		switch signal {
		case "Full":
			stop = true
		case "Another":
			for signal == "Another" {
				kcons := ka.GetCons(&shortlist, k)
				times := len(kcons)
				for i := 0; i < times; i += 1 {
					go ka.IterFindNode(id, kcons[i], ch)
				}
				signal = ka.UpdateShortList(ch, &shortlist, id, times)
			}
		case "Continue":
		}
	}
	ret := ka.CollectFromShortlist(&shortlist)
	return "OK: " + ret
}

func (k *Kademlia) DoIterativeStore(key ID, value []byte) string {
	// For project 2!
	nodes_string := k.DoIterativeFindNode(key)
	contacts := parseResult(nodes_string)
	for _, i := range contacts {
		k.DoStore(&i, key, value)
	}
	return "OK: "+contacts[len(contacts)-1].NodeID.AsString()

	//return "ERR: Not implemented"
}

func (ka *Kademlia) DoIterativeFindValue(id ID) string {
	// For project 2!
	test := ka.LocalFindValue(id)
	if string(test[0]) == "O" {
		res := ka.NodeID.AsString() + " , " + test[3:]
		return res
	}
	var shortlist Shortlist
	ka.InitShortlist(id, &shortlist)
	stop := false
	ch := make(chan string)
	// While loop
	for !stop {
		alphacons := ka.GetCons(&shortlist, alpha)
		times := len(alphacons)
		fmt.Println(times, " parallel RPCs")
		for i := 0; i < times; i += 1 {
			go ka.IterFindValue(id, alphacons[i], ch)
		}
		fmt.Println("before update shortlist")
		// Update shortlist
		signal := ka.UpdateShortList(ch, &shortlist, id, times)

		//continue or stop
		fmt.Println(signal)
		switch signal {
		case "Full":
			stop = true
		case "Another":
			for signal == "Another" {
				kcons := ka.GetCons(&shortlist, k)
				times := len(kcons)
				for i := 0; i < times; i += 1 {
					go ka.IterFindValue(id, kcons[i], ch)
				}
				signal = ka.UpdateShortList(ch, &shortlist, id, times)
				if string(signal[0]) == "O"{
					ka.DoStore(&(shortlist[0].contact),id,[]byte(signal[46:]))
					return signal
				}
			}
		case "Continue":
		default:
			return signal
		}
	}

	return "ERR: Can not Find It"
}



///////////////proj3
func (k *Kademlia) Vanish(VDOID ID, data []byte, N byte, T byte) string {
	vdodata := VanishData(*k, data, N, T)
	e := k.Storevdo(VDOID,vdodata)
	if e != nil{
		log.Fatal(e)
		return "ERR: Cannot store VDO"
	}
	return "OK: VDO has stored with key " + VDOID.AsString()
}

func (k *Kademlia) Unvanish(NodeID ID, VDOID ID) string {
	c, e := k.FindContact(NodeID)
	if e != nil {
		return "ERR: Not a valid NodeID!"
	}
	var vdo VanashingDataObject
	err := k.DoGetVDO(c.Host, c.Port, VDOID, &vdo)
	if err != nil {
		log.Print(err)
		return "ERR: Cannot get VDO object!"
	}
	data := UnvanishData(*k, vdo)
	if(data == nil) {
		return "ERR: Cannot get T shared keys"
	}
	return "OK: " + string(data)
}

func(k *Kademlia) DoGetVDO(host net.IP, port uint16, VDOID ID, vdo *VanashingDataObject) error {
	// dest := ContactToDest(host, port)
	// client, err := rpc.DialHTTP("tcp", dest)
	port_str := strconv.Itoa(int(port))
	client, err := rpc.DialHTTPPath("tcp", host.String()+":"+port_str,rpc.DefaultRPCPath+port_str)


	if err != nil {
		log.Print(err)
		return err
	}
	req := new(GetVDORequest)
	req.MsgID = NewRandomID()
	req.Sender = k.SelfContact
	req.VdoID = VDOID
	var res GetVDOResult
	err = client.Call("KademliaCore.GetVDO", req, &res)
	if err != nil{
		log.Print("Call: ", err)
		return err
	}
	if !res.MsgID.Equals(req.MsgID) {
		log.Print("MessageID not match")
		return &MyError{"MsgID not Match"}
	}
	k.UpdateBuckets(res.Sender)
	*vdo = res.VDO
	return nil
}

type MyError struct{
	what string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("%s",e.what)
}

func (k *Kademlia) Storevdo(VID ID, vdo VanashingDataObject) error {
	k.Lock.Lock()
	k.VDOs[VID] = vdo
	k.Lock.Unlock()
	return nil
}

func (k *Kademlia) LocalFindVDO(VdoID ID) (bool, VanashingDataObject) {
	k.Lock.RLock()
	defer k.Lock.RUnlock()
	vdo, ok := k.VDOs[VdoID]
	return ok,vdo
}
