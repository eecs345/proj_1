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

// func TestPing(t *testing.T) {
//     instance1 := NewKademlia("localhost:7890")
//     instance2 := NewKademlia("localhost:7891")
//     host2, port2, _ := StringToIpPort("localhost:7891")
//     instance1.DoPing(host2, port2)
//     contact2, err := instance1.FindContact(instance2.NodeID)
//     if err != nil {
//         t.Error("Instance 2's contact not found in Instance 1's contact list")
//         return
//     }
//     contact1, err := instance2.FindContact(instance1.NodeID)
//     if err != nil {
//         t.Error("Instance 1's contact not found in Instance 2's contact list")
//         return
//     }
//     if contact1.NodeID != instance1.NodeID {
//         t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
//     }
//     if contact2.NodeID != instance2.NodeID {
//         t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
//     }
//     return
// }
//
// func TestStore(t *testing.T) {
//   // test Dostore() function and LocalFindValue() function
//   instance1 := NewKademlia("localhost:7892")
//   instance2 := NewKademlia("localhost:7893")
//   host2, port2, _ := StringToIpPort("localhost:7893")
//   instance1.DoPing(host2, port2)
//   contact2, err := instance1.FindContact(instance2.NodeID)
//   if err != nil {
//       t.Error("Instance 2's contact not found in Instance 1's contact list")
//       return
//   }
//   Key := NewRandomID()
//   Value := []byte("Hello World")
//   result := instance1.DoStore(contact2, Key, Value)
//   if p:= strings.Index(result,"ERR");p==0 {
//     t.Error("Can not store this value")
//   }
//   result = instance2.LocalFindValue(Key)
//   //using LocalFindValue to verification the result of DoStore()
//   t.Logf(result)
//   //show the result of "DoStore"
//   if p:= strings.Index(result,"ERR");p==0 {
//     t.Error("Can not find this value")
//   }
//   return
// }
//
// func TestFind_Node(t *testing.T) {
//   // tree structure;
//   // A->B->tree
//   /*
//          C
//       /
//   A-B -- D
//       \
//          E
// */
//   instance1 := NewKademlia("localhost:7894")
//   instance2 := NewKademlia("localhost:7895")
//   host2, port2, _ := StringToIpPort("localhost:7895")
//   instance1.DoPing(host2, port2)
//   contact2, err := instance1.FindContact(instance2.NodeID)
//   if err != nil {
//       t.Error("Instance 2's contact not found in Instance 1's contact list")
//       return
//   }
//
//   tree_node := make([]*Kademlia, 30)
//   for i := 0; i < 30; i++ {
//       address := "localhost:"+strconv.Itoa(7896+i)
//       tree_node[i] = NewKademlia(address)
//       host_number, port_number, _ := StringToIpPort(address)
//       instance2.DoPing(host_number, port_number)
//   }
//   Key := NewRandomID()
//   result := instance1.DoFindNode(contact2, Key)
//   t.Logf(result)
//   //return k nodes
//   if p:= strings.Index(result,"ERR");p==0 {
//     t.Error("Can not find this value")
//   }
//   return
// }
//
// func TestFind_Value(t *testing.T) {
//   // tree structure;
//   // A->B->tree
//   /*
//          C
//       /
//   A-B -- D
//       \
//          E
// */
//   instance1 := NewKademlia("localhost:7926")
//   instance2 := NewKademlia("localhost:7927")
//   host2, port2, _ := StringToIpPort("localhost:7927")
//   instance1.DoPing(host2, port2)
//   contact2, err := instance1.FindContact(instance2.NodeID)
//   if err != nil {
//       t.Error("Instance 2's contact not found in Instance 1's contact list")
//       return
//   }
//
//   tree_node := make([]*Kademlia, 30)
//   for i := 0; i < 30; i++ {
//       address := "localhost:"+strconv.Itoa(7928+i)
//       tree_node[i] = NewKademlia(address)
//       host_number, port_number, _ := StringToIpPort(address)
//       instance2.DoPing(host_number, port_number)
//   }
//
//   Key := NewRandomID()
//   Value := []byte("Hello world")
//   result_store := instance2.DoStore(contact2, Key, Value)
//   if p:= strings.Index(result_store,"ERR");p==0 {
//     t.Error("Can not store this value")
//   }
//   // Given the right keyID, it should return the value
//   result_find := instance1.DoFindValue(contact2, Key)
//   t.Logf(result_find)
//   if p:= strings.Index(result_find,"ERR");p==0 {
//     t.Error("Can not find this value")
//   }
//
//   //Given the wrong keyID, it should return k nodes.
//   Key_wrong := NewRandomID()
//   result_find = instance1.DoFindValue(contact2, Key_wrong)
//   t.Logf(result_find)
//   if p:= strings.Index(result_find,"ERR");p==0 {
//     t.Error("Can not find this value")
//   }
// }

// func TestIterativeFindNode(t *testing.T) {
//   // using line structure;
//   /*
//     A->B->C->D->E->F->G
//   */
//   line_node := make([]*Kademlia, 30)
//   line_node[0] = NewKademlia("localhost:7959")
//   for i := 1; i < 30; i++ {
//       address := "localhost:"+strconv.Itoa(7960+i)
//       t.Logf(address)
//       line_node[i] = NewKademlia(address)
//       host_number, port_number, _ := StringToIpPort(address)
//       line_node[i-1].DoPing(host_number, port_number)
//   }
//   key := NewRandomID()
//   result := line_node[0].DoIterativeFindNode(key)
//   if p:= strings.Index(result,"OK");p!=0 {
//     t.Error("Can't find Node")
//   }
//   t.Logf(result)
// }
//
// func TestIterativeStore(t *testing.T) {
//   // using line structure;
//   /*
//     A->B->C->D->E->F->G
//   */
//   line_node := make([]*Kademlia, 30)
//   line_node[0] = NewKademlia("localhost:7990")
//   for i := 1; i < 30; i++ {
//       address := "localhost:"+strconv.Itoa(7991+i)
//       line_node[i] = NewKademlia(address)
//       host_number, port_number, _ := StringToIpPort(address)
//       line_node[i-1].DoPing(host_number, port_number)
//   }
//   Key := NewRandomID()
//   Value := []byte("Hello world")
//   result := line_node[0].DoIterativeStore(Key,Value)
//   if p:= strings.Index(result,"OK");p!=0 {
//     t.Error("Can't store value")
//   }
//   t.Logf(result)
// }

func TestIterativeFindValue(t *testing.T) {
  // using line structure;
  /*
    A->B->C->D->E->F->G
  */
  line_node := make([]*Kademlia, 80)
  line_node[0] = NewKademlia("localhost:8030")
  for i := 1; i < 20; i++ {
      address := "localhost:"+strconv.Itoa(8031+i)
      line_node[i] = NewKademlia(address)
      host_number, port_number, _ := StringToIpPort(address)
      line_node[i-1].DoPing(host_number, port_number)
  }
  contact, err := line_node[18].FindContact(line_node[19].NodeID)
  if err != nil {
      t.Error("Instance 29's contact not found in Instance 28's contact list")
      return
  }
  Key := NewRandomID()
  Value := []byte("Hello world")

  // given the right key
  result := line_node[19].DoStore(contact,Key, Value)
  if p:= strings.Index(result,"OK");p!=0 {
    t.Error("Can't store value at the end of line")
  }
  result = line_node[0].DoIterativeFindValue(Key)
  if p:= strings.Index(result,"OK");p!=0 {
    t.Error("Can't iterative find value")
  }
  //t.Logf(result)
  //test if the key was wrong
  Key_wrong := NewRandomID()
  result = line_node[0].DoIterativeFindValue(Key_wrong)
  if p:= strings.Index(result,"OK");p==0 {
    t.Error("The key was wrong, it should not return OK")
  }
  t.Logf(result)

  //if there is only two node in this network
  first_node := NewKademlia("localhost:8130")
  second_node := NewKademlia("localhost:8131")
  address_second := "localhost:"+strconv.Itoa(8131)
  host_number_second, port_number_second, _ := StringToIpPort(address_second)
  first_node.DoPing(host_number_second, port_number_second)
  contact_second_node, err := first_node.FindContact(second_node.NodeID)
  if err != nil {
      t.Error("the first node and the second node are not connected")
      return
  }
  first_node.DoStore(contact_second_node, Key, Value)
  result = first_node.DoIterativeFindValue(Key)
  if p:= strings.Index(result,"OK");p!=0 {
    t.Error("Can't iterative find value")
  }
  //t.Logf(result)

}
