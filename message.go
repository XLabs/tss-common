// TODO: use To as a single PartyID instead of a slice, and remove the variable of isBroadcast.
// if To is nil -> it's a broadcast. Fix everything accordingly.
package common

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	ProtocolEmpty ProtocolType = "Empty"
	ProtocolECDSA ProtocolType = "ECDSA"
	ProtocolFROST ProtocolType = "FROST"
)

type (
	ProtocolType string

	// Message describes the interface of the TSS Message for all protocols
	Message interface {
		// Type is encoded in the protobuf Any structure
		Type() string
		// The set of parties that this message should be sent to
		GetTo() *PartyID
		// The party that this message is from
		GetFrom() *PartyID
		// Indicates whether the message should be broadcast to other participants
		IsBroadcast() bool
		// Indicates whether the message is to the old committee during re-sharing; used mainly in tests
		IsToOldCommittee() bool
		// Indicates whether the message is to both committees during re-sharing; used mainly in tests
		IsToOldAndNewCommittees() bool
		// Returns the encoded inner message bytes to send over the wire along with metadata about how the message should be delivered
		WireBytes() ([]byte, *MessageRouting, error)
		// Returns the protobuf message wrapper struct
		// Only its inner content should be sent over the wire, not this struct itself
		WireMsg() *MessageWrapper

		String() string

		GetProtocol() ProtocolType
	}

	// ParsedMessage represents a message with inner ProtoBuf message content
	ParsedMessage interface {
		Message
		Content() MessageContent
		ValidateBasic() bool
	}

	// MessageContent represents a ProtoBuf message with validation logic
	MessageContent interface {
		proto.Message
		ValidateBasic() bool
		RoundNumber() int
		GetProtocol() ProtocolType
	}

	// MessageRouting holds the full routing information for the message, consumed by the transport
	MessageRouting struct {
		// which participant this message came from
		From *PartyID
		// when `nil` the message should be broadcast to all parties
		To *PartyID
		// whether the message should be sent to old committee participants rather than the new committee
		IsToOldCommittee bool
		// whether the message should be sent to both old and new committee participants
		IsToOldAndNewCommittees bool
	}

	// Implements ParsedMessage; this is a concrete implementation of what messages produced by a LocalParty look like
	MessageImpl struct {
		MessageRouting
		content  MessageContent
		wire     *MessageWrapper
		protocol ProtocolType
	}
)

// GetProtocol implements Message.
func (mm *MessageImpl) GetProtocol() ProtocolType {
	return mm.protocol
}

var (
	_ Message       = (*MessageImpl)(nil)
	_ ParsedMessage = (*MessageImpl)(nil)
)

// ----- //

// NewMessageWrapper constructs a MessageWrapper from routing metadata and content
// digest is an additional parameter
func NewMessageWrapper(routing MessageRouting, content MessageContent, trackingID ...*TrackingID) *MessageWrapper {
	// marshal the content to the ProtoBuf Any type
	anypbMsg, _ := anypb.New(content)
	// convert given PartyIDs to the wire format
	var to *PartyID = nil // broadcast
	if routing.To != nil {
		// direct message
		to = &PartyID{
			ID: routing.To.ID,
		}
	}

	m := &MessageWrapper{
		IsToOldCommittee:        routing.IsToOldCommittee,
		IsToOldAndNewCommittees: routing.IsToOldAndNewCommittees,
		From:                    routing.From,
		To:                      to,
		Message:                 anypbMsg,
		Protocol:                string(content.GetProtocol()),
		unknownFields:           protoimpl.UnknownFields{},
		sizeCache:               0,
	}

	if len(trackingID) > 0 {
		m.TrackingID = trackingID[0]
	}

	return m
}

// ----- //

func NewMessage(meta MessageRouting, content MessageContent, wire *MessageWrapper) ParsedMessage {
	return &MessageImpl{
		MessageRouting: meta,
		content:        content,
		wire:           wire,
		protocol:       content.GetProtocol(),
	}
}

func (mm *MessageImpl) Type() string {
	return string(proto.MessageName(mm.content))
}

func (mm *MessageImpl) GetTo() *PartyID {
	return mm.To
}

func (mm *MessageImpl) GetFrom() *PartyID {
	return mm.From
}

func (mm *MessageImpl) IsBroadcast() bool {
	return mm.GetTo() == nil || mm.GetTo().ID == ""
}

// only `true` in DGRound2Message (resharing)
func (mm *MessageImpl) IsToOldCommittee() bool {
	return mm.wire.IsToOldCommittee
}

// only `true` in DGRound4Message (resharing)
func (mm *MessageImpl) IsToOldAndNewCommittees() bool {
	return mm.wire.IsToOldAndNewCommittees
}

func (mm *MessageImpl) WireBytes() ([]byte, *MessageRouting, error) {
	tmp := proto.Clone(mm.wire).(*MessageWrapper)

	// reducing space on wire.
	tmp.To = nil
	tmp.From = nil

	bz, err := proto.Marshal(tmp)
	if err != nil {
		return nil, nil, err
	}

	return bz, &mm.MessageRouting, nil
}

func (mm *MessageImpl) WireMsg() *MessageWrapper {
	return mm.wire
}

func (mm *MessageImpl) Content() MessageContent {
	return mm.content
}

func (mm *MessageImpl) ValidateBasic() bool {
	return mm.content.ValidateBasic()
}

func (mm *MessageImpl) String() string {
	toStr := "all"
	if mm.To != nil {
		toStr = fmt.Sprintf("%v", mm.To)
	}
	extraStr := ""
	if mm.IsToOldCommittee() {
		extraStr = " (To Old Committee)"
	}
	return fmt.Sprintf("Type: %s, From: %s, To: %s%s", mm.Type(), mm.From.String(), toStr, extraStr)
}

func (m *MessageRouting) IsBroadcast() bool {
	return m.To == nil || m.To.ID == ""
}
