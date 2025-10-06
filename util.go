package common

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

func ConvertBoolArrayToByteArray(bools []bool) []byte {
	byteArray := make([]byte, (len(bools)+7)/8) // Each byte can hold up to 8 bools, so we round up

	for i, b := range bools {
		if b {
			byteArray[i/8] |= 1 << (i % 8) // Set the bit in the correct position
		}
	}

	return byteArray
}

func (t *TrackingID) BitLen() int {
	return len(t.PartiesState) * 8
}

// Will panic if i is out of bounds
func (t *TrackingID) PartyStateOk(i int) bool {
	// Find the index of the byte containing the bit
	byteIndex := i / 8
	// Find the position of the bit within the byte
	bitPosition := uint(i % 8)

	// Use bitwise AND to check if the specific bit is set
	// we check for != 0 since it can be different byte values (depending on the bit position)
	return t.PartiesState[byteIndex]&(1<<bitPosition) != 0
}

// ConvertByteArrayToBoolArray converts a packed []byte back to a []bool.
func ConvertByteArrayToBoolArray(byteArray []byte, numBools int) []bool {
	bools := make([]bool, numBools)

	for i := 0; i < numBools; i++ {
		bools[i] = (byteArray[i/8] & (1 << (i % 8))) != 0 // Check if the bit is set
	}

	return bools
}

const nilTrackID = "nilTrackID"

// Creates a byte-string representation of the TrackingID.
// output is in the format "ProtocolType-Digest-PartiesState-AuxiliaryData".
// protocolType is an integer corresponding to the ProtocolType enum.
// Each part of Digest-PartiesState-AuxiliaryData is a hexadecimal representation of the respective byte slice.
// If the TrackingID is nil, returns "nilTrackID".
func (t *TrackingID) ToString() string {
	if t == nil {
		return nilTrackID
	}

	return fmt.Sprintf("%d-%x-%x-%x", t.Protocol, pad32(t.Digest), t.PartiesState, t.AuxiliaryData)
}

func (x *TrackingID) ToByteString() []byte {
	return []byte(x.ToString())
}

var (
	errNilTrackID                  = fmt.Errorf("nil TrackingID")
	errTrackidPartTooLong          = fmt.Errorf("TrackingID part too long, must be at most 64 characters (32 bytes) each")
	errTrackidMustHaveDigest       = fmt.Errorf("TrackingID must have a non-empty Digest part")
	errTrackingIDigestLength       = fmt.Errorf("TrackingID Digest must be exactly 64 hex characters (32 bytes)")
	errTrackidMustHaveProtocolType = fmt.Errorf("TrackingID must have a non-empty ProtocolType part")
	errTrackidStringEmpty          = fmt.Errorf("TrackingID string cannot be empty")
	errTrackidInvalidFormat        = fmt.Errorf("invalid TrackingID format, expected 'ProtocolType-Digest-PartiesState-AuxiliaryData'")
	errUnknownProtocolType         = fmt.Errorf("unknown protocol type in TrackingID")
)

// FromString parses a string representation of a TrackingID into the
// TrackingID struct. The string should be in the format
// "ProtocolType-Digest-PartiesState-AuxiliaryData", where protocolType is an integer, and the rest are
// hexadecimal representation of the respective byte slice.
//
// The tracking ID should always have at least four 'dashes' in the string,
// even if the PartiesState or AuxiliaryData are nil.
// Furthermore, an Empty digest or ProtocolType is not allowed.
// Expects Digest, ProtocolType, PartiesState, and AuxiliaryData to be in hexadecimal
// format and have at most 32 bytes worth of data each.
//
// example: "1-a1b2c3-d4e5f6-1f", "0-a1b2c3-d4e5f6-", "2-a1b2c3--1f", 0-a1b2c3--
func (t *TrackingID) FromString(s string) error {
	if t == nil {
		return errNilTrackID
	}

	if s == nilTrackID {
		return errNilTrackID
	}

	if len(s) == 0 {
		return errTrackidStringEmpty
	}

	// Split the string into parts
	parts := strings.Split(s, "-")
	if len(parts) != 4 {
		return errTrackidInvalidFormat
	}

	if len(parts[0]) == 0 {
		return errTrackidMustHaveProtocolType
	}

	if len(parts[0]) > 2 {
		return errTrackidPartTooLong
	}

	protocolInt, err := strconv.Atoi(parts[0]) // just to check if it's a valid integer string
	if err != nil {
		return fmt.Errorf("failed to parse TrackingID from string: %w", err)
	}

	if !isValidProtocolType(protocolInt) {
		return errUnknownProtocolType
	}

	t.Protocol = uint32(protocolInt)

	if len(parts[1]) == 0 { // must be exactly 32 bytes in hex
		return errTrackidMustHaveDigest
	}

	if len(parts[1]) != 64 {
		return errTrackingIDigestLength
	}

	t.Digest = nil
	t.PartiesState = nil
	t.AuxiliaryData = nil

	byteParts := make([][]byte, 3)
	for i, hexstring := range parts[1:] {
		if len(hexstring) > 64 {
			return errTrackidPartTooLong
		}

		if len(hexstring) == 0 {
			byteParts = append(byteParts, nil)

			continue
		}

		tmp, err := hex.DecodeString(hexstring)
		if err != nil {
			return fmt.Errorf("failed to parse TrackingID from string: %w", err)
		}

		byteParts[i] = tmp
	}

	t.Digest = byteParts[0]
	t.PartiesState = byteParts[1]
	t.AuxiliaryData = byteParts[2]

	return nil
}

func pad32(b []byte) [32]byte {
	padded := [32]byte{}
	copy(padded[:], b)
	return padded
}

func (t *TrackingID) Equals(other *TrackingID) bool {
	if t == nil && other == nil {
		return true
	}

	if t == nil || other == nil {
		return false
	}

	return t.Protocol == other.Protocol &&
		pad32(t.Digest) == pad32(other.Digest) &&
		pad32(t.PartiesState) == pad32(other.PartiesState) &&
		pad32(t.AuxiliaryData) == pad32(other.AuxiliaryData)
}

/*
Returns the ProtocolType enum corresponding to the Protocol field in the TrackingID.
If the TrackingID is nil, returns an error.
If the Protocol field does not correspond to a valid ProtocolType, returns an error.
*/
func (t *TrackingID) GetProtocolType() (ProtocolType, error) {
	if t == nil {
		return "", errNilTrackID
	}

	if !isValidProtocolType(int(t.Protocol)) {
		return "", errUnknownProtocolType
	}

	switch int(t.Protocol) {
	case protocolTypeFROSTSign:
		return ProtocolFROSTSign, nil
	case protocolTypeFROSTDKG:
		return ProtocolFROSTDKG, nil
	case protocolTypeECDSASign:
		return ProtocolECDSASign, nil
	case protocolTypeECDSADKG:
		return ProtocolECDSADKG, nil
	default:
		return "", errUnknownProtocolType
	}
}
