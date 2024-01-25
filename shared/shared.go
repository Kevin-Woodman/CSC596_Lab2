package shared

import (
	"math/rand"
	"time"
)

const (
	MAX_NODES = 8
)

// Node struct represents a computing node.
type Node struct {
	ID        int
	Hbcounter int
	Time      float64
	Alive     bool
}

// Generate random crash time from 10-60 seconds
func (n Node) CrashTime() int {
	rand.Seed(time.Now().UnixNano())
	max := 60
	min := 10
	return rand.Intn(max-min) + min
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

/*---------------*/

// Membership struct represents participanting nodes
type Membership struct {
	Members map[int]Node
}

// Returns a new instance of a Membership (pointer).
func NewMembership() *Membership {
	return &Membership{
		Members: make(map[int]Node),
	}
}

// Adds a node to the membership list.
func (m *Membership) Add(payload Node, reply *Node) error {
	m.Members[payload.ID] = payload
	return nil
}

// Updates a node in the membership list.
func (m *Membership) Update(payload Node, reply *Node) error { //TODO Should this be different?
	m.Members[payload.ID] = payload
	return nil
}

// Returns a node with specific ID.
func (m *Membership) Get(payload int, reply *Node) error {
	*reply = m.Members[payload]
	return nil
}

/*---------------*/

// Request struct represents a new message request to a client
type Request struct {
	ID    int
	Table Membership
}

// Requests struct represents pending message requests
type Requests struct {
	Pending map[int]Membership
}

// Returns a new instance of a Membership (pointer).
func NewRequests() *Requests {
	return &Requests{
		Pending: make(map[int]Membership),
	}
}

// Adds a new message request to the pending list
func (req *Requests) Add(payload Request, reply *bool) error {
	req.Pending[payload.ID] = payload.Table //combineTables(req.Pending[payload.ID], payload.Table)
	return nil
}

// Listens to communication from neighboring nodes.
func (req *Requests) Listen(ID int, reply *Membership) error {
	*reply = req.Pending[ID]
	req.Pending[ID] = Membership{Members: make(map[int]Node)}
	return nil
}

func combineTables(oldTable Membership, recivedTable Membership) Membership {
	newMembership := Membership{Members: make(map[int]Node)}
	for _, node := range oldTable.Members {
		if _, ok := recivedTable.Members[node.ID]; ok { //If it exists in the new table
			if node.Hbcounter > recivedTable.Members[node.ID].Hbcounter { //Old table is more up to date
				newMembership.Members[node.ID] = node
			} else {
				newMembership.Members[node.ID] = recivedTable.Members[node.ID]
			}
		} else {
			newMembership.Members[node.ID] = node
		}
	}
	for _, node := range recivedTable.Members {
		if _, ok := newMembership.Members[node.ID]; !ok { //If the node isn't in the table
			newMembership.Members[node.ID] = recivedTable.Members[node.ID]
		}

	}

	return newMembership
}
