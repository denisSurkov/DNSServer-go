package structures

import (
	"DNSServer/lib/helpers"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
)

const HeaderLength = 12

// RequestType A one bit field that specifies whether this message is a query (0), or a response (1).
type RequestType byte

const (
	QRQuery RequestType = iota
	QRResponse
)

// OpcodeType A four bit field that specifies kind of query in this
// message. This value is set by the originator of a query
// and copied into the response.
type OpcodeType byte

const (
	OpStandardQuery OpcodeType = iota
	OpInverseQuery
	OpServerStatusRequest
)

type DNSHeader struct {
	/*
		DNSHeader as specified in
		https://datatracker.ietf.org/doc/html/rfc1035#section-4.1.1
	*/

	// A 16 bit identifier assigned by the program that
	// generates any kind of query.  This identifier is copied
	// the corresponding reply and can be used by the requester
	// to match up replies to outstanding queries.
	Id uint16

	QR RequestType

	Opcode OpcodeType

	// Authoritative Answer - this bit is valid in responses,
	// and specifies that the responding name server is an
	// authority for the domain name in question section.
	AA byte

	// TrunCation - specifies that this message was truncated
	// due to length greater than that permitted on the
	// transmission channel.
	TC byte

	// Recursion Desired - this bit may be set in a query and
	// is copied into the response.  If RD is set, it directs
	// the name server to pursue the query recursively.
	// Recursive query support is optional.
	RD byte

	// Recursion Available - this be is set or cleared in a
	// response, and denotes whether recursive query support is
	// available in the name server.
	RA byte

	// Reserved for future use.  Must be zero in all queries
	// and responses.
	Z byte

	// Response code - this 4 bit field is set as part of
	// responses
	RCODE byte

	// an unsigned 16 bit integer specifying the number of
	// entries in the question section.
	QDCOUNT uint16

	// an unsigned 16 bit integer specifying the number of
	// resource records in the answer section.
	ANCOUNT uint16

	// an unsigned 16 bit integer specifying the number of name
	// server resource records in the authority records
	// section.
	NSCOUNT uint16

	// an unsigned 16 bit integer specifying the number of
	// resource records in the additional records section.
	ARCOUNT uint16
}

type marshaledHeaderPacket struct {
	Id                                 uint16
	FirstPartOfFlags                   byte
	SecondPartOfFlags                  byte
	Qdcount, Ancount, Nscount, Arcount uint16
}

func (d *DNSHeader) Marshal() (res []byte) {
	/*
			The header contains the following fields:

		                                    1  1  1  1  1  1
		      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		    |                      ID                       |
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|          firstFlags      |      secondFlags   | <- byte variables
		    |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		    |                    QDCOUNT                    |
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		    |                    ANCOUNT                    |
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		    |                    NSCOUNT                    |
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		    |                    ARCOUNT                    |
		    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/

	buffer := new(bytes.Buffer)

	_ = binary.Write(buffer, binary.BigEndian, d.Id)

	var (
		firstFlags, secondFlags byte
	)

	firstFlags = byte(d.QR)
	firstFlags = helpers.AppendToByteFromRight(firstFlags, 4, byte(d.Opcode))
	firstFlags = helpers.AppendToByteFromRight(firstFlags, 1, byte(d.AA))
	firstFlags = helpers.AppendToByteFromRight(firstFlags, 1, d.TC)
	firstFlags = helpers.AppendToByteFromRight(firstFlags, 1, d.RD)
	firstFlags = helpers.AppendToByteFromRight(firstFlags, 1, d.RA)

	buffer.WriteByte(firstFlags)

	secondFlags = d.Z
	secondFlags = helpers.AppendToByteFromRight(firstFlags, 4, d.RCODE)

	buffer.WriteByte(secondFlags)

	_ = binary.Write(buffer, binary.BigEndian, d.QDCOUNT)
	_ = binary.Write(buffer, binary.BigEndian, d.ANCOUNT)
	_ = binary.Write(buffer, binary.BigEndian, d.NSCOUNT)
	_ = binary.Write(buffer, binary.BigEndian, d.ARCOUNT)

	res = buffer.Bytes()
	return
}

func UnmarshalHeader(data []byte) (header *DNSHeader, unreadData []byte, err error) {
	headerReader := bytes.NewReader(data[:HeaderLength])

	var packet marshaledHeaderPacket
	err = binary.Read(headerReader, binary.BigEndian, &packet)
	if err != nil {
		fmt.Println(err)
	}

	qr, opcode, aa, tc, rd, ra := parseFirstPartOfFlags(packet.FirstPartOfFlags)
	z, rcode := parseSecondPartOfFlags(packet.SecondPartOfFlags)

	header = &DNSHeader{
		Id:      packet.Id,
		QR:      RequestType(qr),
		Opcode:  OpcodeType(opcode),
		AA:      aa,
		TC:      tc,
		RD:      rd,
		RA:      ra,
		Z:       z,
		RCODE:   rcode,
		QDCOUNT: packet.Qdcount,
		ANCOUNT: packet.Ancount,
		NSCOUNT: packet.Nscount,
		ARCOUNT: packet.Arcount,
	}

	unreadData = data[HeaderLength:]
	return
}

func parseFirstPartOfFlags(firstPartOfFlags byte) (qr, opcode, aa, tc, rd, ra byte) {
	reversed := bits.Reverse8(firstPartOfFlags)

	qr, reversed = helpers.ReadLastNBitsAndShift(reversed, 1)
	opcode, reversed = helpers.ReadLastNBitsAndShift(reversed, 4)
	aa, reversed = helpers.ReadLastNBitsAndShift(reversed, 1)
	tc, reversed = helpers.ReadLastNBitsAndShift(reversed, 1)
	rd, reversed = helpers.ReadLastNBitsAndShift(reversed, 1)
	ra, _ = helpers.ReadLastNBitsAndShift(reversed, 1)

	return
}

func parseSecondPartOfFlags(secondPartOfFlags byte) (z, rcode byte) {
	reversed := bits.Reverse8(secondPartOfFlags)
	z, reversed = helpers.ReadLastNBitsAndShift(reversed, 3)
	rcode, _ = helpers.ReadLastNBitsAndShift(reversed, 4)

	return
}
