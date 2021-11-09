package lib

import (
	"bytes"
	"log"
	"net"
)

func tryToRetrieveDNSDataFromServers(
	message []byte,
	attemptCountForOne int,
	dialType string,
	servers ...string) (
	retrievedFromServer string,
	data []byte,
	succeeded bool,
) {

	log.Println(len(servers))
	for _, server := range servers {
		currentAttempt := 1

		for currentAttempt <= attemptCountForOne {
			log.Printf("making %s call to server %s", dialType, server)

			data, err := makeNetDNSCall(server, dialType, message)
			if err != nil {
				log.Printf("error %s while trying to make %s call for server %s, attempt %d",
					err, dialType, server, currentAttempt)
				currentAttempt += 1
				continue
			}

			log.Printf("succeded making %s call to server %s", dialType, server)
			return server, data, true
		}
	}

	return "", nil, false
}

func makeNetDNSCall(ipAddressWithoutPort, dialType string, message []byte) (buffer []byte, err error) {
	ipAddressWithCorrectPort := ipAddressWithoutPort + ":53"

	log.Printf("making %s call to %s", dialType, ipAddressWithCorrectPort)
	conn, err := net.Dial(dialType, ipAddressWithCorrectPort)

	if err != nil {
		log.Printf("error while making %s call to %s, error %s", dialType, ipAddressWithCorrectPort, err)
		return
	}

	if err != nil {
		log.Fatalf("%s", err)
	}
	_, err = conn.Write(message)
	if err != nil {
		log.Printf("error while writing as %s to %s, error %s", dialType, ipAddressWithCorrectPort, err)
		return
	}

	if dialType == "udp" {
		buffer = make([]byte, 1024)
		_, err = conn.Read(buffer)
	} else {
		tcpBuffer := bytes.NewBuffer(make([]byte, 1024))
		n := 1025
		totalLength := 0
		for n >= 1024 {
			tempBuff := make([]byte, 1024)
			n, err = conn.Read(tempBuff)
			tcpBuffer.Write(tempBuff)
			totalLength += n
			tcpBuffer.Grow(1024)
		}
		buffer = make([]byte, totalLength)
		_, _ = tcpBuffer.Read(buffer)
	}

	defer func() {
		_ = conn.Close()
	}()

	return
}
