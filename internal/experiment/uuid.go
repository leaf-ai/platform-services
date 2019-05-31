package experiment

// This file contains an interim implementation for UUID V4 until such time
// as the google golang implementation arrives at 1.0 from https://github.com/google/uuid

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"
)

const hexDigit = "0123456789abcdef"

func GetMacAddrWithoutDelimiters() (macAddr string, err errors.Error) {
	interfaces, errGo := net.Interfaces()
	if errGo != nil {
		return "", errors.Wrap(errGo, "machine could not be queried for network interfaces").With("stack", stack.Trace().TrimRuntime())
	}
	for _, anInterface := range interfaces {
		if 0 != anInterface.Flags&net.FlagLoopback {
			continue
		}
		if len(anInterface.HardwareAddr) != 0 {
			// convert each byte into 2 hex characters,
			// so set aside *2 size in the resulting buffer
			buffer := make([]byte, 0, len(anInterface.HardwareAddr)*2)
			for _, byte := range anInterface.HardwareAddr {
				buffer = append(buffer, hexDigit[byte>>4])
				buffer = append(buffer, hexDigit[byte&0xF])
			}
			return string(buffer), nil
		}
	}
	return "", errors.Wrap(errGo, "no MAC address could be found on the available network interfaces").With("stack", stack.Trace().TrimRuntime())
}

// GetComponentUniqueID produces a non RFC 4122 UUID that is tied to the MAC address if
// available for is generated randomly if not
//
func GetComponentUniqueID(component string) (id string) {

	macHwAddr, err := GetMacAddrWithoutDelimiters()
	var buf bytes.Buffer
	if err == nil && len(macHwAddr) != 0 {
		// Prefix component and "M" (to denote MAC Address)
		buf.WriteString(component)
		buf.WriteString("M")
		buf.WriteString(macHwAddr)
		return buf.String()
	}
	// generate a random number and use it as the identifier
	b := make([]byte, 6)
	rand.Read(b)

	buf.WriteString(component)
	buf.WriteString("R")
	buf.WriteString(fmt.Sprintf("%x%x%x%x%x%x", b[0], b[1], b[2], b[3], b[4], b[5]))
	return buf.String()
}

// GetPseudoUUID will generate an RFC 4122 compliant UUID using Version 4
func GetPseudoUUID() (uuid string) {
	// generate a random number and use it as the identifier
	b := make([]byte, 16)
	rand.Read(b)

	b[8] = (b[8] | 0x80) & 0xBF // Clamped values for a range of bytes specified in the RFC
	b[6] = (b[6] | 0x40) & 0x4F

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
