package kademlia

import (
    "testing"
    "net"
    "strconv"
    "strings"
)

func StringToIpPort(laddr string) (ip net.IP, port uint16, err error){
    hostString, portString, err := net.SplitHostPort(laddr)
    if err != nil {
        return
    }
    ipStr, err := net.LookupHost(hostString)
    if err != nil {
        return
    }
    for i := 0; i < len(ipStr); i++ {
        ip = net.ParseIP(ipStr[i])
        if ip.To4() != nil {
            break
        }
    }
    portInt, err := strconv.Atoi(portString)
    port = uint16(portInt)
    return
}

func TestPing(t *testing.T) {
    instance1 := NewKademlia("localhost:7890")
    instance2 := NewKademlia("localhost:7891")
    host2, port2, _ := StringToIpPort("localhost:7891")
    instance1.DoPing(host2, port2)
    contact2, err := instance1.FindContact(instance2.NodeID)
    if err != nil {
        t.Error("Instance 2's contact not found in Instance 1's contact list")
        return
    }
    contact1, err := instance2.FindContact(instance1.NodeID)
    if err != nil {
        t.Error("Instance 1's contact not found in Instance 2's contact list")
        return
    }
    if contact1.NodeID != instance1.NodeID {
        t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
    }
    if contact2.NodeID != instance2.NodeID {
        t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
    }
    return
}

func TestStore(t *testing.T) {
  // test Dostore() function and LocalFindValue() function
  instance1 := NewKademlia("localhost:7892")
  instance2 := NewKademlia("localhost:7893")
  host2, port2, _ := StringToIpPort("localhost:7893")
  instance1.DoPing(host2, port2)
  contact2, err := instance1.FindContact(instance2.NodeID)
  if err != nil {
      t.Error("Instance 2's contact not found in Instance 1's contact list")
      return
  }
  Key := NewRandomID()
  Value := []byte("Hello World")
  result := instance1.DoStore(contact2, Key, Value)
  if p:= strings.Index(result,"ERR");p==0 {
    t.Error("Can not store this value")
  }
  result = instance2.LocalFindValue(Key)
  t.Logf(result)
  //show the result of "DoStore"
  if p:= strings.Index(result,"ERR");p==0 {
    t.Error("Can not find this value")
  }
  return
}

func TestFind_Node(t *testing.T) {
  // tree structure;
  // A->B->tree
  instance1 := NewKademlia("localhost:7894")
  instance2 := NewKademlia("localhost:7895")
  host2, port2, _ := StringToIpPort("localhost:7895")
  instance1.DoPing(host2, port2)
  contact2, err := instance1.FindContact(instance2.NodeID)
  if err != nil {
      t.Error("Instance 2's contact not found in Instance 1's contact list")
      return
  }

  tree_node := make([]*Kademlia, 30)
  for i := 0; i < 30; i++ {
      address := "localhost:"+strconv.Itoa(7896+i)
      tree_node[i] = NewKademlia(address)
      host_number, port_number, _ := StringToIpPort(address)
      instance2.DoPing(host_number, port_number)
  }


  Key := NewRandomID()
  result := instance1.DoFindNode(contact2, Key)
  t.Logf(result)
  if p:= strings.Index(result,"ERR");p==0 {
    t.Error("Can not find this value")
  }
  return
}


func TestFind_Value(t *testing.T) {
  instance1 := NewKademlia("localhost:7926")
  instance2 := NewKademlia("localhost:7927")
  host2, port2, _ := StringToIpPort("localhost:7927")
  instance1.DoPing(host2, port2)
  contact2, err := instance1.FindContact(instance2.NodeID)
  if err != nil {
      t.Error("Instance 2's contact not found in Instance 1's contact list")
      return
  }

  tree_node := make([]*Kademlia, 30)
  for i := 0; i < 30; i++ {
      address := "localhost:"+strconv.Itoa(7928+i)
      tree_node[i] = NewKademlia(address)
      host_number, port_number, _ := StringToIpPort(address)
      instance2.DoPing(host_number, port_number)
  }

  Key := NewRandomID()
  Value := []byte("Hello world")
  result_store := instance2.DoStore(contact2, Key, Value)
  if p:= strings.Index(result_store,"ERR");p==0 {
    t.Error("Can not store this value")
  }
  t.Logf(result_store)
  result_find := instance1.DoFindValue(contact2, Key)
  if p:= strings.Index(result_find,"ERR");p==0 {
    t.Error("Can not find this value")
  }
  t.Logf(result_find)
  if (result_store != result_find) {
    t.Error("Find the wrong value")
  }


}
