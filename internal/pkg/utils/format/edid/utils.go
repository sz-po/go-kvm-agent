package edid

import (
	"errors"
	"strings"
)

func EncodeManufacturer(code string) ([2]byte, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 3 {
		return [2]byte{}, errors.New("manufacturer code must be exactly 3 letters")
	}

	v := uint16(code[0]-'A'+1)<<10 |
		uint16(code[1]-'A'+1)<<5 |
		uint16(code[2]-'A'+1)

	return [2]byte{byte(v >> 8), byte(v & 0xFF)}, nil
}

func DecodeManufacturer(b [2]byte) (string, error) {
	v := uint16(b[0])<<8 | uint16(b[1])

	l1 := byte((v>>10)&0x1F) - 1 + 'A'
	l2 := byte((v>>5)&0x1F) - 1 + 'A'
	l3 := byte(v&0x1F) - 1 + 'A'

	if l1 < 'A' || l1 > 'Z' || l2 < 'A' || l2 > 'Z' || l3 < 'A' || l3 > 'Z' {
		return "", errors.New("decoded value is not a valid manufacturer code")
	}

	return string([]byte{l1, l2, l3}), nil
}
