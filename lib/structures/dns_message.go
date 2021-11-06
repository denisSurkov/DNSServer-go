package structures

import (
	"bytes"
)

type DNSMessage struct {
	Header *DNSHeader

	Questions  []*DNSQuestion
	Answer     []*DNSRecord
	Authority  []*DNSRecord
	Additional []*DNSRecord
}

func (m *DNSMessage) Marshal() (res []byte) {
	buffer := new(bytes.Buffer)
	namePositions := make(map[string]int)

	buffer.Write(m.Header.Marshal())

	for _, question := range m.Questions {
		namePositions[question.QName] = buffer.Len() - 1
		buffer.Write(question.Marshal())
	}

	//for _, answer := range m.Answer {
	//	buffer.Write(answer.Marshal(namePositions))
	//}
	//
	//for _, additional := range m.Additional {
	//	buffer.Write(additional.Marshal(namePositions))
	//}

	res = buffer.Bytes()
	return
}

func UnmarshalMessage(data []byte) (message *DNSMessage, err error) {
	header, unreadData, _ := UnmarshalHeader(data)
	message = &DNSMessage{
		Header: header,
	}

	questionsCount := int(header.QDCOUNT)
	if questionsCount >= 1 {
		message.Questions, unreadData = UnmarshalQuestions(unreadData, questionsCount)
	}

	answersCount := int(header.ANCOUNT)
	if answersCount >= 1 {
		message.Answer, unreadData = UnmarshalRecords(unreadData, data, answersCount)
	}

	authorityCount := int(header.NSCOUNT)
	if authorityCount >= 1 {
		message.Authority, unreadData = UnmarshalRecords(unreadData, data, authorityCount)
	}

	additionalCount := int(header.ARCOUNT)
	if additionalCount >= 1 {
		message.Additional, unreadData = UnmarshalRecords(unreadData, data, additionalCount)
	}

	return
}
