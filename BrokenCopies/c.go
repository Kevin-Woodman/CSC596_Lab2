package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_NODES  = 8
	X_TIME     = 500
	Y_TIME     = 1
	DEAD_TIME  = 5
	Z_TIME_MAX = 1000
	Z_TIME_MIN = 500
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

var startTime time.Time
var wg = &sync.WaitGroup{}

func main() {
	// Address to this variable will be sent to the RPC server
	// Type of reply should be same as that specified on server

	rand.Seed(time.Now().UnixNano())
	Z_TIME := rand.Intn(Z_TIME_MAX-Z_TIME_MIN) + Z_TIME_MIN

	args := os.Args[1:]
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

	// DialHTTP connects to an HTTP RPC server at the specified network
	client, err := rpc.DialHTTP("tcp", "localhost"+":1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	neighbors := InitializeNeighbors(id)
	fmt.Println("Neighbors:", neighbors)

	membership := make(map[int]Node)
	self_node := Node{ID: id, Hbcounter: 0, Time: 0, Alive: true}

	time.AfterFunc(time.Millisecond*X_TIME, func() { runAfterX(client, &self_node, &membership, id) })
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(client, neighbors, &membership, id) })
	time.AfterFunc(time.Second*time.Duration(Z_TIME), func() { runAfterZ(client, id) })

	wg.Add(1)
	wg.Wait()
}

func calcTime() float64 {
	return time.Now().Sub(startTime).Seconds()
}

func runAfterX(client *rpc.Client, node *Node, membership *map[int]Node, id int) {
	node.Hbcounter += 1
	node.Time = calcTime()
	(*membership)[id] = *node

	//See if you have messages
	*membership = readMessages(*client, id, *membership)

	time.AfterFunc(time.Millisecond*X_TIME, func() { runAfterX(client, node, membership, id) })
}

func runAfterY(client *rpc.Client, neighbors [2]int, membership *map[int]Node, id int) {
	//See if any members have died
	newMem := make(map[int]Node)
	sentMem := make(map[int]Node)
	for _, m := range *membership {
		if !m.Alive && m.Time < calcTime()-(DEAD_TIME*2) { //If m has been dead for longer then 2 times the dead rate
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
		sendMessage(*client, n, sentMem)
	}

	printMembership(*membership)
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(client, neighbors, membership, id) })
}

func runAfterZ(client *rpc.Client, id int) { //end itself
	fmt.Printf("Node %d ending\n", id)
	client.Close()
	wg.Done()
	os.Exit(0)

}

// Send the current membership table to a neighboring node with the provided ID
func sendMessage(server rpc.Client, id int, membership map[int]Node) {
	var ret int
	if err := server.Call("API.AddMember", Request{Table: membership, ID: id}, &ret); err != nil {
		fmt.Println("Error: AddMember()", err)
	}
}

// Read incoming messages from other nodes
func readMessages(server rpc.Client, id int, membership map[int]Node) map[int]Node {
	newMembership := make(map[int]Node)
	if err := server.Call("API.GetMember", id, &newMembership); err != nil {
		fmt.Println("Error: GetMember()", err)
	}
	return membership //combineTables(membership, newMembership)
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
	/*for _, node := range oldTable.Members {
		if _, ok := recivedTable.Members[node.ID]; ok { //If it exists in the new table
			if node.Hbcounter >= recivedTable.Members[node.ID].Hbcounter { //Old table is more up to date
				newMembership.Members[node.ID] = node
			} else {
				if !recivedTable.Members[node.ID].Alive {
					continue
				}
				newNode := recivedTable.Members[node.ID]
				newNode.Time = calcTime()
				newMembership.Members[node.ID] = newNode
			}
		} else {
			newMembership.Members[node.ID] = node //If it isn't in the new table
		}
	}

	for _, node := range recivedTable.Members {
		if !node.Alive {
			continue
		}
		if _, ok := newMembership.Members[node.ID]; !ok { //If the node isn't in the table
			newNode := node
			newNode.Time = calcTime()
			newMembership.Members[node.ID] = newNode
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

func InitializeNeighbors(id int) [2]int {
	neighbor1 := (id + 1) % MAX_NODES
	neighbor2 := (id - 1 + MAX_NODES) % MAX_NODES
	/*neighbor1 := RandInt()
	for neighbor1 == id {
		neighbor1 = RandInt()
	}
	neighbor2 := RandInt()
	for neighbor1 == neighbor2 || neighbor2 == id {
		neighbor2 = RandInt()
	}*/
	return [2]int{neighbor1, neighbor2}
}
func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(MAX_NODES-1+1) + 1
}
