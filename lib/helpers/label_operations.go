package helpers

import (
	"bytes"
	"encoding/binary"
	"log"
	"strings"
)

func ReadLabel(bufferWithLabelStarting *bytes.Buffer, fullMessage []byte) (label string) {
	// TODO: can be optimized with cache ?
	var (
		parts              []string
		hadRealLabelBefore bool
	)

	for {
		lengthOfName, _ := bufferWithLabelStarting.ReadByte()

		if lengthOfName == 0 {
			break
		}

		if lengthOfName > 63 { // Compressed part
			// The compression scheme allows a domain name in a message to be represented as either:
			//   - a sequence of labels ending in a zero octet
			//   - a pointer
			//   - a sequence of labels ending with a pointer

			_ = bufferWithLabelStarting.UnreadByte()
			compressed := readCompressed(bufferWithLabelStarting, fullMessage)
			parts = append(parts, compressed)

			if hadRealLabelBefore {
				break
			}

			continue
		}

		labelBytes := make([]byte, lengthOfName)
		_, _ = bufferWithLabelStarting.Read(labelBytes)

		parts = append(parts, string(labelBytes))
		hadRealLabelBefore = true
	}

	label = strings.Join(parts, ".")
	log.Println(label)
	return
}

func readCompressed(buffer *bytes.Buffer, fullMessage []byte) (part string) {
	twoOctets := make([]byte, 2)
	_, _ = buffer.Read(twoOctets)

	readerTwoOctets := bytes.NewReader(twoOctets)
	var indicatorAndOffset uint16

	_ = binary.Read(readerTwoOctets, binary.BigEndian, &indicatorAndOffset)

	offset := (indicatorAndOffset << 2) >> 2

	updatedBuffer := bytes.NewBuffer(fullMessage[offset:])
	part = ReadLabel(updatedBuffer, fullMessage)
	return
}
