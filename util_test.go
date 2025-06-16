package common

import (
	"encoding/hex"
	"errors"
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
			input:    &TrackingID{Digest: []byte{1, 2, 3}, PartiesState: []byte{4, 5, 6}, AuxilaryData: []byte{7, 8, 9}},
			expected: "010203-040506-070809",
		},
		{
			input:    &TrackingID{Digest: []byte{10, 11, 12}, PartiesState: []byte{13, 14, 15}, AuxilaryData: []byte{16, 17, 18}},
			expected: "0a0b0c-0d0e0f-101112",
		},
		{
			input:    &TrackingID{Digest: []byte{1, 2, 3}, PartiesState: nil, AuxilaryData: nil},
			expected: "010203--",
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
				input:    "010203-040506-070809",
				expected: &TrackingID{Digest: []byte{1, 2, 3}, PartiesState: []byte{4, 5, 6}, AuxilaryData: []byte{7, 8, 9}},
			},
			{
				input:    "0a0b0c-0d0e0f-101112",
				expected: &TrackingID{Digest: []byte{10, 11, 12}, PartiesState: []byte{13, 14, 15}, AuxilaryData: []byte{16, 17, 18}},
			},
			{
				input:    "010203--",
				expected: &TrackingID{Digest: []byte{1, 2, 3}, PartiesState: nil, AuxilaryData: nil},
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
			{"1111111111111111111111111111111111111111111111111111111111111111111-abcd-defgh", false, errTrackidPartTooLong}, // one part too long
			{"1234-1111111111111111111111111111111111111111111111111111111111111111111-defgh", false, errTrackidPartTooLong}, // one part too long
			{"1234-abcd-1111111111111111111111111111111111111111111111111111111111111111111", false, errTrackidPartTooLong},  // one part too long

			{"abc-abcd-defgh", false, hex.ErrLength}, // one part too long

			{"", false, errTrackidStringEmpty}, // empty string

			{"nilTrackID", false, errNilTrackID}, // nil tracking ID

			{"---", false, errTrackidInvalidFormat},       // too many dashes
			{"-abcdabcd", false, errTrackidInvalidFormat}, // empty digest

			{"-abcd-abcd", false, errTrackidMustHaveDigest}, // empty digest parts
			{"--", false, errTrackidMustHaveDigest},         // empty parts

			{"abcd-efff-", true, nil},   // within limit
			{"abcd-efff-ff", true, nil}, // within limit
			{"abcd--ff", true, nil},     // within limit
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
