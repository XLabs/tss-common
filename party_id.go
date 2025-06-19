// Copyright © 2019 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package common

import (
	"sort"
)

type (
	UnSortedPartyIDs []*PartyID
	// SortedPartyIDs is a slice of PartyID sorted by the string order of ther IDs.
	SortedPartyIDs []*PartyID
)

func (pid *PartyID) ValidateBasic() bool {
	return pid != nil && pid.ID != ""
}

func (p *PartyID) Equals(other *PartyID) bool {
	if p == nil && other == nil {
		return true
	}

	if p == nil || other == nil {
		return false
	}

	return p.GetID() == other.GetID()
}

func (p *PartyID) ToString() string {
	if p == nil {
		return ""
	}

	return p.GetID()
}

// SortPartyIDs sorts a list of []*PartyID by their keys in ascending order
func SortPartyIDs(ids UnSortedPartyIDs) SortedPartyIDs {
	sorted := make(SortedPartyIDs, 0, len(ids))
	for _, id := range ids {
		sorted = append(sorted, id)
	}
	sort.Sort(sorted)
	// assign party indexes

	return sorted
}

func (committee UnSortedPartyIDs) IsInCommittee(self *PartyID) bool {
	return committee.IndexInCommittee(self) != -1
}

func (committee UnSortedPartyIDs) IndexInCommittee(self *PartyID) int {
	for i, v := range committee {
		if self.Equals(v) {
			return i
		}
	}

	return -1
}

// Sortable

func (spids SortedPartyIDs) Len() int {
	return len(spids)
}

func (spids SortedPartyIDs) Less(a, b int) bool {
	return spids[a].GetID() < spids[b].GetID()
}

func (spids SortedPartyIDs) Swap(a, b int) {
	spids[a], spids[b] = spids[b], spids[a]
}
