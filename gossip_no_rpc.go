/*
CPE 569 Lab 2: Gossip Protocol w/o RPC
Author: Lorenzo Pedroza
Date: 1-12-24
*/
package main

import (
	"fmt"
	"lab2/shared"
)

func main() {

	fmt.Printf("Execution Time")
	nodes := shared.NewMembership() //New membership list
	fmt.Printf("%v", nodes)

	//Make a go routine to 'run' a node (runNode(X, Y, Z, channel, membership))

	//go runNode(X,Y,Z)

	//use wait groups to wait on all nodes to die..1f
	//
	//Each node spawns with a predetermined death (Z) time
	//All

	// rand.Float32() //returns range [0, 1.00). Clarify with Pantoja
}
