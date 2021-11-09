package structures

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// QType is a two octet code which specifies the type of the query.
// The values for this field include all codes valid for a
// TYPE field, together with some more general codes which
// can match more than one type of RR.
type QType uint16

const (
	_ QType = iota

	QTypeA     // 1 a host address
	QTypeNS    // 2 an authoritative name server
	QTypeMD    // 3 a mail destination (Obsolete - use MX)
	QTypeMF    // 4 a mail forwarder (Obsolete - use MX)
	QTypeCNAME // 5 the canonical name for an alias
	QTypeSOA   // 6 marks the start of a zone of authority
	QTypeMB    // 7 a mailbox domain name (EXPERIMENTAL)
	QTypeMG    // 8 a mail group member (EXPERIMENTAL)
	QTypeMR    // 9 a mail rename domain name (EXPERIMENTAL)
	QTypeNULL  // 10 a null RR (EXPERIMENTAL)
	QTypeWKS   // 11 a well known service description
	QTypePTR   // 12 a domain name pointer
	QTypeHINFO // 13 host information
	QTypeMINFO // 14 mailbox or mail list information
	QTypeMX    // 15 mail exchange
	QTypeTXT   // 16 text strings

	QTypeAXFR  = 252 // A request for a transfer of an entire zone
	QTypeMAILB = 253 // A request for mailbox-related records (MB, MG or MR)
	QTypeMAILA = 254 // A request for mail agent RRs (Obsolete - see MX)
	QTypeALL   = 255 // A request for all records
)

// QClass a two octet code that specifies the class of the query.
// For example, the QCLASS field is IN for the Internet.
type QClass uint16

const (
	_        QClass = iota
	QClassIN        //  1 the Internet
	QClassCS        //  2 the CSNET class (Obsolete - used only for examples in some obsolete RFCs)
	QClassCH        //  3 the CHAOS class
	QClassHS        //  4 Hesiod [Dyer 87]

	QClassALL = 255 // any class
)

type DNSQuestion struct {
	/*
		DNSQuestion as specified in
		https://datatracker.ietf.org/doc/html/rfc1035#section-4.1.2
	*/

	// QName is a domain name represented as a sequence of labels, where
	// each label consists of a length octet followed by that
	// number of octets.  The domain name terminates with the
	// zero length octet for the null label of the root.  Note
	// that this field may be an odd number of octets; no
	// padding is used.
	QName string

	QType  QType
	QClass QClass
}

func NewDNSQuestion(QName string, QType QType, QClass QClass) *DNSQuestion {
	return &DNSQuestion{QName: QName, QType: QType, QClass: QClass}
}

type marshaledQuestionPacket struct {
	QType  uint16
	QClass uint16
}

func (question *DNSQuestion) Marshal() (res []byte) {
	buffer := new(bytes.Buffer)

	writeQName(buffer, question.QName)

	_ = binary.Write(buffer, binary.BigEndian, uint16(question.QType))
	_ = binary.Write(buffer, binary.BigEndian, uint16(question.QClass))

	res = buffer.Bytes()
	return
}

func UnmarshalQuestions(questionStartData []byte, questionsCount int) (questions []*DNSQuestion, unparsedData []byte) {
	buffer := bytes.NewBuffer(questionStartData)

	for questionsCount > 0 {
		fullLabel := parseLabel(buffer)

		flagsReader := bytes.NewReader(buffer.Next(4))
		var packet marshaledQuestionPacket
		_ = binary.Read(flagsReader, binary.BigEndian, &packet)

		questions = append(questions, &DNSQuestion{
			QName:  fullLabel,
			QType:  QType(packet.QType),
			QClass: QClass(packet.QClass),
		})
		questionsCount -= 1
	}

	unparsedData = buffer.Bytes()
	return
}

func writeQName(buffer *bytes.Buffer, qName string) {
	labels := strings.Split(qName, ".")
	for _, label := range labels {
		buffer.WriteByte(uint8(len(label)))
		buffer.WriteString(label)
	}
	buffer.WriteByte(0)
}

func parseLabel(buffer *bytes.Buffer) (label string) {
	var (
		lengthOfCurrentLabel uint8
		currentLabel         string
		labels               []string
	)

	for {
		lengthOfCurrentLabel, _ = buffer.ReadByte()
		if lengthOfCurrentLabel == 0 {
			break
		}

		labelBytes := make([]byte, lengthOfCurrentLabel)
		_, _ = buffer.Read(labelBytes)

		currentLabel = string(labelBytes)
		labels = append(labels, currentLabel)
	}

	label = strings.Join(labels, ".")
	return
}
