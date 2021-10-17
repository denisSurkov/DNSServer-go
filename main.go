package main

import (
	"DNSServer/lib/structures"
	"fmt"
	"net"
)

func main() {
	testMessage := structures.DNSMessage{
		Header: structures.NewQueryDNSHeader(structures.OpStandardQuery, 0, 1, 0, 0, 0),
		Questions: []*structures.DNSQuestion{
			{
				QName:  "urfu.ru",
				QType:  structures.QTypeA,
				QClass: structures.QClassIN,
			},
		},
		Answer: nil,
	}

	conn, err := net.Dial("udp", "212.193.66.21:53")
	if err != nil {
		fmt.Println(err)
	}

	_, err = conn.Write(testMessage.Marshal())
	if err != nil {
		fmt.Println(err)
		return
	}

	buffer := make([]byte, 512)
	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	parsedMessage, _ := structures.UnmarshalMessage(buffer)

	fmt.Println(testMessage.Header.Id)
	fmt.Println(parsedMessage.Header.Id)
	fmt.Println(parsedMessage.Header.QDCOUNT)
	fmt.Println(parsedMessage.Header.ANCOUNT)
	fmt.Println(parsedMessage.Header.NSCOUNT)
	fmt.Println(parsedMessage.Header.ARCOUNT)
	fmt.Println(parsedMessage.Questions[0].QName)
	fmt.Println(parsedMessage.Questions[0].QClass)
	fmt.Println(parsedMessage.Questions[0].QType)
	fmt.Println(parsedMessage.Answer[0].Name)
	fmt.Println(parsedMessage.Answer[0].Type)
	fmt.Println(parsedMessage.Authority[0].Name)
	fmt.Println(parsedMessage.Authority[1].Name)
	fmt.Println(parsedMessage.Additional[0].Name)
}
