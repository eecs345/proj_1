package kademlia

import (
    "testing"
    "net"
    "strconv"

    "fmt"
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
    fmt.Println("Normal:"+ instance2.NodeID.AsString())
    // fmt.Println("ours:" + contact2.NodeID.AsString())
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
//   Value := []byte("Hello world")
//   result := instance1.DoStore(contact2, Key, Value)
//   if p:= strings.Index(result,"ERR");p==0 {
//     t.Error("Can not store this value")
//   }
//   result = instance2.LocalFindValue(Key)
//   if p:= strings.Index(result,"ERR");p==0 {
//     t.Error("Can not find this value")
//   }

//
//   return
// }
//


//   return
// }

// func TestFind_Node(t *testing.T) {
//   instance1 := NewKademlia("localhost:7894")
//   instance2 := NewKademlia("localhost:7895")
//   host2, port2, _ := StringToIpPort("localhost:7895")
//   instance1.DoPing(host2, port2)

//



//   contact2, err := instance1.FindContact(instance2.NodeID)
//   if err != nil {
//       t.Error("Instance 2's contact not found in Instance 1's contact list")
//       return
//   }
//   Key := NewRandomID()
//   result := instance1.DoFindNode(contact2, Key)
//   if p:= strings.Index(result,"ERR");p==0 {
//     t.Error("Can not store this value")
//   }
// }

//



//   func TestFind_Value(t *testing.T) {
//     instance1 := NewKademlia("localhost:7896")
//     instance2 := NewKademlia("localhost:7897")
//     host2, port2, _ := StringToIpPort("localhost:7897")
//     instance1.DoPing(host2, port2)
//     contact2, err := instance1.FindContact(instance2.NodeID)
//     if err != nil {
//         t.Error("Instance 2's contact not found in Instance 1's contact list")
//         return
//     }

//



//     Key := NewRandomID()
//     Value := []byte("Hello world")
//     result := instance2.DoStore(contact2, Key, Value)
//     if p:= strings.Index(result,"ERR");p==0 {
//       t.Error("Can not store this value")
//     }

//



//     result = instance1.DoFindValue(contact2, Key)
//     if p:= strings.Index(result,"ERR");p==0 {
//       t.Error("Can not find this value")
//     }
//   }
