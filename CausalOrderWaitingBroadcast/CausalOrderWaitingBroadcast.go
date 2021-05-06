package CausalOrderWaitingBroadcast

import (
	"andriuslima/CausalOrderWaitingBroadcastChat/CausalOrderWaitingBroadcast/BestEffortBroadcast"
	"fmt"
)

type SendMessageRequest struct {
	Message string
}

type DeliverMessageRequest struct {
	Process int
	W       []int
	Message string
}

type Module struct {
	Send      chan SendMessageRequest
	Deliver   chan DeliverMessageRequest
	Me        int
	Addresses []string

	BEB     BestEffortBroadcast.Module
	V       []int
	LSN     int
	Pending []DeliverMessageRequest
}

func (module *Module) Init(address string) {
	module.log("Initializing Causal Order Broadcast")
	module.BEB = BestEffortBroadcast.Module{
		Me:  module.Me,
		Req: make(chan BestEffortBroadcast.ReqMessage),
		Ind: make(chan BestEffortBroadcast.IndMessage),
	}

	module.V = make([]int, len(module.Addresses)+1) // +2 so each process has it's own index.
	module.LSN = 0
	module.BEB.Init(address)
	module.Start()
	module.log("Initializing Causal Order Broadcast done")
}

// Start function creates a goroutines that checks either if theres is a message to be sent on the Send channel or if there a message to be delivery by the BEB (Best Effort Broadcast module).
// It is important to note that because we are using a select statement there is not concurrency to be dealt it.
// If there is a message to be send on the Send channel the routines calls the DispatchMessageToBroadcast function
// If the BestEffortBroadcast BEB received a message through Ind channel the routine adds the message on the Pending queue and then calls the ProcessPendingQueue function
func (module *Module) Start() {
	module.log("Listening...")
	go func() {
		for {
			select {
			case msg := <-module.Send:
				module.DispatchMessageToBroadcast(msg)
			case msg := <-module.BEB.Ind:
				module.log(fmt.Sprintf("Message received from %v: %v", msg.Process, msg))
				module.Pending = append(module.Pending, COBFromBEB(msg))
				module.ProcessPendingQueue()
			}
		}
	}()
}

// DispatchMessageToBroadcast function receives a SendMessageRequest to be sent to the Req channel from BEB
// Prior to sending to BEB, the function creates a new vector clock W and increment by 1 unit the process clock, increments the LSN field
func (module *Module) DispatchMessageToBroadcast(message SendMessageRequest) {
	W := module.V
	W[module.Me] = module.LSN
	module.LSN = module.LSN + 1
	req := BestEffortBroadcast.ReqMessage{
		Message:   message.Message,
		W:         W,
		Addresses: module.Addresses,
		Process:   module.Me,
	}
	module.log(fmt.Sprintf("LSN: %v, V: %v, W: %v \n", module.LSN, module.V, W))
	module.log(fmt.Sprintf("Message sended to BestEffortBroadcast: %v", req))
	module.BEB.Req <- req
}

// ProcessPendingQueue function checks if there is a message to be delivery by calling the retrieveNextMessage function.
// If there is a message to be delivered the function calls DeliversMessage and ProcessPendingQueue functions.
// This recursion is necessary because when a message is delivered it creates the possibility of another message in the queue to be delivery.
// The recursion ends when there is no available message that matches the criteria to be delivered.
func (module *Module) ProcessPendingQueue() {
	thereIsAMessageToDeliver, message := module.retrieveNextMessage()
	if thereIsAMessageToDeliver {
		module.log(fmt.Sprintf("'%v' happended before '%v'", message.W, module.V))
		module.DeliversMessage(message)
		module.ProcessPendingQueue()
	}
}

// DeliversMessage function receives a DeliverMessageRequest.
// The function use this message to update it vector clock V and finally send the message to the Deliver channel
func (module *Module) DeliversMessage(msg DeliverMessageRequest) {
	module.V[msg.Process] = module.V[msg.Process] + 1
	module.Deliver <- msg
	module.log(fmt.Sprintf("Message delivered from %v: %v", msg.Process, msg))
}

// retrieveNextMessage function calculates if there is a message in the Pending queue that can be delivered.
// To calculate if a message can be delivered the function checks if the message happened before the process logical clock V
// If a message can be delivered it will be removed from the Pending queue and returned.
func (module *Module) retrieveNextMessage() (bool, DeliverMessageRequest) {
	for i, message := range module.Pending {
		if module.happenedBefore(message) {
			module.Pending = append(module.Pending[:i], module.Pending[i+1:]...) // this removes de ith element
			return true, message
		}
	}

	return false, DeliverMessageRequest{}
}

// happenedBefore function receives a DeliverMessageRequest and calculates if this message logical clock W happened before the process logical clock V
// W <= V -> for every i = 1..n that W[i] <= V[i}
func (module Module) happenedBefore(msg DeliverMessageRequest) bool {
	for i := range msg.W {
		if module.V[i] < msg.W[i] {
			return false
		}
	}
	return true
}

func COBFromBEB(message BestEffortBroadcast.IndMessage) DeliverMessageRequest {
	return DeliverMessageRequest{
		Process: message.Process,
		W:       message.W,
		Message: message.Message,
	}
}

func (module *Module) log(msg string) {
	fmt.Printf("[COB] - %v\n", msg)
}
