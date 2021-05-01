package CausalOrderWaitingBroadcast

import (
	"andriuslima/CausalOrderWaitingBroadcastChat/CausalOrderWaitingBroadcast/BestEffortBroadcast"
	"fmt"
	"log"
)

type SendMessage struct {
	Message string
}

type DeliverMessage struct {
	Process int
	W       []int
	Message string
}

type Module struct {
	Send      chan SendMessage
	Deliver   chan DeliverMessage
	Me        int
	Addresses []string
	Logger    *log.Logger

	BEB     BestEffortBroadcast.Module
	V       []int
	LSN     int
	Pending []DeliverMessage
}

func (module *Module) Init(address string) {
	module.log("Initializing Causal Order Broadcast")
	module.BEB = BestEffortBroadcast.Module{
		Me:     module.Me,
		Req:    make(chan BestEffortBroadcast.ReqMessage),
		Ind:    make(chan BestEffortBroadcast.IndMessage),
		Logger: module.Logger,
	}

	module.V = make([]int, len(module.Addresses)+1) // +2 so each process has it's own index.
	module.LSN = 0
	module.BEB.Init(address)
	module.Start()
	module.log("Initializing Causal Order Broadcast done")
}

func (module *Module) Start() {
	module.log("Listening...")
	go func() {
		for {
			select {
			case msg := <-module.Send:
				module.SendMessage(msg)
			case msg := <-module.BEB.Ind:
				module.log(fmt.Sprintf("Message received from %v: %v \n", msg.Process, msg))
				module.Pending = append(module.Pending, COBFromBEB(msg))
				module.ProcessPendingQueue()
			}
		}
	}()
}

func (module *Module) SendMessage(message SendMessage) {
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
	module.log(fmt.Sprintf("Message sended to BestEffortBroadcast: %v \n", req))
	module.BEB.Req <- req
}

func (module *Module) ProcessPendingQueue() {
	thereIsAMessageToDeliver, message := module.retrieveNextMessage()
	if thereIsAMessageToDeliver {
		module.log(fmt.Sprintf("'%v' happended before '%v'", message.W, module.V))
		module.DeliverMessage(message)
		module.ProcessPendingQueue()
	}
}

func (module *Module) DeliverMessage(msg DeliverMessage) {
	module.V[msg.Process] = module.V[msg.Process] + 1
	module.Deliver <- msg
	module.log(fmt.Sprintf("Message delivered from %v: %v \n", msg.Process, msg))
}

func (module *Module) retrieveNextMessage() (bool, DeliverMessage) {
	for i, message := range module.Pending {
		if module.happenedBefore(message) {
			module.Pending = append(module.Pending[:i], module.Pending[i+1:]...) // this removes de ith element
			return true, message
		}
	}

	return false, DeliverMessage{}
}

func (module *Module) happenedBefore(msg DeliverMessage) bool {
	for i := range msg.W {
		if module.V[i] < msg.W[i] {
			return false
		}
	}
	return true
}

func COBFromBEB(message BestEffortBroadcast.IndMessage) DeliverMessage {
	return DeliverMessage{
		Process: message.Process,
		W:       message.W,
		Message: message.Message,
	}
}

func (module *Module) log(msg string) {
	module.Logger.Printf("[COB] - %v", msg)
}
