package helpers

import (
	"bytes"
	"encoding/binary"
	"strings"
)

func WriteLabel(buffer *bytes.Buffer, label string) {
	labels := strings.Split(label, ".")
	for _, label := range labels {
		buffer.WriteByte(uint8(len(label)))
		buffer.WriteString(label)
	}
	buffer.WriteByte(0)
}

func ReadLabel(bufferWithLabelStarting *bytes.Buffer, fullMessage []byte) (label string, shouldUnreadByte bool) {
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
			shouldUnreadByte = true
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
	part, _ = ReadLabel(updatedBuffer, fullMessage)
	return
}
