package lib

import (
	"DNSServer/lib/structures"
	"bufio"
	"log"
	"net"
	"time"
)

func Resolve(incomingRequest *IncomingRequest, conn net.PacketConn) {
	foundAnswers, lastMessage, nextServersToAsk := innerResolve(incomingRequest, RootIPServers...)

	for !foundAnswers {
		foundAnswers, lastMessage, nextServersToAsk = innerResolve(incomingRequest, nextServersToAsk...)
	}

	log.Printf("found! %s", lastMessage)
	_, _ = conn.WriteTo(lastMessage.Marshal(), incomingRequest.Address)
}

func innerResolve(baseIncomingRequest *IncomingRequest, nextServersToAsk ...string) (
	foundAnswers bool, lastReceivedMsg *structures.DNSMessage, nextServersToAskAgain []string) {
	baseIncomingRequest.DNSMessage.Header.ARCOUNT = 0 // TODO: fix it
	marshaledIncomingRequest := baseIncomingRequest.DNSMessage.Marshal()

	_, ans, succeeded := tryToRetrieveDNSDataFromServers(marshaledIncomingRequest, 1, "udp", nextServersToAsk...)

	if !succeeded {
		log.Fatalf("failed to retrieve dns data from servers")
	}

	msg, err := structures.UnmarshalMessage(ans)

	if err != nil {
		log.Fatalf("error while unmarshallind root answer err = %s", err)
	}

	if msg.Header.TC == 1 {
		log.Println("message is to big, have to make TCP call")
		retrievedFrom, data, succeeded := tryToRetrieveDNSDataFromServers(marshaledIncomingRequest, 1, "tcp", nextServersToAsk...)

		if !succeeded {
			log.Fatalf("didnt succeed with retrieving data")
		}

		log.Printf("retrieved from %s", retrievedFrom)
		msg, _ = structures.UnmarshalMessage(data)
	}

	if msg.Header.ANCOUNT >= baseIncomingRequest.DNSMessage.Header.QDCOUNT {
		log.Println("found full answer count for incoming questions count")
		foundAnswers = true
		lastReceivedMsg = msg
		return
	}

	log.Printf("answer is not full, checking another, msg %d", msg.Header.ANCOUNT)
	nsWithIps := collectAllIPAuthorityNSFromAdditional(msg)

	for ns, ip := range nsWithIps {
		if len(ip) == 0 {
			continue
		}
		log.Printf("adding new server name=%s, ip=%s to ask", ns, ip)
		nextServersToAskAgain = append(nextServersToAskAgain, ip)
	}

	foundAnswers = false
	return
}

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

			data, err := makeUDPCallDNS(server, dialType, message)
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

func makeUDPCallDNS(ipAddressWithoutPort, dialType string, message []byte) (buffer []byte, err error) {
	buffer = make([]byte, 1048)

	ipAddressWithCorrectPort := ipAddressWithoutPort + ":53"
	conn, err := net.Dial(dialType, ipAddressWithCorrectPort)

	if err != nil {
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	_, err = conn.Write(message)

	if err != nil {
		return
	}

	_, err = bufio.NewReader(conn).Read(buffer)
	return
}

func collectAllIPAuthorityNSFromAdditional(fromMessage *structures.DNSMessage) map[string]string {
	nsNamesWithIps := make(map[string]string)

	for _, authorityNsRecord := range fromMessage.Authority {
		nsNamesWithIps[authorityNsRecord.RDataRepresentation] = ""
		log.Println(authorityNsRecord.RDataRepresentation)
	}

	for _, additionalRecord := range fromMessage.Additional {
		if additionalRecord.Type == structures.RecordTypeA &&
			additionalRecord.Class == structures.RecordClassIN {

			log.Println(additionalRecord.Name)
			_, ok := nsNamesWithIps[additionalRecord.Name]
			log.Println(ok)
			if !ok {
				continue
			}

			nsNamesWithIps[additionalRecord.Name] = additionalRecord.RDataRepresentation
		}
	}

	return nsNamesWithIps
}
