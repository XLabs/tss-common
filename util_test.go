package common

import (
	"encoding/hex"
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
			input:    &TrackingID{Digest: []byte{1, 2, 3}, Protocol: []byte(ProtocolFROST), PartiesState: []byte{4, 5, 6}, AuxilaryData: []byte{7, 8, 9}},
			expected: "010203-46524f5354-040506-070809",
		},
		{
			input:    &TrackingID{Digest: []byte{10, 11, 12}, Protocol: []byte(ProtocolFROST), PartiesState: []byte{13, 14, 15}, AuxilaryData: []byte{16, 17, 18}},
			expected: "0a0b0c-46524f5354-0d0e0f-101112",
		},
		{
			input:    &TrackingID{Digest: []byte{1, 2, 3}, Protocol: []byte(ProtocolEmpty), PartiesState: nil, AuxilaryData: nil},
			expected: "010203-456d707479--",
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
				input:    "010203-456d707479-040506-070809",
				expected: &TrackingID{Digest: []byte{1, 2, 3}, Protocol: []byte(ProtocolEmpty), PartiesState: []byte{4, 5, 6}, AuxilaryData: []byte{7, 8, 9}},
			},
			{
				input:    "0a0b0c-46524f5354-0d0e0f-101112",
				expected: &TrackingID{Digest: []byte{10, 11, 12}, Protocol: []byte(ProtocolFROST), PartiesState: []byte{13, 14, 15}, AuxilaryData: []byte{16, 17, 18}},
			},
			{
				input:    "010203-46524f5354--",
				expected: &TrackingID{Digest: []byte{1, 2, 3}, Protocol: []byte(ProtocolFROST), PartiesState: nil, AuxilaryData: nil},
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
		tests := []struct {
			input string
			valid bool
			err   error
		}{
			{"1111111111111111111111111111111111111111111111111111111111111111111-321-abcd-defgh", false, errTrackidPartTooLong}, // already 4 parts
			{"1234-1111111111111111111111111111111111111111111111111111111111111111111-342-defgh", false, errTrackidPartTooLong}, // already 4 parts
			{"1234-abcd-1111111111111111111111111111111111111111111111111111111111111111111-", false, errTrackidPartTooLong},     // 4th part empty, 3rd too long
			{"1234-abcd-def0-1111111111111111111111111111111111111111111111111111111111111111111", false, errTrackidPartTooLong}, // all 4 parts present, last too long

			{"abc-abcd-defgh-", false, hex.ErrLength}, // odd-length hex in first part;

			{"", false, errTrackidStringEmpty}, // empty string

			{"nilTrackID", false, errNilTrackID}, // nil tracking ID

			{"----", false, errTrackidInvalidFormat},      // too many dashes
			{"-abcdabcd", false, errTrackidInvalidFormat}, // empty digest

			{"-abcd-abcd-efgh", false, errTrackidMustHaveDigest}, // empty digest parts
			{"---", false, errTrackidMustHaveDigest},             // empty parts
			{"abcd---", false, errTrackidMustHaveProtocolType},   // empty protocolType

			{"abcd-efff--", false, errUnknownProtocolType}, // unknown protocol type

			{"abcd-456d707479--", true, nil},   // within limit
			{"abcd-456d707479-ff-", true, nil}, // within limit
			{"abcd-456d707479--ff", true, nil}, // within limit
		}

		trackid := &TrackingID{}
		for _, tt := range tests {
			err := trackid.FromString(tt.input)
			if (err == nil) != tt.valid {
				t.Errorf("FromString(%q) = %v; want valid: %v", tt.input, err, tt.valid)

				continue
			}

			// check that the error is the expected one or wraps it:
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Errorf("FromString(%q) = %v; want error: %v", tt.input, err, tt.err)
			}

		}

	})

}

func hx(s string) string { return fmt.Sprintf("%x", []byte(s)) }

func TestConvertBoolArrayToByteArray_RoundTrip(t *testing.T) {
	cases := [][]bool{
		{},                               // empty
		{false},                          // single false
		{true},                           // single true
		{true, false, true, false, true}, // odd length
		{true, true, true, true, true, true, true, true},            // exactly one byte
		{true, false, false, true, false, true, false, false, true}, // cross byte
	}
	for i, in := range cases {
		got := ConvertByteArrayToBoolArray(ConvertBoolArrayToByteArray(in), len(in))
		if len(got) != len(in) {
			t.Fatalf("case %d: length mismatch: got %d want %d", i, len(got), len(in))
		}
		if len(got) != len(in) {
			t.Fatalf("case %d: length mismatch: got %d want %d", i, len(got), len(in))
		}

		for j := range in {
			if got[j] != in[j] {
				t.Fatalf("case %d: idx %d mismatch: got %v want %v", i, j, got[j], in[j])
			}
		}
	}
}

func TestPartyStateOkAndBitLen(t *testing.T) {
	// 10 bits, with ones at positions 0, 3, 8, 9
	bools := []bool{true, false, false, true, false, false, false, false, true, true}
	tid := &TrackingID{PartiesState: ConvertBoolArrayToByteArray(bools)}

	boolsLensInBytes := (len(bools) + 7) / 8
	if got := tid.BitLen() / 8; got != boolsLensInBytes { // rounds up to full byte
		t.Fatalf("BitLen() = %d, want %d", got, boolsLensInBytes)
	}

	for i, v := range bools {
		if tid.PartyStateOk(i) != v {
			t.Fatalf("PartyStateOk(%d) = %v, want %v", i, tid.PartyStateOk(i), v)
		}
	}
}

func TestFromString_RoundTrip(t *testing.T) {
	orig := &TrackingID{
		Digest:       []byte("digest-123"),
		Protocol:     []byte(ProtocolEmpty.ToString()), // valid protocol sentinel
		PartiesState: []byte{0x00, 0x10, 0xff},
		AuxilaryData: []byte("aux"),
	}
	var parsed TrackingID
	if err := parsed.FromString(orig.ToString()); err != nil {
		t.Fatalf("FromString round-trip failed: %v", err)
	}

	if !parsed.Equals(orig) {
		t.Fatalf("parsed TrackingID != original")
	}
	// change Protocol and check Equals still true
	parsed.Protocol = []byte(ProtocolFROST)
	if parsed.Equals(orig) {
		t.Fatalf("parsed TrackingID != original after changing Protocol")
	}
	// please extend this test
}

func TestFromString_NilReceiver(t *testing.T) {
	var tid *TrackingID = nil
	// Minimal valid string: non-empty digest, valid protocol, empty parties/aux
	s := "01-" + hx(ProtocolEmpty.ToString()) + "--"
	err := tid.FromString(s)
	if !errors.Is(err, errNilTrackID) {
		t.Fatalf("FromString(nil receiver) err=%v, want errNilTrackID", err)
	}
}

// // --- Negative/edge cases for FromString ------------------------------------

// func TestFromString_Errors(t *testing.T) {
// 	cases := []struct {
// 		name string
// 		in   string
// 		want error
// 	}{
// 		{ProtocolEmpty.ToString(), "", errTrackidStringEmpty},
// 		{"wrongParts", "aa-bb-cc", errTrackidInvalidFormat},
// 		{"missingDigest", "-" + hx(ProtocolEmpty.ToString()) + "--", errTrackidMustHaveDigest},
// 		{"missingProtocol", "aa---", errTrackidMustHaveProtocolType},
// 		{"partTooLong", strings.Repeat("a", 65) + "-bb-cc-dd", errTrackidPartTooLong},
// 		{"unknownProtocol", "aa-" + hx("SOME_UNKNOWN_PROTO") + "-cc-dd", errUnknownProtocolType},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var tid TrackingID
// 			err := tid.FromString(tc.in)
// 			if !errors.Is(err, tc.want) {
// 				t.Fatalf("FromString(%q) err=%v, want %v", tc.in, err, tc.want)
// 			}
// 		})
// 	}
// }

func TestFromString_InvalidHex(t *testing.T) {
	// 'zz' is invalid hex
	bad := "zz-" + hx(ProtocolEmpty.ToString()) + "-cc-dd"
	var tid TrackingID
	err := tid.FromString(bad)
	if err == nil || !strings.Contains(err.Error(), "failed to parse TrackingID") {
		t.Fatalf("expected hex parse error, got %v", err)
	}
}

func TestFromString_OddLengthHex(t *testing.T) {
	// odd-length hex for digest -> hex.ErrLength inside wrapped error
	odd := "a-" + hx(ProtocolEmpty.ToString()) + "-cc-dd"
	var tid TrackingID
	err := tid.FromString(odd)
	if err == nil {
		t.Fatalf("expected odd-length hex parse error, got nil")
	}
}
