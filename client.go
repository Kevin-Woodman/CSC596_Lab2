package main

import (
	"fmt"
	"lab2/shared"
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
	Z_TIME_MAX = 100
	Z_TIME_MIN = 10
)

var self_node shared.Node
var startTime time.Time

// Send the current membership table to a neighboring node with the provided ID
func sendMessage(server rpc.Client, id int, membership shared.Membership) {
	server.Call("Requests.Add")
}

// Read incoming messages from other nodes
func readMessages(server rpc.Client, id int, membership shared.Membership) *shared.Membership {
	var newMembership shared.Membership
	if err := server.Call("Requests.Listen", self_node.ID, newMembership); err != nil {
		fmt.Println("Error:2 Requests()", err)
	}

	return combineTables(membership, newMembership)

}

func calcTime() time.Duration {
	return time.Now().Sub(startTime)
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

	// Construct self
	startTime = time.Now()

	self_node = shared.Node{
		ID: id, Hbcounter: 0, Time: 0, Alive: true}

	//var self_node_response shared.Node // Allocate space for a response to overwrite this

	// Add node with input ID
	/*if err := server.Call("Membership.Add", self_node, &self_node_response); err != nil {
		fmt.Println("Error:2 Membership.Add()", err)
	} else {
		fmt.Printf("Success: Node created with id= %d\n", id)
	}*/

	neighbors := self_node.InitializeNeighbors(id)
	fmt.Println("Neighbors:", neighbors)

	membership := shared.NewMembership()
	membership.Add(self_node, &self_node)

	sendMessage(*server, neighbors[0], *membership)

	// crashTime := self_node.CrashTime()

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, &self_node, &membership, id) })
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, &membership, id) })
	time.AfterFunc(time.Second*time.Duration(Z_TIME), func() { runAfterZ(server, id) })

	wg.Add(1)
	wg.Wait()
}

func runAfterX(server *rpc.Client, node *shared.Node, membership **shared.Membership, id int) {
	//TODO
}

func runAfterY(server *rpc.Client, neighbors [2]int, membership **shared.Membership, id int) {
	//TODO
}

func runAfterZ(server *rpc.Client, id int) { //end itself
	//TODO
}

func combineTables(oldTable *shared.Membership, recivedTable *shared.Membership) *shared.Membership {
	var newMembership = shared.NewMembership()
	for _, node := range oldTable.Members {
		if node.Hbcounter > recivedTable.Members[node.ID].Hbcounter { //Old table is more up to date
			newMembership.Members[node.ID] = node
		} else {
			newMembership.Members[node.ID] = recivedTable.Members[node.ID]
			newMembership.Members[node.ID].time = calcTime()
		}
	}
	return newMembership
}

func printMembership(m shared.Membership) {
	for _, val := range m.Members {
		status := "is Alive"
		if !val.Alive {
			status = "is Dead"
		}
		fmt.Printf("Node %d has hb %d, time %.1f and %s\n", val.ID, val.Hbcounter, val.Time, status)
	}
	fmt.Println("")
}