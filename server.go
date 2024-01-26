// Author: Kevin Woodman
// main() and structs provided by Dr. Pantoja

package main

import (
	"io"
	"net/http"
	"net/rpc"
)

// Request struct represents a new message request to a client
type Request struct {
	ID    int
	Table map[int]Node
}

// Requests struct represents pending message requests
type Requests int

// Node struct represents a computing node.
type Node struct {
	ID        int
	Hbcounter int
	Time      float64
	Alive     bool
}

var db map[int](map[int]Node) = make(map[int](map[int]Node))

func main() {
	// create a Membership list
	requests := new(Requests)

	// register nodes with `rpc.DefaultServer`
	rpc.Register(requests)

	// register an HTTP handler for RPC communication
	rpc.HandleHTTP()

	// sample test endpoint
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "RPC SERVER LIVE!")
	})

	// listen and serve default HTTP server
	http.ListenAndServe("localhost:9005", nil)
}

// Adds a new message request to the pending list
func (req *Requests) Add(payload Request, reply *bool) error {
	_, ok := db[payload.ID]
	if !ok {
		db[payload.ID] = make(map[int]Node)
	}
	db[payload.ID] = combineTablesServer(db[payload.ID], payload.Table)
	return nil
}

// Listens to communication from neighboring nodes.
func (req *Requests) Listen(ID int, reply *map[int]Node) error {
	// get pendingRequests for ID from the membership list
	_, ok := db[ID]
	if !ok {
		db[ID] = make(map[int]Node)
	}
	*reply = db[ID]
	db[ID] = make(map[int]Node)

	return nil
}

func combineTablesServer(oldTable map[int]Node, recivedTable map[int]Node) map[int]Node {
	newMembership := make(map[int]Node)
	for id, node := range oldTable {
		if _, ok := recivedTable[id]; ok && node.Hbcounter < recivedTable[id].Hbcounter {
			newMembership[id] = recivedTable[id] // New table's entry is more recent
		} else {
			newMembership[id] = node // Old table's entry is more recent
		}
	}

	for id, node := range recivedTable {
		if _, ok := newMembership[id]; !ok {
			newMembership[id] = node // Old table's entry is more recent
		}
	}

	return newMembership
}
