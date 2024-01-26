//Author: Kevin Woodman
package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_NODES  = 8
	X_TIME     = 1
	Y_TIME     = 2
	DEAD_TIME  = 4
	Z_TIME_MAX = 100
	Z_TIME_MIN = 10
)

var self_node Node
var startTime time.Time
var wg = &sync.WaitGroup{}

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

// Send the current membership table to a neighboring node with the provided ID
func sendMessage(server rpc.Client, id int, membership Membership) {
	var ret Request
	if err := server.Call("TEST.Add", Request{Table: membership, ID: id}, &ret); err != nil {
		fmt.Println("Error: Requests.Add()", err)
	}
}

// Read incoming messages from other nodes
func readMessages(server rpc.Client, id int, membership Membership) *Membership {
	newMembership := Membership{Members: make(map[int]Node)}

	/*
		Error: Requests.Listen reading body gob: decoding into local type *map[int]shared.Node,
		received remote type Request = struct { ID int; Table Membership = struct { Members map[int] =
		 struct { ID int; Hbcounter int; Time float; Alive bool; }; }; }

		How does this even make sense? It objectivly didn't get sent a request
	*/
	if err := server.Call("TEST.Listen", self_node.ID, &(newMembership.Members)); err != nil {
		fmt.Println("Error: Requests.Listen", err)
	}
	return combineTables(&membership, &newMembership)
}

func calcTime() float64 {
	return time.Now().Sub(startTime).Seconds()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	Z_TIME := rand.Intn(Z_TIME_MAX-Z_TIME_MIN) + Z_TIME_MIN

	// Connect to RPC server
	server, _ := rpc.DialHTTP("tcp", "localhost:9005")

	args := os.Args[1:]

	// Get ID from command line argument
	if len(args) == 0 {
		fmt.Println("No args given")
		return
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Found Error", err)
	}

	fmt.Println("Node", id, "will fail after", Z_TIME, "seconds")

	// Construct self
	startTime = time.Now()

	self_node = Node{
		ID: id, Hbcounter: 0, Time: 0, Alive: true}

	neighbors := self_node.InitializeNeighbors(id)
	fmt.Println("Neighbors:", neighbors)

	membership := &Membership{Members: make(map[int]Node)}
	membership.Members[self_node.ID] = self_node

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, &self_node, &membership, id) })
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, &membership, id) })
	time.AfterFunc(time.Second*time.Duration(Z_TIME), func() { runAfterZ(server, id) })

	wg.Add(1)
	wg.Wait()
}

func runAfterX(server *rpc.Client, node *Node, membership **Membership, id int) {
	node.Hbcounter += 1
	node.Time = calcTime()
	(*membership).Members[id] = *node

	//See if you have messages
	temp := readMessages(*server, self_node.ID, **membership)
	if temp != nil {
		*membership = temp
	}

	//See if any members have died
	var newMem = &Membership{Members: make(map[int]Node)}
	for _, m := range (*membership).Members {
		if !m.Alive && m.Time < calcTime()-(DEAD_TIME*2) { //If m has been dead for longer then 2 times the dead rate
			continue
		}
		if m.Time < calcTime()-DEAD_TIME { //If m has died
			m.Alive = false
		}
		newMem.Members[m.ID] = m
	}
	*membership = newMem

	printMembership(**membership)

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, &self_node, membership, id) })
}

func runAfterY(server *rpc.Client, neighbors [2]int, membership **Membership, id int) {
	for _, n := range neighbors {
		sendMessage(*server, n, **membership)
	}

	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, membership, id) })
}

func runAfterZ(server *rpc.Client, id int) { //end itself
	fmt.Printf("Node %d ending\n", id)
	server.Close()
	wg.Done()
	os.Exit(0)

}

func combineTables(oldTable *Membership, recivedTable *Membership) *Membership {
	var newMembership = &Membership{Members: make(map[int]Node)}
	for _, node := range oldTable.Members {
		if node.Hbcounter >= recivedTable.Members[node.ID].Hbcounter { //Old table is more up to date
			newMembership.Members[node.ID] = node
		} else {
			newNode := recivedTable.Members[node.ID]
			newNode.Time = calcTime()
			newMembership.Members[node.ID] = newNode
		}
	}
	for _, node := range recivedTable.Members {
		if _, ok := newMembership.Members[node.ID]; node.Alive && !ok { //If the node isn't in the table
			newNode := recivedTable.Members[node.ID]
			newNode.Time = calcTime()
			newMembership.Members[node.ID] = newNode
		}

	}
	return newMembership
}

func printMembership(m Membership) {
	for _, val := range m.Members {
		status := "is Alive"
		if !val.Alive {
			status = "is Dead"
		}
		fmt.Printf("Node %d has hb %d, time %.1f and %s\n", val.ID, val.Hbcounter, val.Time, status)
	}
	fmt.Println("")
}

func (n Node) InitializeNeighbors(id int) [2]int {
	//neighbor1 := (id + 1) % MAX_NODES
	//neighbor2 := (id - 1 + MAX_NODES) % MAX_NODES
	neighbor1 := RandInt()
	for neighbor1 == id {
		neighbor1 = RandInt()
	}
	neighbor2 := RandInt()
	for neighbor1 == neighbor2 || neighbor2 == id {
		neighbor2 = RandInt()
	}
	return [2]int{neighbor1, neighbor2}
}

func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(MAX_NODES-1+1) + 1
}
