package main

import (
	"andriuslima/CausalOrderWaitingBroadcastChat/CausalOrderWaitingBroadcast"
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main() {
	validateArgs()

	index, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	myAddress := os.Args[2]
	addresses := os.Args[3:]
	fmt.Printf("Addresses: %v \n", addresses)
	fmt.Printf("Process address: %v \n", myAddress)
	fmt.Printf("Process index: %v \n", index)

	fmt.Println("Initializing Chat")

	broadcast := CausalOrderWaitingBroadcast.Module{
		Send:      make(chan CausalOrderWaitingBroadcast.SendMessageRequest),
		Deliver:   make(chan CausalOrderWaitingBroadcast.DeliverMessageRequest),
		Me:        index,
		Addresses: append(addresses, myAddress),
	}

	broadcast.Init(myAddress)

	// Broadcast Messages
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			if scanner.Scan() {
				msg := scanner.Text()
				req := CausalOrderWaitingBroadcast.SendMessageRequest{Message: msg}
				broadcast.Send <- req
			}
		}
	}()

	// Deliver Value
	go func() {
		for {
			in := <-broadcast.Deliver
			if in.Process != index {
				fmt.Printf("[%v] says: %v \n", in.Process, in.Message)
			}
		}
	}()

	blq := make(chan int)
	<-blq
	fmt.Println("Initializing Chat done")
}

func validateArgs() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run Chat.go 1 127.0.0.1:5001 127.0.0.1:6001")
		fmt.Println("go run Chat.go 2 127.0.0.1:6001 127.0.0.1:5001")
		panic("Invalid arguments!")
	}
}
