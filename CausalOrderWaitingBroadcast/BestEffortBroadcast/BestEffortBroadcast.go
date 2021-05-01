package BestEffortBroadcast

import (
	"andriuslima/CausalOrderWaitingBroadcastChat/CausalOrderWaitingBroadcast/BestEffortBroadcast/PP2PLink"
	"encoding/json"
	"fmt"
	"log"
)

type ReqMessage struct {
	Addresses []string
	Process   int
	W         []int
	Message   string
}

type IndMessage struct {
	From    string
	Process int
	W       []int
	Message string
}

type Module struct {
	Me     int
	Req    chan ReqMessage
	Ind    chan IndMessage
	Logger *log.Logger

	Pp2plink PP2PLink.Module
}

func (module *Module) Init(address string) {
	module.log("Initializing Best Effort Broadcast")
	module.Pp2plink = PP2PLink.Module{
		Me:     module.Me,
		Req:    make(chan PP2PLink.ReqMessage),
		Ind:    make(chan PP2PLink.IndMessage),
		Logger: module.Logger,
	}
	module.Pp2plink.Init(address)
	module.Start()
	module.log("Initializing Best Effort Broadcast done")
}

func (module *Module) Start() {
	go func() {
		for {
			select {
			case y := <-module.Req:
				module.Broadcast(y)
			case y := <-module.Pp2plink.Ind:
				module.Deliver(PP2PLinkBEB(y))
			}
		}
	}()
}

func (module *Module) Broadcast(message ReqMessage) {
	for i := 0; i < len(message.Addresses); i++ {
		msg := BEB2PP2PLink(message)
		msg.To = message.Addresses[i]
		module.Pp2plink.Req <- msg
		module.log(fmt.Sprintf("Message sent to %v", message.Addresses[i]))
	}
}

func (module *Module) Deliver(message IndMessage) {
	module.log(fmt.Sprintf("Message received {%v} from `%v`", message.Message, message.From))
	module.Ind <- message
}

func BEB2PP2PLink(message ReqMessage) PP2PLink.ReqMessage {
	WAsBytes, err := json.Marshal(message.W)
	if err != nil {
		panic(err)
	}
	ProcessAsBytes, err := json.Marshal(message.Process)
	if err != nil {
		panic(err)
	}
	Data := make(map[string]string)
	Data["Process"] = string(ProcessAsBytes)
	Data["W"] = string(WAsBytes)
	return PP2PLink.ReqMessage{
		To: message.Addresses[0],
		Message: PP2PLink.Message{
			Value: message.Message,
			Data:  Data,
		},
	}

}

func (module *Module) log(msg string) {
	module.Logger.Printf("[BEB] - %v", msg)
}

func PP2PLinkBEB(message PP2PLink.IndMessage) IndMessage {
	rawProcess := json.RawMessage(message.Message.Data["Process"])
	processBytes, err := rawProcess.MarshalJSON()
	if err != nil {
		panic(err)
	}
	var Process int
	err = json.Unmarshal(processBytes, &Process)
	if err != nil {
		panic(err)
	}

	rawW := json.RawMessage(message.Message.Data["W"])
	WBytes, err := rawW.MarshalJSON()
	if err != nil {
		panic(err)
	}
	var W []int
	err = json.Unmarshal(WBytes, &W)
	if err != nil {
		panic(err)
	}

	return IndMessage{
		From:    message.From,
		Message: message.Message.Value,
		Process: Process,
		W:       W,
	}
}
