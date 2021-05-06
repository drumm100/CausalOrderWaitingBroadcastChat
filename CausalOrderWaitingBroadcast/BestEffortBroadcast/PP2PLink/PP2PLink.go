package PP2PLink

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Value string
	Data  map[string]string
}

type ReqMessage struct {
	To      string
	Message Message
}

type IndMessage struct {
	From    string
	Message Message
}

type Module struct {
	Me    int
	Ind   chan IndMessage
	Req   chan ReqMessage
	Run   bool
	Cache map[string]net.Conn
}

func (module *Module) Init(address string) {
	//logFile, err := os.Create(fmt.Sprintf("logs/pp2plink_%v_%v.log", module.Me, time.Now().Unix()))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//module.Logger = log.New(logFile, fmt.Sprintf("[%v] ", module.Me), log.LstdFlags)

	module.log("Initializing PP2P")

	if module.Run {
		return
	}

	module.Cache = make(map[string]net.Conn)
	module.Run = true
	module.Start(address)
}

func (module *Module) Start(address string) {
	go func() {
		listen, _ := net.Listen("tcp4", address)
		for {
			conn, err := listen.Accept()
			go func() {
				for {
					if err != nil {
						fmt.Println(err)
						continue
					}
					bufTam := make([]byte, 4)
					_, err := io.ReadFull(conn, bufTam)
					if err != nil {
						fmt.Println(err)
						continue
					}
					tam, err := strconv.Atoi(string(bufTam))
					bufMsg := make([]byte, tam)
					_, err = io.ReadFull(conn, bufMsg)
					if err != nil {
						fmt.Println(err)
						continue
					}
					bufMsgAsStruct := StringToMessage(string(bufMsg))
					msg := IndMessage{
						From:    conn.RemoteAddr().String(),
						Message: bufMsgAsStruct,
					}
					module.Ind <- msg
				}
			}()
		}
	}()

	go func() {
		for {
			message := <-module.Req
			go module.Send(message)
		}
	}()

}

func (module *Module) Send(message ReqMessage) {
	value := message.Message.Value
	if strings.Contains(value, "delay") {
		time.Sleep(10 * time.Second)
	}

	var conn net.Conn
	var ok bool
	var err error

	if conn, ok = module.Cache[message.To]; ok {

	} else {
		conn, err = net.Dial("tcp", message.To)
		if err != nil {
			fmt.Println(err)
			return
		}
		module.Cache[message.To] = conn
	}
	messageAsString := MessageToString(message.Message)
	strSize := strconv.Itoa(len(messageAsString))
	for len(strSize) < 4 {
		strSize = "0" + strSize
	}
	if !(len(strSize) == 4) {
		module.log("ERROR AT PP2PLink MESSAGE SIZE CALCULATION - INVALID MESSAGES MAY BE IN TRANSIT")
	}
	_, err = fmt.Fprintf(conn, strSize)
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(conn, messageAsString)
	if err != nil {
		return
	}
}

func (module *Module) log(msg string) {
	fmt.Printf("[LINK] - %v\n", msg)
}

func StringToMessage(s string) Message {
	rawIn := json.RawMessage(s)
	bytes, err := rawIn.MarshalJSON()
	if err != nil {
		panic(err)
	}

	var message Message
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		panic(err)
	}
	return message
}

func MessageToString(message Message) string {
	bytes, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
