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

func NewDNSMessage(header *DNSHeader, questions []*DNSQuestion, answer []*DNSRecord, authority []*DNSRecord, additional []*DNSRecord) *DNSMessage {
	return &DNSMessage{Header: header, Questions: questions, Answer: answer, Authority: authority, Additional: additional}
}

func NewQueryDNSMessage(questions ...*DNSQuestion) *DNSMessage {
	header := NewDNSQuestionHeader()
	header.QDCOUNT = uint16(len(questions))
	return NewDNSMessage(header, questions, nil, nil, nil)
}

func NewAnswerDNSMessage(questions []*DNSQuestion, answers []*DNSRecord) *DNSMessage {
	header := NewDNSAnswerHeader()
	header.QDCOUNT = uint16(len(questions))
	header.ANCOUNT = uint16(len(answers))
	return NewDNSMessage(header, questions, answers, nil, nil)
}

func (m *DNSMessage) Marshal() (res []byte) {
	buffer := new(bytes.Buffer)
	namePositions := make(map[string]int)

	buffer.Write(m.Header.Marshal())

	for _, question := range m.Questions {
		namePositions[question.QName] = buffer.Len() - 1
		buffer.Write(question.Marshal())
	}

	for _, answer := range m.Answer {
		buffer.Write(answer.Marshal(namePositions))
	}

	for _, additional := range m.Additional {
		buffer.Write(additional.Marshal(namePositions))
	}

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
