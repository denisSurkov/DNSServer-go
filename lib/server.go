package lib

import (
	"DNSServer/lib/structures"
	"log"
	"net"
)

type IncomingRequest struct {
	Address    net.Addr
	DNSMessage *structures.DNSMessage
}

func RequestsReceiver(exit chan bool) {
	log.Println("starting server")

	pc, err := net.ListenPacket("udp", "localhost:53")
	if err != nil {
		log.Fatalf("failed to start server because of %s", err)
	}

	buffer := make([]byte, 1024)
	for {
		n, addr, err := pc.ReadFrom(buffer)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("new request from %s bytes read %d", addr, n)

		allMessage := buffer[:n]
		parsedMessage, err := structures.UnmarshalMessage(allMessage)
		incomingRequest := &IncomingRequest{
			Address:    addr,
			DNSMessage: parsedMessage,
		}
		go Resolve(incomingRequest, pc)
	}
}
