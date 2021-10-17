package helpers

import "math"

func AppendToByteFromRight(original byte, lengthOfByteToAppend int, byteToAppend byte) byte {
	original <<= lengthOfByteToAppend
	original |= byteToAppend

	return original
}

func ReadLastNBitsAndShift(original byte, lengthToRead int) (nBits byte, shiftedOriginal byte) {
	if lengthToRead <= 0 {
		panic("length is lower or equal than zero")
	}

	nBits = original & uint8(math.Pow(2, float64(lengthToRead))-1)
	shiftedOriginal >>= lengthToRead
	return
}
