package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/xlabs/tss-common"
)

func TestPartyID_ValidateBasic(t *testing.T) {
	tests := []struct {
		name     string
		pid      *PartyID
		expected bool
	}{
		{"valid party id", &PartyID{ID: "party-1"}, true},
		{"nil party id", nil, false},
		{"empty id", &PartyID{ID: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pid.ValidateBasic())
		})
	}
}

func TestPartyID_Equals(t *testing.T) {
	p1 := &PartyID{ID: "party-1"}
	p1_clone := &PartyID{ID: "party-1"}
	p2 := &PartyID{ID: "party-2"}

	tests := []struct {
		name     string
		p1       *PartyID
		p2       *PartyID
		expected bool
	}{
		{"two equal parties", p1, p1_clone, true},
		{"two different parties", p1, p2, false},
		{"one party is nil (left)", nil, p1, false},
		{"one party is nil (right)", p1, nil, false},
		{"both parties are nil", nil, nil, true},
		{"same party instance", p1, p1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.p1.Equals(tt.p2))
		})
	}
}

func TestPartyID_ToString(t *testing.T) {
	tests := []struct {
		name     string
		pid      *PartyID
		expected string
	}{
		{"valid party id", &PartyID{ID: "party-1"}, "party-1"},
		{"nil party id", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pid.ToString())
		})
	}
}

func TestSortPartyIDs(t *testing.T) {
	p1 := &PartyID{ID: "party-c"}
	p2 := &PartyID{ID: "party-a"}
	p3 := &PartyID{ID: "party-b"}

	unsorted := UnSortedPartyIDs{p1, p2, p3}
	expected := SortedPartyIDs{p2, p3, p1}

	sorted := SortPartyIDs(unsorted)

	assert.Equal(t, len(expected), len(sorted))
	for i := range sorted {
		assert.True(t, expected[i].Equals(sorted[i]), "Element at index %d should be equal", i)
	}

	// Test empty
	sortedEmpty := SortPartyIDs(UnSortedPartyIDs{})
	assert.Empty(t, sortedEmpty)

	// Test single
	sortedSingle := SortPartyIDs(UnSortedPartyIDs{p1})
	assert.Equal(t, SortedPartyIDs{p1}, sortedSingle)
}

func TestUnSortedPartyIDs_IsInCommittee(t *testing.T) {
	p1 := &PartyID{ID: "party-1"}
	p2 := &PartyID{ID: "party-2"}
	p3 := &PartyID{ID: "party-3"}
	p4 := &PartyID{ID: "party-4"}

	committee := UnSortedPartyIDs{p1, p2, p3}

	tests := []struct {
		name     string
		party    *PartyID
		expected bool
	}{
		{"party in committee", p2, true},
		{"party not in committee", p4, false},
		{"nil party", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, committee.IsInCommittee(tt.party))
		})
	}
}

func TestUnSortedPartyIDs_IndexInCommittee(t *testing.T) {
	p1 := &PartyID{ID: "party-1"}
	p2 := &PartyID{ID: "party-2"}
	p3 := &PartyID{ID: "party-3"}
	p4 := &PartyID{ID: "party-4"}

	committee := UnSortedPartyIDs{p1, p2, p3}

	tests := []struct {
		name     string
		party    *PartyID
		expected int
	}{
		{"party at index 0", p1, 0},
		{"party at index 1", p2, 1},
		{"party at index 2", p3, 2},
		{"party not in committee", p4, -1},
		{"nil party", nil, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, committee.IndexInCommittee(tt.party))
		})
	}
}
