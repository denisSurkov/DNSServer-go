package lib

import (
	"DNSServer/lib/structures"
	"log"
	"net"
	"sync"
)

var sendMutex sync.Mutex
var cache = structures.NewQueryCache()

func Resolve(incomingRequest *IncomingRequest, conn net.PacketConn) {
	// dig sends weird (name Root, type OPT) stuff, so I ignore it
	incomingRequest.DNSMessage.Header.ARCOUNT = 0
	incomingRequest.DNSMessage.Additional = nil

	answer := resolveQueryDNS(incomingRequest.DNSMessage)
	makeAnswerLookLikeThisDNSServerSendIt(answer, incomingRequest.DNSMessage)

	sendMutex.Lock()
	_, _ = conn.WriteTo(answer.Marshal(), incomingRequest.Address)
	sendMutex.Unlock()
}

func resolveQueryDNS(queryMessage *structures.DNSMessage) *structures.DNSMessage {
	answers, cacheFound := askCache(queryMessage)
	log.Printf("asked cache? %t", cacheFound)
	if cacheFound {
		return structures.NewAnswerDNSMessage(queryMessage.Questions, answers)
	}

	foundAnswer, lastMessage := askDNS(queryMessage, RootIPServers...)

	for !foundAnswer {
		nextServersToAsk := collectNamespaceIp(lastMessage)
		foundAnswer, lastMessage = askDNS(queryMessage, nextServersToAsk...)
	}

	log.Println("adding cache")
	setCache(queryMessage, lastMessage)
	return lastMessage
}

func askCache(queryMessage *structures.DNSMessage) ([]*structures.DNSRecord, bool) {
	question := queryMessage.Questions[0]
	return cache.Get(question)
}

func setCache(originalMessage *structures.DNSMessage, answerMessage *structures.DNSMessage) {
	question := originalMessage.Questions[0]
	cache.Set(question, answerMessage.Answer)
}

func askDNS(queryMessage *structures.DNSMessage, serversToAsk ...string) (
	foundAnswers bool, lastReceivedMsg *structures.DNSMessage) {
	marshaledIncomingRequest := queryMessage.Marshal()

	_, ans, succeeded := tryToRetrieveDNSDataFromServers(marshaledIncomingRequest, 1, "udp", serversToAsk...)
	if !succeeded {
		log.Fatalf("Failed to receive dns data from all servers")
	}

	lastReceivedMsg, err := structures.UnmarshalMessage(ans)
	if err != nil {
		log.Fatalf("error while unmarshallind root answer err = %s", err)
	}

	// for some reason, dns servers declined to use tcp..
	// i always got EOF while making requests over tcp

	//if msg.Header.TC == 1 {
	//	log.Println("message is to big, have to make TCP call")
	//	retrievedFrom, data, succeeded := tryToRetrieveDNSDataFromServers(marshaledIncomingRequest, 1, "tcp", serversToAsk...)
	//
	//	if !succeeded {
	//		log.Fatalf("didnt succeed with retrieving data")
	//	}
	//
	//	log.Printf("retrieved from %s", retrievedFrom)
	//	msg, _ = structures.UnmarshalMessage(data)
	//}

	if lastReceivedMsg.Header.ANCOUNT >= queryMessage.Header.QDCOUNT {
		log.Println("found full answer count for incoming questions count")
		foundAnswers = true
		return
	}

	log.Printf("answer count %d is lower than query message", lastReceivedMsg.Header.ANCOUNT)
	foundAnswers = false
	return
}

func collectNamespaceIp(fromMessage *structures.DNSMessage) (namespaceIp []string) {
	nsWithIps := collectAllIPAuthorityNSFromAdditional(fromMessage)

	var allNamespaces []string
	for ns, ip := range nsWithIps {
		allNamespaces = append(allNamespaces, ns)

		if len(ip) == 0 {
			continue
		}

		log.Printf("adding new server name=%s, ip=%s to ask", ns, ip)
		namespaceIp = append(namespaceIp, ip)
	}

	if namespaceIp == nil {
		nsWithIps = retrieveNameserversIps(allNamespaces[0])

		for _, ip := range nsWithIps {
			namespaceIp = append(namespaceIp, ip)
		}
	}

	return
}

func collectAllIPAuthorityNSFromAdditional(fromMessage *structures.DNSMessage) map[string]string {
	nsNamesWithIps := make(map[string]string)

	for _, authorityNsRecord := range fromMessage.Authority {
		if authorityNsRecord.Type == structures.RecordTypeNS &&
			authorityNsRecord.Class == structures.RecordClassIN {
			nsNamesWithIps[authorityNsRecord.RDataRepresentation] = ""
		}
	}

	for _, additionalRecord := range fromMessage.Additional {
		if additionalRecord.Type == structures.RecordTypeA &&
			additionalRecord.Class == structures.RecordClassIN {

			_, ok := nsNamesWithIps[additionalRecord.Name]
			if !ok {
				continue
			}

			nsNamesWithIps[additionalRecord.Name] = additionalRecord.RDataRepresentation
		}
	}

	return nsNamesWithIps
}

func retrieveNameserversIps(nameserversToRetrieve ...string) map[string]string {
	var nsNamesWithIps = make(map[string]string)

	for _, name := range nameserversToRetrieve {
		// https://stackoverflow.com/a/4083071
		// "No one support multiply questions in DNS Message today"

		currentQuestion := structures.NewDNSQuestion(name, structures.QTypeA, structures.QClassIN)
		message := structures.NewQueryDNSMessage(currentQuestion)

		answerMessage := resolveQueryDNS(message)

		for _, answer := range answerMessage.Answer {
			nsNamesWithIps[answer.Name] = answer.RDataRepresentation
		}
	}

	return nsNamesWithIps
}

func makeAnswerLookLikeThisDNSServerSendIt(answer *structures.DNSMessage,
	originalMessage *structures.DNSMessage) {
	answer.Header.Id = originalMessage.Header.Id
	answer.Header.AA = 0
	answer.Header.RA = 1
	answer.Header.RD = originalMessage.Header.RD
	answer.Header.QR = structures.QRResponse
}
