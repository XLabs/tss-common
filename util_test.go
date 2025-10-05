package common

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestTrackidToStr(t *testing.T) {
	testCases := []struct {
		input    *TrackingID
		expected string
	}{
		{
			input:    nil,
			expected: "nilTrackID",
		},
		{
			input:    &TrackingID{Digest: []byte{1, 2, 3}, Protocol: uint32(ProtocolFROSTSign.ToInt()), PartiesState: []byte{4, 5, 6}, AuxiliaryData: []byte{7, 8, 9}},
			expected: "1-0102030000000000000000000000000000000000000000000000000000000000-040506-070809",
		},
		{
			input:    &TrackingID{Digest: []byte{10, 11, 12}, Protocol: uint32(ProtocolFROSTDKG.ToInt()), PartiesState: []byte{13, 14, 15}, AuxiliaryData: []byte{16, 17, 18}},
			expected: "2-0a0b0c0000000000000000000000000000000000000000000000000000000000-0d0e0f-101112",
		},
		{
			input:    &TrackingID{Digest: []byte{1, 2, 3}, Protocol: uint32(ProtocolFROSTSign.ToInt()), PartiesState: nil, AuxiliaryData: nil},
			expected: "1-0102030000000000000000000000000000000000000000000000000000000000--",
		},
	}

	for _, tc := range testCases {
		result := tc.input.ToString()
		if result != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, result)
		}
	}
}

func TestTrackingIdFromString(t *testing.T) {

	t.Run("Valid TrackingID", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected *TrackingID
		}{
			{
				input:    "1-0102030000000000000000000000000000000000000000000000000000000000-040506-070809",
				expected: &TrackingID{Digest: []byte{1, 2, 3}, Protocol: uint32(ProtocolFROSTSign.ToInt()), PartiesState: []byte{4, 5, 6}, AuxiliaryData: []byte{7, 8, 9}},
			},
			{
				input:    "1-0a0b0c0000000000000000000000000000000000000000000000000000000000-0d0e0f-101112",
				expected: &TrackingID{Digest: []byte{10, 11, 12}, Protocol: uint32(ProtocolFROSTSign.ToInt()), PartiesState: []byte{13, 14, 15}, AuxiliaryData: []byte{16, 17, 18}},
			},
			{
				input:    "1-0102030000000000000000000000000000000000000000000000000000000000--",
				expected: &TrackingID{Digest: []byte{1, 2, 3}, Protocol: uint32(ProtocolFROSTSign.ToInt()), PartiesState: nil, AuxiliaryData: nil},
			},
		}

		for _, tc := range testCases {
			result := &TrackingID{}
			err := result.FromString(tc.input)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
				continue
			}

			if !tc.expected.Equals(result) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		}
	})

	t.Run("Invalid TrackingID", func(t *testing.T) {
		// helpers to compose long/edge case strings without cluttering the table
		b64 := strings.Repeat("b", 64) // valid 64 hex chars
		z64 := strings.Repeat("z", 64) // invalid hex (still 64 chars)

		tests := []struct {
			input string
			valid bool
			err   error // if non-nil, we assert errors.Is(err, this)
		}{
			// --- structural / parts count issues ---
			{"", false, errTrackidStringEmpty},                // empty string
			{"nilTrackID", false, errNilTrackID},              // nil tracking id sentinel
			{"abcdsa", false, errTrackidInvalidFormat},        // no dashes
			{"----", false, errTrackidInvalidFormat},          // too many dashes
			{"1-deadbeef-ff", false, errTrackidInvalidFormat}, // 3 parts instead of 4

			// --- protocol issues ---
			{"---", false, errTrackidMustHaveProtocolType},                       // all empty parts
			{"-abcd--", false, errTrackidMustHaveProtocolType},                   // empty protocol
			{fmt.Sprintf("-%s-ff-", b64), false, errTrackidMustHaveProtocolType}, // empty protocol with valid 64-hex digest
			{fmt.Sprintf("123-%s-ff-", b64), false, errTrackidPartTooLong},       // protocol too long (>2 chars)
			{"0-efff--", false, errUnknownProtocolType},                          // unknown protocol
			{"9-efff--", false, errUnknownProtocolType},                          // unknown protocol
			{"12140-efff--", false, errTrackidPartTooLong},
			{"ab-efff--", false, nil}, // non-integer protocol (no sentinel; just expect non-nil error)

			// --- digest problems ---
			{"1--abcd-abcd", false, errTrackidMustHaveDigest}, // empty digest
			{"1-0a-ff-", false, errTrackingIDigestLength},     // digest too short
			{"1-0102030000000000000000000000000000000000000000000000000000000000-1111111111111111111111111111111111111111111111111111111111111111111-", false, errTrackidPartTooLong},     // 3rd too long
			{"1-0102030000000000000000000000000000000000000000000000000000000000-def0-1111111111111111111111111111111111111111111111111111111111111111111", false, errTrackidPartTooLong}, // 4th too long
			{fmt.Sprintf("1-%s-ff-", z64), false, nil}, // non-hex digest (no sentinel; just expect non-nil error)

			// --- valid boundary/within-limit sanity checks (keep here so one place asserts behavior) ---
			{"1-0102030000000000000000000000000000000000000000000000000000000000--", true, nil},   // within limit
			{"2-0102030000000000000000000000000000000000000000000000000000000000-ff-", true, nil}, // within limit
			{"1-0102030000000000000000000000000000000000000000000000000000000000--ff", true, nil}, // within limit
		}

		trackid := &TrackingID{}
		for _, tt := range tests {
			err := trackid.FromString(tt.input)
			if (err == nil) != tt.valid {
				t.Errorf("FromString(%q) = %v; want valid=%v", tt.input, err, tt.valid)
				continue
			}
			// if an error sentinel is specified, assert it matches (or is wrapped)
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Errorf("FromString(%q) = %v; want errors.Is(...,%v)", tt.input, err, tt.err)
			}
		}
	})
}

func TestFromString_NilReceiver(t *testing.T) {
	var tid *TrackingID = nil
	err := tid.FromString("some-string")
	if !errors.Is(err, errNilTrackID) {
		t.Fatalf("FromString(nil receiver) err=%v, want errNilTrackID", err)
	}
}

func TestFromString_BadHex(t *testing.T) {
	// odd-length hex for digest -> hex.ErrLength inside wrapped error
	oddHex := "1-0102030000000000000000000000000000000000000000000000000000000000--ff0"
	var tid TrackingID

	if tid.FromString(oddHex) == nil {
		t.Fatalf("expected odd-length hex parse error, got nil")
	}

	badHex := "1-01020300000000000000000000000000000000000000000000000000000000000000zz--ffg0"
	if tid.FromString(badHex) == nil {
		t.Fatalf("expected bad hex parse error, got nil")
	}
}

func TestBoolByteRoundTrip(t *testing.T) {
	cases := [][]bool{
		{},
		{true},
		{false},
		{true, false, true, false, true, false, true, false},                                              // exactly 8
		{true, true, true, true, true, true, true, true, true},                                            // 9 (spills to second byte)
		{false, true, false, true, false, true, false, true, true},                                        // mixed 9
		{true, false, false, true, true, false, false, false, true, true, false, true, false, true, true}, // 15
	}

	for i, in := range cases {
		gotBytes := ConvertBoolArrayToByteArray(in)
		gotBools := ConvertByteArrayToBoolArray(gotBytes, len(in))
		if len(gotBools) != len(in) {
			t.Fatalf("case %d: length mismatch: got %d want %d", i, len(gotBools), len(in))
		}
		for j := range in {
			if gotBools[j] != in[j] {
				t.Fatalf("case %d: index %d mismatch: got %v want %v", i, j, gotBools[j], in[j])
			}
		}
	}
}

func TestConvertBoolArrayToByteArray_BitPacking(t *testing.T) {
	// bits: 1 0 1 1 0 0 1 0  |  1  (i.e., 9 bits)
	in := []bool{true, false, true, true, false, false, true, false, true}
	got := ConvertBoolArrayToByteArray(in)

	if len(got) != 2 {
		t.Fatalf("expected 2 bytes, got %d", len(got))
	}
	// little-endian bit order within a byte as per code: (i%8)th bit
	// first byte should be: b0..b7 => 1,0,1,1,0,0,1,0 = 0b01001101 = 0x4D
	if got[0] != 0x4D {
		t.Fatalf("first byte expected 0x4D got 0x%02X", got[0])
	}
	// second byte should have only bit 0 set => 0x01
	if got[1] != 0x01 {
		t.Fatalf("second byte expected 0x01 got 0x%02X", got[1])
	}
}

func TestTrackingID_BitLenAndPartyStateOk(t *testing.T) {
	bools := []bool{
		true, false, true, true, false, false, true, false, // 0..7
		true, // 8
	}
	ps := ConvertBoolArrayToByteArray(bools)
	tid := &TrackingID{PartiesState: ps}

	boolsLensInBytes := (len(bools) + 7) / 8
	if got := tid.BitLen() / 8; got != boolsLensInBytes { // rounds up to full byte
		t.Fatalf("BitLen() = %d, want %d", got, boolsLensInBytes)
	}

	if tid.BitLen() != len(ps)*8 {
		t.Fatalf("BitLen expected %d got %d", len(ps)*8, tid.BitLen())
	}

	// Check that PartyStateOk mirrors the original bools
	for i, b := range bools {
		if tid.PartyStateOk(i) != b {
			t.Fatalf("PartyStateOk(%d) = %v, want %v", i, tid.PartyStateOk(i), b)
		}
	}
}

func TestTrackingID_ToStringAndToByteString(t *testing.T) {
	tid := &TrackingID{
		Protocol:      1,
		Digest:        []byte{0xDE, 0xAD, 0xBE, 0xEF},
		PartiesState:  []byte{0x01, 0x02},
		AuxiliaryData: []byte{0xCA, 0xFE},
	}
	want := tid.ToString()

	if got := tid.ToByteString(); !bytes.Equal(got, []byte(want)) {
		t.Fatalf("ToByteString mismatch, got %q", string(got))
	}
}

func TestTrackingID_Equals(t *testing.T) {
	base := &TrackingID{
		Protocol:      2,
		Digest:        []byte{1, 2, 3, 4},
		PartiesState:  []byte{5, 6, 7, 8},
		AuxiliaryData: []byte{9, 10},
	}
	// Same values (different backing slices)
	otherSame := &TrackingID{
		Protocol:      2,
		Digest:        []byte{1, 2, 3, 4},
		PartiesState:  []byte{5, 6, 7, 8},
		AuxiliaryData: []byte{9, 10},
	}
	if !base.Equals(otherSame) {
		t.Fatalf("Equals should be true for identical content")
	}

	// Different protocol
	diffProt := &TrackingID{Protocol: 3, Digest: base.Digest, PartiesState: base.PartiesState, AuxiliaryData: base.AuxiliaryData}
	if base.Equals(diffProt) {
		t.Fatalf("Equals should be false for different Protocol")
	}

	// Different digest (but same length)
	diffDigest := &TrackingID{Protocol: base.Protocol, Digest: []byte{1, 2, 3, 9}, PartiesState: base.PartiesState, AuxiliaryData: base.AuxiliaryData}
	if base.Equals(diffDigest) {
		t.Fatalf("Equals should be false for different Digest")
	}

	// Different lengths: Equals pads to 32 bytes; ensure mismatch still detected
	diffLen := &TrackingID{Protocol: base.Protocol, Digest: []byte{1, 2, 3}, PartiesState: base.PartiesState, AuxiliaryData: base.AuxiliaryData}
	if base.Equals(diffLen) {
		t.Fatalf("Equals should be false for different Digest content/length")
	}

	// Nil comparisons
	if !(*TrackingID)(nil).Equals((*TrackingID)(nil)) {
		t.Fatalf("nil.Equals(nil) should be true")
	}
	if (*TrackingID)(nil).Equals(base) {
		t.Fatalf("nil.Equals(non-nil) should be false")
	}
	if base.Equals((*TrackingID)(nil)) {
		t.Fatalf("non-nil.Equals(nil) should be false")
	}
}

// Note: Positive FromString round-trip depends on the package's isValidProtocolType.
func TestTrackingID_FromString_RoundTripIfProtocolAllowed(t *testing.T) {
	var tid TrackingID

	digest := bytes.Repeat([]byte{0x01}, 32) // 32 bytes
	parties := []byte{0x10, 0x20, 0x30}
	aux := []byte{0xAA, 0xBB}

	s := fmt.Sprintf("1-%x-%x-%x", digest, parties, aux)
	err := tid.FromString(s)
	if err != nil {
		t.Fatalf("unexpected error parsing valid-looking string: %v", err)
	}

	if tid.Protocol != 1 {
		t.Fatalf("Protocol got %d want 1", tid.Protocol)
	}
	if !bytes.Equal(tid.Digest, digest) {
		t.Fatalf("Digest mismatch")
	}
	if !bytes.Equal(tid.PartiesState, parties) {
		t.Fatalf("PartiesState mismatch")
	}
	if !bytes.Equal(tid.AuxiliaryData, aux) {
		t.Fatalf("AuxiliaryData mismatch")
	}

	// Round-trip format
	if tid.ToString() != s {
		t.Fatalf("round-trip ToString mismatch:\n got: %q\nwant: %q", tid.ToString(), s)
	}
}

func TestTrackingID_FromString_ValidBoundaryLengths(t *testing.T) {
	var tid TrackingID

	// Exactly 64 hex chars (32 bytes) for digest, parties, and aux.
	digest := bytes.Repeat([]byte{'a'}, 64) // "aaaaaaaa..." (64)
	parties := bytes.Repeat([]byte{'b'}, 64)
	aux := bytes.Repeat([]byte{'c'}, 64)

	s := fmt.Sprintf("1-%s-%s-%s", digest, parties, aux)
	if err := tid.FromString(s); err != nil {
		t.Fatalf("unexpected error for valid boundary lengths: %v", err)
	}

	if tid.Protocol != 1 {
		t.Fatalf("Protocol: got %d want 0", tid.Protocol)
	}
	if !bytes.Equal(tid.Digest, bytes.Repeat([]byte{0xaa}, 32)) {
		t.Fatalf("Digest bytes mismatch: got %x", tid.Digest)
	}
	if !bytes.Equal(tid.PartiesState, bytes.Repeat([]byte{0xbb}, 32)) {
		t.Fatalf("PartiesState bytes mismatch: got %x", tid.PartiesState)
	}
	if !bytes.Equal(tid.AuxiliaryData, bytes.Repeat([]byte{0xcc}, 32)) {
		t.Fatalf("AuxiliaryData bytes mismatch: got %x", tid.AuxiliaryData)
	}
}

func TestTrackingID_FromString_EmptyPartiesAndAuxAllowed(t *testing.T) {
	var tid TrackingID

	// Empty parties and aux should be accepted (digest still 64 hex chars).
	digest := bytes.Repeat([]byte{'f'}, 64)
	s := fmt.Sprintf("1-%s--", digest)

	err := tid.FromString(s)
	if err != nil {
		t.Fatalf("unexpected error with empty parties/aux: %v", err)
	}
	if len(tid.PartiesState) != 0 || len(tid.AuxiliaryData) != 0 {
		t.Fatalf("expected empty PartiesState and AuxiliaryData")
	}
}

func TestTrackingID_FromString_OddLengthHexInDigest(t *testing.T) {
	var tid TrackingID

	// 63 hex chars (odd / not 64) -> errTrackingIDigestLength
	digest63 := bytes.Repeat([]byte{'a'}, 63)
	s := fmt.Sprintf("1-%s--", digest63)
	if err := tid.FromString(s); err == nil || err != errTrackingIDigestLength {
		t.Fatalf("expected errTrackingIDigestLength for 63-char digest, got %v", err)
	}

	// 65 hex chars (also not 64) -> errTrackingIDigestLength
	digest65 := bytes.Repeat([]byte{'a'}, 65)
	s = fmt.Sprintf("1-%s--", digest65)
	if err := tid.FromString(s); err == nil || err != errTrackingIDigestLength {
		t.Fatalf("expected errTrackingIDigestLength for 65-char digest, got %v", err)
	}
}

func TestTrackingID_FromString_OddLengthHexInPartiesOrAux(t *testing.T) {
	var tid TrackingID

	// Valid digest, parties odd-length -> expect a decode error (non-nil)
	validDigest := bytes.Repeat([]byte{'d'}, 64)
	oddParties := bytes.Repeat([]byte{'e'}, 63) // odd
	s := fmt.Sprintf("2-%s-%s-", validDigest, oddParties)

	err := tid.FromString(s)
	if err == nil {
		t.Fatalf("expected error for odd-length hex in parties, got nil")
	}

	// Valid digest, aux odd-length -> expect a decode error (non-nil)
	oddAux := bytes.Repeat([]byte{'f'}, 63)
	s = fmt.Sprintf("2-%s--%s", validDigest, oddAux)

	err = tid.FromString(s)
	if err == nil {
		t.Fatalf("expected error for odd-length hex in aux, got nil")
	}
}

func TestTrackingID_FromString_NonHexInPartiesOrAux(t *testing.T) {
	var tid TrackingID
	validDigest := bytes.Repeat([]byte{'a'}, 64)

	// 'g' is not a hex digit
	s := fmt.Sprintf("1-%s-%s-", validDigest, "gg")
	if err := tid.FromString(s); err == nil {
		t.Fatalf("expected error for non-hex in parties, got nil")
	}

	s = fmt.Sprintf("1-%s--%s", validDigest, "zz")
	if err := tid.FromString(s); err == nil {
		t.Fatalf("expected error for non-hex in aux, got nil")
	}
}

func TestTrackingID_FromString_HexCaseInsensitivity(t *testing.T) {
	var tid TrackingID

	// Mixed-case hex should be accepted and decoded identically
	digest := "AaBbCcDdEeFf" + "00" + "11" + "22" + "33" + "44" + "55" + "66" + "77" + "88" + "99" + "aA" + "Bb" + "Cc" + "Dd" + "Ee"
	// Ensure digest totals 64 hex chars
	for len(digest) < 64 {
		digest += "0A"
	}
	digest = digest[:64]

	parties := "FfEeDdCcBbAa" // 12 (6 bytes)
	aux := "abcdef"           // 6 (3 bytes)

	s := fmt.Sprintf("1-%s-%s-%s", digest, parties, aux)
	if err := tid.FromString(s); err != nil {
		t.Fatalf("unexpected error parsing mixed-case hex: %v", err)
	}

	// Re-serialize and parse again to confirm stable round-trip formatting
	rt := tid.ToString()
	var tid2 TrackingID
	if err := tid2.FromString(rt); err != nil {
		t.Fatalf("unexpected error re-parsing ToString output: %v", err)
	}
	if !tid.Equals(&tid2) {
		t.Fatalf("round-trip equals mismatch")
	}
}
