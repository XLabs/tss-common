// Copyright © 2019 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package common

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

// Used externally to update a LocalParty with a valid ParsedMessage
// set the `to` field to nil if the message is a broadcast.
// if it was direct communication, set the `to` field to the PartyID of the recipient.
func ParseWireMessage(wireBytes []byte, from, to *PartyID) (ParsedMessage, error) {
	wire := new(MessageWrapper)
	if err := proto.Unmarshal(wireBytes, wire); err != nil {
		return nil, err
	}

	return parseWrappedMessage(wire, from, to)
}

var errParse = errors.New("ParseWireMessage: the message contained unknown content")

func parseWrappedMessage(wire *MessageWrapper, from, to *PartyID) (ParsedMessage, error) {
	m, err := wire.Message.UnmarshalNew()
	if err != nil {
		return nil, err
	}

	meta := MessageRouting{
		From: from,
		To:   to,
	}
	// wire is marshalled without these fields to reduce network bandwidth, but we need them in a parsed message to
	// match the messageRouting struct.
	wire.From = meta.From
	wire.To = meta.To

	content, ok := m.(MessageContent)
	if !ok {
		return nil, errParse
	}

	return NewMessage(meta, content, wire), nil
}
