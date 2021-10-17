package helpers

import (
	"bytes"
	"encoding/binary"
	"strings"
)

func ReadLabel(bufferWithLabelStarting *bytes.Buffer, fullMessage []byte) (label string) {
	var parts []string

	for {
		lengthOfName, _ := bufferWithLabelStarting.ReadByte()

		if lengthOfName == 0 {
			break
		}

		if lengthOfName > 63 { // Compressed part
			_ = bufferWithLabelStarting.UnreadByte()
			parts = append(parts, readCompressed(bufferWithLabelStarting, fullMessage))
			continue
		}

		labelBytes := make([]byte, lengthOfName)
		_, _ = bufferWithLabelStarting.Read(labelBytes)

		parts = append(parts, string(labelBytes))
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
	part = ReadLabel(updatedBuffer, fullMessage)
	return
}
