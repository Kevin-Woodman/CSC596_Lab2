// Author: Kevin Woodman
// sendMessage(), readMessage(), and structs provided by Dr. Pantoja
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
	Z_TIME_MAX = 200
	Z_TIME_MIN = 50
	DEAD_TIME  = 6
)

// Request struct represents a new message request to a client
type Request struct {
	ID    int
	Table map[int]Node
}

type Node struct {
	ID        int
	Hbcounter int
	Time      float64
	Alive     bool
}

var self_node Node
var startTime time.Time

// Send the current membership table to a neighboring node with the provided ID
func sendMessage(server rpc.Client, id int, membership map[int]Node) {
	if err := server.Call("Requests.Add", Request{ID: id, Table: membership}, nil); err != nil {
		fmt.Println("Error:3 Requests.Add()", err)
	}
}

// Read incoming messages from other nodes
func readMessages(server rpc.Client, id int, membership map[int]Node) *map[int]Node {
	table := make(map[int]Node)
	if err := server.Call("Requests.Listen", id, &table); err != nil {
		fmt.Println("Error:4 Requests.Listen()", err)
	}
	return &table
}

var wg = &sync.WaitGroup{}

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

	startTime = time.Now()
	// Construct self
	self_node = Node{ID: id, Hbcounter: 0, Time: 0, Alive: true}

	neighbors := InitializeNeighbors(id)
	fmt.Println("Neighbors:", neighbors)

	membership := make(map[int]Node)

	sendMessage(*server, neighbors[0], membership)

	// crashTime := self_node.CrashTime()

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, &self_node, &membership, id) })
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, &membership, id) })
	time.AfterFunc(time.Second*time.Duration(Z_TIME), func() { runAfterZ(server, id) })

	wg.Add(1)
	wg.Wait()
}

func runAfterX(server *rpc.Client, node *Node, membership *map[int]Node, id int) {
	m := *membership
	node.Hbcounter = node.Hbcounter + 1
	node.Time = calcTime()
	m[node.ID] = *node
	*membership = m

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, node, membership, id) })
}

func runAfterY(server *rpc.Client, neighbors [2]int, membership *map[int]Node, id int) {
	//See if any members have died
	newMem := make(map[int]Node)
	sentMem := make(map[int]Node)
	for _, m := range *membership {
		if !m.Alive && m.Time < calcTime()-(DEAD_TIME*3) { //If m has been dead for longer then 2 times the dead rate
			continue
		}
		if m.Time < calcTime()-DEAD_TIME { //If m has died
			m.Alive = false
		} else {
			m.Alive = true
			sentMem[m.ID] = m
		}
		newMem[m.ID] = m
	}

	*membership = newMem
	for _, n := range neighbors {
		sendMessage(*server, n, sentMem)
	}

	m := *membership
	table := readMessages(*server, id, m)
	m = combineTables(m, *table)
	*membership = m

	printMembership(*membership)
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, membership, id) })
}

func runAfterZ(server *rpc.Client, id int) {
	server.Close()
	fmt.Printf("NODE %d FAILED\n", id)
	os.Exit(0)
}

func combineTables(oldTable map[int]Node, recivedTable map[int]Node) map[int]Node {
	newMembership := make(map[int]Node)
	for id, node := range oldTable {
		if _, ok := recivedTable[id]; ok && node.Hbcounter < recivedTable[id].Hbcounter {
			tempNode := recivedTable[id]
			tempNode.Time = calcTime()
			newMembership[id] = tempNode // New table's entry is more recent
		} else {
			newMembership[id] = node // Old table's entry is more recent
		}
	}

	for id, node := range recivedTable {
		if _, ok := newMembership[id]; !ok {
			tempNode := node
			tempNode.Time = calcTime()
			newMembership[id] = tempNode
		}
	}

	return newMembership
}

func InitializeNeighbors(id int) [2]int {
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

func printMembership(m map[int]Node) {
	for _, val := range m {
		status := "is Alive"
		if !val.Alive {
			status = "is Dead"
		}
		fmt.Printf("Node %d has hb %d, time %.1f and %s\n", val.ID, val.Hbcounter, val.Time, status)
	}
	fmt.Println("")
}

func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(MAX_NODES)
}

func calcTime() float64 {
	return time.Now().Sub(startTime).Seconds()
}
