package structures

import "bytes"

type DNSMessage struct {
	Header *DNSHeader

	Questions  []*DNSQuestion
	Answer     []*DNSRecord
	Authority  []*DNSRecord
	Additional []*DNSRecord
}

func (m *DNSMessage) Marshal() (res []byte) {
	buffer := new(bytes.Buffer)

	buffer.Write(m.Header.Marshal())

	for _, question := range m.Questions {
		buffer.Write(question.Marshal())
	}

	//for _, answer := range m.Answer {
	//	buffer.Write(answer.Marshal())
	//}

	res = buffer.Bytes()
	return
}

func UnmarshalMessage(data []byte) (message *DNSMessage, err error) {
	header, unreadData, _ := UnmarshalHeader(data)

	message = &DNSMessage{
		Header: header,
	}

	message.Questions, unreadData = UnmarshalQuestions(unreadData, int(header.QDCOUNT))
	message.Answer, unreadData = UnmarshalRecords(unreadData, data, int(header.ANCOUNT))
	message.Authority, unreadData = UnmarshalRecords(unreadData, data, int(header.NSCOUNT))
	message.Additional, unreadData = UnmarshalRecords(unreadData, data, int(header.ARCOUNT))

	return
}
