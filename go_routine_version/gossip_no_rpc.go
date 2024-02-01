/*
CPE 569 Lab 2: Gossip Protocol w/o RPC
Author: Lorenzo Pedroza with some helpp by Kevin Woodman
Date: 1-25-24
*/
package main

import (
	"Lab2/shared"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func elapsedTime(start time.Time) float64 {
	return time.Now().Sub(start).Seconds()
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

func combineTables(oldTable shared.Membership, recivedTable shared.Membership, elapsedTime float64) *shared.Membership {
	var newMembership = shared.NewMembership()
	for _, node := range oldTable.Members {
		if node.Hbcounter >= recivedTable.Members[node.ID].Hbcounter { //Old table is more up to date
			newMembership.Members[node.ID] = node
		} else {
			newNode := recivedTable.Members[node.ID]
			newNode.Time = elapsedTime
			newMembership.Members[node.ID] = newNode
		}
	}
	for _, node := range recivedTable.Members {
		if _, ok := newMembership.Members[node.ID]; node.Alive && !ok { //If the node isn't in the table
			newNode := recivedTable.Members[node.ID]
			newNode.Time = elapsedTime
			newMembership.Members[node.ID] = newNode
		}

	}
	return newMembership
}

// Mark unreponsive members and clear them after being unresponsive for too long
func update_deaths(membership shared.Membership, start_time time.Time) *shared.Membership {
	var newMem = shared.NewMembership()
	for _, m := range (membership).Members {
		if !m.Alive && m.Time < elapsedTime(start_time)-(shared.DEAD_TIME+shared.CLEAN_UP_TIME) { //If m has been dead for longer then 2 times the dead rate
			continue
		}
		if m.Time < elapsedTime(start_time)-shared.DEAD_TIME { //If m has died
			m.Alive = false
		}
		newMem.Add(m, &m)
	}
	return newMem
}

// remember maps are already reference types
func run_node(wg *sync.WaitGroup, memb_chans map[int](chan shared.Membership), z, y, x, id int) {
	defer wg.Done()
	start_time := time.Now()

	self_node := shared.Node{
		ID: id, Hbcounter: 0, Time: 0, Alive: true}

	membership := shared.NewMembership()
	neighbors := self_node.InitializeNeighborsNotRandom(id)
	fmt.Println("Node", id, "will fail after", z, "seconds\n")
	fmt.Println("Node", id, "Neighbors:", neighbors, "\n")

	membership.Add(self_node, &self_node)
	printMembership(*membership)

	z_die := time.After(time.Duration(z) * time.Second)
	y_tick_send_tables := time.Tick(time.Duration(y) * time.Second)
	x_tick_increment_hb := time.Tick(time.Duration(x) * time.Second)

	//go ahead and broadcast local table to two eighbors
	//beat heart, listen for updates.
	for {
		select {
		case <-z_die:
			fmt.Println("\nNode", id, "failed 💀")
			// don't close channels!!
			return //done running
		case <-y_tick_send_tables: //
			//transmit membership table to neighbors //fanout 2
			fmt.Println("DEBUG: Node", id, ": sending tables to ", neighbors)
			//do not send to a full channel (dead nodes have full channles)!! Use select...case to only send when ready
			select {
			case memb_chans[neighbors[0]] <- *membership:
			default:
				fmt.Println("DEBUG: Node ", id, " can't send tables to neighbor", neighbors[0], " channel is full")
			}

			select {
			case memb_chans[neighbors[1]] <- *membership:
			default:
				fmt.Println("DEBUG: Node ", id, " can't send tables to neighbor", neighbors[1], " channel is full")
			}

		case <-x_tick_increment_hb:
			//
			self_node.Hbcounter++
			self_node.Time = elapsedTime(start_time)
			membership.Update(self_node, &self_node)
			fmt.Println("\nNode:", id, " 💗", "local time:=", elapsedTime(start_time))
			fmt.Println("\nNode:", id, " Membership Table")
			printMembership(*membership)

		case new_membership := <-memb_chans[id]:
			//see if we got some new tables to merge
			fmt.Println("\nDEBUG: Node ", id, ": recived new tables, updating")
			membership = combineTables(*membership, new_membership, elapsedTime(start_time))

		default:
			//fmt.Println("Node", id, "Checking for dead members")
			//Identify dead members every Tdead and remove dead members after Tdead+Tclenaup

			membership = update_deaths(*membership, start_time)
			// var newMem = shared.NewMembership()
			// for _, m := range (*membership).Members {
			// 	if !m.Alive && m.Time < elapsedTime(start_time)-(shared.DEAD_TIME+shared.CLEAN_UP_TIME) { //If m has been dead for longer then 2 times the dead rate
			// 		continue
			// 	}
			// 	if m.Time < elapsedTime(start_time)-shared.DEAD_TIME { //If m has died
			// 		m.Alive = false
			// 	}
			// 	newMem.Add(m, &m)
			// }
			// membership = newMem
		}
	}

}

func main() {

	var wg sync.WaitGroup //keep going until everyone dies
	memb_chans := make(map[int](chan shared.Membership))

	for i := 0; i < shared.MAX_NODES; i++ {
		memb_chans[i] = make(chan shared.Membership, 2*shared.MAX_NODES)
	}

	fmt.Println(memb_chans)

	var Z_TIME int

	for i := 0; i < shared.MAX_NODES; i++ {
		wg.Add(1)
		Z_TIME = rand.Intn(shared.Z_TIME_MAX-shared.Z_TIME_MIN) + shared.Z_TIME_MIN
		go run_node(&wg, memb_chans, Z_TIME, shared.Y_TIME, shared.X_TIME, i)
	}

	// //Basic Test 1
	// wg.Add(1)
	// go run_node(&wg, memb_chans, 12, shared.Y_TIME, shared.X_TIME, 0)
	// wg.Add(1)
	// go run_node(&wg, memb_chans, 16, shared.Y_TIME, shared.X_TIME, 1)
	// wg.Add(1)
	// go run_node(&wg, memb_chans, 24, shared.Y_TIME, shared.X_TIME, 2)

	wg.Wait()
}
