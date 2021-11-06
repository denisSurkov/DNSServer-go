package structures

import (
	"DNSServer/lib/helpers"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// RecordType  two octets containing one of the RR TYPE codes.
type RecordType uint16

const (
	_ RecordType = iota

	RecordTypeA     // 1 a host address
	RecordTypeNS    // 2 an authoritative name server
	RecordTypeMD    // 3 a mail destination (Obsolete - use MX)
	RecordTypeMF    // 4 a mail forwarder (Obsolete - use MX)
	RecordTypeCNAME // 5 the canonical name for an alias
	RecordTypeSOA   // 6 marks the start of a zone of authority
	RecordTypeMB    // 7 a mailbox domain name (EXPERIMENTAL)
	RecordTypeMG    // 8 a mail group member (EXPERIMENTAL)
	RecordTypeMR    // 9 a mail rename domain name (EXPERIMENTAL)
	RecordTypeNULL  // 10 a null RR (EXPERIMENTAL)
	RecordTypeWKS   // 11 a well known service description
	RecordTypePTR   // 12 a domain name pointer
	RecordTypeHINFO // 13 host information
	RecordTypeMINFO // 14 mailbox or mail list information
	RecordTypeMX    // 15 mail exchange
	RecordTypeTXT   // 16 text strings

	RecordTypeAAAA = 28
	RecordTypeOPT  = 41
)

type RecordClass uint16

const (
	_ RecordClass = iota

	RecordClassIN //  1 the Internet
	RecordClassCS //  2 the CSNET class (Obsolete - used only for examples in some obsolete RFCs)
	RecordClassCH //  3 the CHAOS class
	RecordClassHS //  4 Hesiod [Dyer 87]
)

type DNSRecord struct {
	Name string

	Type RecordType

	// two octets which specify the class of the data in the
	//                RDATA field.
	Class RecordClass

	// a 32 bit unsigned integer that specifies the time
	//                interval (in seconds) that the resource record may be
	//                cached before it should be discarded.  Zero values are
	//                interpreted to mean that the RR can only be used for the
	//                transaction in progress, and should not be cached.
	TimeToLive uint32

	// an unsigned 16 bit integer that specifies the length in
	//                octets of the RDATA field.
	RDLENGTH uint16

	// a variable length string of octets that describes the
	//                resource.  The format of this information varies
	//                according to the TYPE and CLASS of the resource record.
	//                For example, the if the TYPE is A and the CLASS is IN,
	//                the RDATA field is a 4 octet ARPA Internet address.
	RDATA []byte

	// Here I store parsed correctly RDATA based on record Type
	RDataRepresentation string
}

type marshaledRecordPacket struct {
	_    byte
	Type byte

	_     byte
	Class byte

	TTL      uint32
	RDLength uint16
}

func (r *DNSRecord) Marshal(namesPositions map[string]int) (res []byte) {
	//                                1  1  1  1  1  1
	//      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//    |                                               |
	//    /                                               /
	//    /                      NAME                     /
	//    |                                               |
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//    |                      TYPE                     |
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//    |                     CLASS                     |
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//    |                      TTL                      |
	//    |                                               |
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//    |                   RDLENGTH                    |
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
	//    /                     RDATA                     /
	//    /                                               /
	//    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

	buffer := new(bytes.Buffer)

	position, ok := namesPositions[r.Name]
	if ok {
		twoOctets := uint16(0) << 16
		twoOctets |= uint16(position)
		_ = binary.Write(buffer, binary.BigEndian, twoOctets)
	} else {
		buffer.WriteByte(uint8(len(r.Name)))
		buffer.WriteString(r.Name)
	}

	_ = binary.Write(buffer, binary.BigEndian, uint16(r.Type))
	_ = binary.Write(buffer, binary.BigEndian, uint16(r.Class))
	_ = binary.Write(buffer, binary.BigEndian, r.TimeToLive)
	_ = binary.Write(buffer, binary.BigEndian, r.RDLENGTH)

	buffer.Write(r.RDATA)

	res = buffer.Bytes()
	return res
}

func UnmarshalRecords(recordsStartBytes, fullMessage []byte, recordsCount int) (records []*DNSRecord, unreadData []byte) {
	buffer := bytes.NewBuffer(recordsStartBytes)

	for recordsCount > 0 {
		name, shouldUnreadByte := helpers.ReadLabel(buffer, fullMessage)

		if shouldUnreadByte {
			_ = buffer.UnreadByte()
		}

		var packet marshaledRecordPacket

		foo := make([]byte, 2+2+4+2)
		_, _ = buffer.Read(foo)

		packetReader := bytes.NewReader(foo)
		_ = binary.Read(packetReader, binary.BigEndian, &packet)

		log.Println(packet)

		rdata := make([]byte, packet.RDLength)

		_, _ = buffer.Read(rdata)

		currentRecord := &DNSRecord{
			Name:       name,
			Type:       RecordType(packet.Type),
			Class:      RecordClass(packet.Class),
			TimeToLive: packet.TTL,
			RDLENGTH:   packet.RDLength,
			RDATA:      rdata,
		}

		var rDataRepresentation string
		switch currentRecord.Type {
		case RecordTypeA:
			rDataRepresentation = unmarshalRDataAsIpv4Address(currentRecord)
		case RecordTypeNS:
			rDataRepresentation = unmarshalRDataAsNS(currentRecord, fullMessage)
		case RecordTypeOPT, RecordTypeAAAA:
		default:
			fmt.Print(hex.Dump(fullMessage))
			fmt.Println("--------------")
			fmt.Println(hex.Dump(recordsStartBytes))
			fmt.Println("--------------")
			fmt.Println(hex.Dump(foo))
			fmt.Println("--------------")
			fmt.Print(currentRecord.Type)
			log.Fatal("UNKNOWN TYPE!!")
			rDataRepresentation = "hello yopta"
		}

		currentRecord.RDataRepresentation = rDataRepresentation
		records = append(records, currentRecord)

		recordsCount -= 1
	}

	unreadData = buffer.Bytes()
	return
}

func unmarshalRDataAsIpv4Address(r *DNSRecord) string {
	numbers := make([]string, 4)

	for i, octet := range r.RDATA {
		numbers[i] = strconv.Itoa(int(octet))
	}

	return strings.Join(numbers, ".")
}

func unmarshalRDataAsNS(r *DNSRecord, fullMessage []byte) string {
	buffer := bytes.NewBuffer(r.RDATA)
	name, _ := helpers.ReadLabel(buffer, fullMessage)

	return name
}
