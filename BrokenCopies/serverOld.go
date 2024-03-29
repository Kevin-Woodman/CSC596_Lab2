//Author Kevin Woodman
package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

type safeRequests struct {
	mu sync.Mutex
	pR map[int]Membership
}

// Node struct represents a computing node.
type Node struct {
	ID        int
	Hbcounter int
	Time      float64
	Alive     bool
}

// Membership struct represents participanting nodes
type Membership struct {
	Members map[int]Node
}

// Request struct represents a new message request to a client
type Request struct {
	ID    int
	Table Membership
}

// Requests struct represents pending message requests
type Requests struct {
	Pending map[int]Membership
}

var publicRequests safeRequests

type TEST int

func main() {
	publicRequests.pR = make(map[int]Membership)
	// create a Membership list

	requests := new(TEST)

	// register nodes with `rpc.DefaultServer`
	rpc.Register(requests)

	// register an HTTP handler for RPC communication
	rpc.HandleHTTP()

	listener, _ := net.Listen("tcp", ":4040")

	http.Serve(listener, nil)
}

// Returns a new instance of a Membership (pointer).
func NewMembership() *Membership {
	return &Membership{
		Members: make(map[int]Node),
	}
}

// Returns a new instance of a Membership (pointer).
func NewRequests() *Requests {
	return &Requests{
		Pending: make(map[int]Membership),
	}
}

// Adds a new message request to the pending list
func (req *TEST) Add(payload Request, reply *Request) error {
	_, ok := publicRequests.pR[payload.ID]
	if ok {
		publicRequests.pR[payload.ID] = *NewMembership()
	}
	publicRequests.pR[payload.ID] = combineTablesServer(publicRequests.pR[payload.ID], payload.Table)
	return nil
}

// Listens to communication from neighboring nodes.
func (req *TEST) Listen(ID int, reply *map[int]Node) error {
	_, ok := publicRequests.pR[ID]
	if ok {
		*reply = publicRequests.pR[ID].Members
	}
	publicRequests.pR[ID] = Membership{Members: make(map[int]Node)}
	return nil
}

func combineTablesServer(oldTable Membership, recivedTable Membership) Membership {
	newMembership := Membership{Members: make(map[int]Node)}
	for _, node := range oldTable.Members {
		if _, ok := recivedTable.Members[node.ID]; ok { //If it exists in the new table
			if node.Hbcounter > recivedTable.Members[node.ID].Hbcounter { //Old table is more up to date
				newMembership.Members[node.ID] = node
			} else { //New table is more up to date
				newMembership.Members[node.ID] = recivedTable.Members[node.ID]
			}
		} else { //Not in the new table
			newMembership.Members[node.ID] = node
		}
	}

	for _, node := range recivedTable.Members {
		if _, ok := newMembership.Members[node.ID]; !ok {
			newMembership.Members[node.ID] = recivedTable.Members[node.ID]
		}
	}

	return newMembership
}
