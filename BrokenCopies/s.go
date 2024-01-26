package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

type Node struct {
	ID        int
	Hbcounter int
	Time      float64
	Alive     bool
}

type Request struct {
	ID    int
	Table map[int]Node
}

var db map[int](map[int]Node) = make(map[int](map[int]Node))
var mu sync.Mutex

type API int

func (t *API) AddMember(payload *Request, reply *int) error {
	/*
		_, ok := db[payload.ID]
		if !ok {
			db[payload.ID] = make(map[int]Node)
		}
		db[payload.ID] = combineTablesServer(db[payload.ID], payload.Table) */
	//db[payload.ID] = payload.Table

	//fmt.Printf("Adding table %d\n", payload.ID)
	//printMembership(payload.Table)

	return nil
}

func (t *API) GetMember(id int, reply *map[int]Node) error {
	/*_, ok := db[id]
	if !ok {
		db[id] = make(map[int]Node)
	}*/
	//*reply = db[id]
	//*reply = make(map[int]Node)
	//db[id] = make(map[int]Node)

	//fmt.Printf("Getting table %d\n", id)
	return nil
}

func main() {
	api := new(API)

	rpc.Register(api)
	rpc.HandleHTTP()
	// Start listening for the requests on port 1234
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listener error: ", err)
	}

	http.Serve(listener, nil)
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

	/*for _, node := range oldTable.Members {
		if _, ok := recivedTable[node.ID]; ok { //If it exists in the new table
			if node.Hbcounter > recivedTable[node.ID].Hbcounter { //Old table is more up to date
				newMembership.Members[node.ID] = node
			} else { //New table is more up to date
				newMembership.Members[node.ID] = recivedTable[node.ID]
			}
		} else { //Not in the new table
			newMembership.Members[node.ID] = node
		}
	}

	for _, node := range recivedTable {
		if _, ok := newMembership.Members[node.ID]; !ok {
			newMembership.Members[node.ID] = node
		}
	}

	return newMembership*/
}

func printMembership(m map[int]Node) {
	i := 0
	for _, val := range m {
		status := "is Alive"
		i++
		if !val.Alive {
			status = "is Dead"
		}
		fmt.Printf("Node %d has hb %d, time %.1f and %s\n", val.ID, val.Hbcounter, val.Time, status)
	}
	fmt.Printf("%d Nodes\n", i)
	fmt.Println("")
}
