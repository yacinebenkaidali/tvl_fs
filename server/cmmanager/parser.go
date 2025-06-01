package cmmanager

import (
	"encoding/binary"
	"io"
)

func handleTypeParsing(r io.Reader) (uint16, error) {
	// READING_TYPE
	typeBytes := make([]byte, 2)
	_, err := io.ReadFull(r, typeBytes)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(typeBytes), nil
}

func handleLengthParsing(r io.Reader, expectedLgth int) (uint32, error) {
	// READING_LENGTH
	lengthBytes := make([]byte, expectedLgth)
	_, err := io.ReadFull(r, lengthBytes)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(lengthBytes), nil
}

func handleFileNameParsing(r io.Reader, fileNameLgth int) (string, error) {
	// READING_VALUE
	fileNameBytes := make([]byte, fileNameLgth)
	n, err := io.ReadFull(r, fileNameBytes)
	if err != nil {
		return "", err
	}

	return string(fileNameBytes[:n]), nil
}
