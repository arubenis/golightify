package golightify

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Request struct {
	ch      chan interface{}
	message LightifyRequest
}

var requestId uint32
var requests = make(map[uint32]Request)
var requestsMutex = &sync.Mutex{}
var conn net.Conn

type LightifyMessageSerializer interface {
	LightifySerialize(w io.Writer) error
}

type LightifyMessageDeserializer interface {
	LightifyDeserialize(r io.Reader) error
}

func deafultMessageSerializer(w io.Writer, msg interface{}) error {
	return binary.Write(w, binary.LittleEndian, msg)
}

func deafultMessageDeserializer(r io.Reader, t interface{}) error {
	return binary.Read(r, binary.LittleEndian, &t)
}

func SendLightifyRequest(lightifyRequest LightifyRequest) interface{} {

	message := new(LightifyMessage)
	message.Data = lightifyRequest

	buf := new(bytes.Buffer)

	var err error

	if v, ok := message.Data.(LightifyMessageSerializer); ok {
		err = v.LightifySerialize(buf)
	} else {
		err = deafultMessageSerializer(buf, message.Data)
	}

	if err != nil {
		log.Println(err)
	}

	header := new(LightifyMessageHeader)
	header.Length = uint16(LightifyMessageHeader_DataLength + buf.Len())
	header.Command = message.Data.Command()
	header.Id = atomic.AddUint32(&requestId, 1)
	//header.Unknown1 = 2

	messageBuf := new(bytes.Buffer)

	err = binary.Write(messageBuf, binary.LittleEndian, header)
	if err != nil {
		log.Println(err)
	}

	_, err = messageBuf.Write(buf.Bytes())

	err = binary.Write(conn, binary.LittleEndian, messageBuf.Bytes())
	if err != nil {
		log.Println(err)
	}
	//log.Printf("Send Request to lightify (%d): % x\n", messageBuf.Len(), messageBuf.Bytes())

	requestsMutex.Lock()
	requests[header.Id] = Request{ch: make(chan interface{}), message: lightifyRequest}
	requestsMutex.Unlock()

	chResponse := requests[header.Id]
	//log.Printf("chResponse: %x\n", <-chResponse)
	return <-chResponse.ch
}

func handleResponse(conn net.Conn) {
	var messageHeader LightifyMessageHeader
	for {
		err := binary.Read(conn, binary.LittleEndian, &messageHeader)
		if err != nil {
			log.Println(err)
		}

		requestsMutex.Lock()
		request := requests[messageHeader.Id]
		delete(requests, messageHeader.Id)
		requestsMutex.Unlock()

		dataLen := messageHeader.Length - LightifyMessageHeader_DataLength
		//log.Printf("handleResponse from lightify, messageId %v, dataLen: %d, requestMessage: command=%d, %v\n", messageHeader, dataLen, request.message.Command(), request.message)
		if dataLen > 0 {
			p := make([]byte, dataLen)

			_, err = conn.Read(p)
			if nil != err {
				log.Println(conn.RemoteAddr(), err)
			}
			buf := bytes.NewBuffer(p)

			response := request.message.NewResponse()
			if response != nil {
				if v, ok := response.(LightifyMessageDeserializer); ok {
					v.LightifyDeserialize(buf)
				} else {
					err = deafultMessageDeserializer(buf, response)
					if err != nil {
						log.Println("Error, can not deserialize response")
					}
				}
			}

			request.ch <- response

			log.Printf("\n\nhandleResponse from lightify, raw data:\n% x\n\nDeserialized: %+v\n", p, response)
		}

		close(request.ch)

	}
}

func NewLightifyBridge(address string) error {
	var err error
	conn, err = net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}
	log.Println("Connected to", conn.RemoteAddr())

	go handleResponse(conn)

	return err
}
