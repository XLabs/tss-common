package common

import "errors"

// ProtocolType represents the type of cryptographic protocol in use.
type ProtocolType string

const (
	// String protocol identifiers
	ProtocolFROSTSign ProtocolType = "FROST:SIGN"
	ProtocolFROSTDKG  ProtocolType = "FROST:DKG"
	ProtocolECDSASign ProtocolType = "ECDSA:SIGN"
	ProtocolECDSADKG  ProtocolType = "ECDSA:DKG"
)

// Integer protocol identifiers (useful for internal indexing, enums, etc.)
const (
	protocolTypeMin = iota
	protocolTypeFROSTSign
	protocolTypeFROSTDKG
	protocolTypeECDSASign
	protocolTypeECDSADKG
	protocolTypeMax
)

func (p ProtocolType) ToString() string {
	return string(p)
}

func (p ProtocolType) ToInt() int {
	switch p {
	case ProtocolFROSTSign:
		return protocolTypeFROSTSign
	case ProtocolFROSTDKG:
		return protocolTypeFROSTDKG
	case ProtocolECDSASign:
		return protocolTypeECDSASign
	case ProtocolECDSADKG:
		return protocolTypeECDSADKG
	default:
		return -1
	}
}

var ErrUnknownProtocolType = errors.New("unknown protocol type")

func ProtocolTypeFromString(s string) (ProtocolType, error) {
	switch s {
	case string(ProtocolFROSTSign):
		return ProtocolFROSTSign, nil
	case string(ProtocolFROSTDKG):
		return ProtocolFROSTDKG, nil
	case string(ProtocolECDSASign):
		return ProtocolECDSASign, nil
	case string(ProtocolECDSADKG):
		return ProtocolECDSADKG, nil
	default:
		return "", ErrUnknownProtocolType
	}
}

func ProtocolTypeFromInt(pInt int) (ProtocolType, error) {
	switch pInt {
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

func isValidProtocolType(n int) bool {
	return n > protocolTypeMin && n < protocolTypeMax
}
