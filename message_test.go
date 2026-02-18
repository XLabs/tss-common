package common_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	common "github.com/xlabs/tss-common"
	"github.com/xlabs/tss-common/service/signer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// mockMessageContent is a mock implementation of MessageContent for testing.
type mockMessageContent struct {
	// Embed SignRequest to satisfy the proto.Message interface,
	// which is required for anypb.New() to work.
	signer.SignRequest
}

//	MessageContent interface {
//			proto.Message
//			ValidateBasic() bool
//			RoundNumber() int
//			GetProtocol() ProtocolType
//		}
func (m *mockMessageContent) RoundNumber() int                 { return 0 }
func (m *mockMessageContent) GetProtocol() common.ProtocolType { return common.ProtocolFROSTSign }
func (m *mockMessageContent) ValidateBasic() bool              { return true }

func TestNewMessageWrapper(t *testing.T) {
	from := &common.PartyID{ID: "from-party"}
	to := &common.PartyID{ID: "to-party"}
	// Use the mock's control fields for initialization
	content := &mockMessageContent{}
	trackingID := &common.TrackingID{Protocol: 1}

	t.Run("broadcast message", func(t *testing.T) {
		routing := common.MessageRouting{From: from, To: nil}
		wrapper := common.NewMessageWrapper(routing, content)

		assert.NotNil(t, wrapper)
		assert.Equal(t, from, wrapper.From)
		assert.Nil(t, wrapper.To)
		assert.False(t, wrapper.IsToOldCommittee)
		assert.False(t, wrapper.IsToOldAndNewCommittees)
		// This assertion now works because NewMessageWrapper calls our
		// mock GetProtocol() method.
		assert.Equal(t, string(common.ProtocolFROSTSign), wrapper.Protocol)
		assert.Nil(t, wrapper.GetTrackingID())

		anyMsg, err := anypb.New(content)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(anyMsg, wrapper.Message))
	})

	t.Run("direct message", func(t *testing.T) {
		routing := common.MessageRouting{From: from, To: to}
		wrapper := common.NewMessageWrapper(routing, content)

		assert.NotNil(t, wrapper)
		assert.Equal(t, from, wrapper.From)
		assert.NotNil(t, wrapper.To)
		assert.Equal(t, to.ID, wrapper.To.ID)
	})

	t.Run("with tracking ID", func(t *testing.T) {
		routing := common.MessageRouting{From: from}
		wrapper := common.NewMessageWrapper(routing, content, trackingID)

		assert.NotNil(t, wrapper)
		assert.Equal(t, trackingID, wrapper.GetTrackingID())
	})

	t.Run("resharing flags", func(t *testing.T) {
		routing := common.MessageRouting{
			From:                    from,
			IsToOldCommittee:        true,
			IsToOldAndNewCommittees: true,
		}
		wrapper := common.NewMessageWrapper(routing, content)

		assert.True(t, wrapper.IsToOldCommittee)
		assert.True(t, wrapper.IsToOldAndNewCommittees)
	})
}

func TestMessageImpl(t *testing.T) {
	from := &common.PartyID{ID: "from-party"}
	to := &common.PartyID{ID: "to-party"}
	// Use the mock's control fields for initialization
	content := &mockMessageContent{}

	// Direct message
	directRouting := common.MessageRouting{From: from, To: to}
	directWire := common.NewMessageWrapper(directRouting, content)
	directMsg := common.NewMessage(directRouting, content, directWire)

	// Broadcast message
	broadcastRouting := common.MessageRouting{From: from, To: nil}
	broadcastWire := common.NewMessageWrapper(broadcastRouting, content)
	broadcastMsg := common.NewMessage(broadcastRouting, content, broadcastWire)

	t.Run("GetProtocol", func(t *testing.T) {
		assert.Equal(t, common.ProtocolFROSTSign, directMsg.GetProtocol())
	})

	t.Run("Type", func(t *testing.T) {
		m := mockMessageContent{}

		assert.Equal(t, string(m.ProtoReflect().Type().Descriptor().FullName()), directMsg.Type())
	})

	t.Run("GetTo", func(t *testing.T) {
		assert.Equal(t, to, directMsg.GetTo())
		assert.Nil(t, broadcastMsg.GetTo())
	})

	t.Run("GetFrom", func(t *testing.T) {
		assert.Equal(t, from, directMsg.GetFrom())
	})

	t.Run("IsBroadcast", func(t *testing.T) {
		assert.False(t, directMsg.IsBroadcast())
		assert.True(t, broadcastMsg.IsBroadcast())

		// Test edge case of To being non-nil but with an empty ID
		routing := common.MessageRouting{From: from, To: &common.PartyID{ID: ""}}
		wire := common.NewMessageWrapper(routing, content)
		msg := common.NewMessage(routing, content, wire)
		assert.True(t, msg.IsBroadcast())
	})

	t.Run("Resharing flags", func(t *testing.T) {
		routing := common.MessageRouting{
			From:                    from,
			IsToOldCommittee:        true,
			IsToOldAndNewCommittees: true,
		}
		wire := common.NewMessageWrapper(routing, content)
		msg := common.NewMessage(routing, content, wire)

		assert.True(t, msg.IsToOldCommittee())
		assert.True(t, msg.IsToOldAndNewCommittees())
		assert.False(t, directMsg.IsToOldCommittee())
		assert.False(t, directMsg.IsToOldAndNewCommittees())
	})

	t.Run("WireBytes", func(t *testing.T) {
		bz, routing, err := directMsg.WireBytes()
		assert.NoError(t, err)
		assert.NotNil(t, bz)
		assert.Equal(t, &directRouting, routing)

		// Check that From and To are nilled on the wire to save space
		unmarshalledWire := &common.MessageWrapper{}
		err = proto.Unmarshal(bz, unmarshalledWire)
		assert.NoError(t, err)
		assert.Nil(t, unmarshalledWire.From)
		assert.Nil(t, unmarshalledWire.To)
		assert.Equal(t, directWire.Protocol, unmarshalledWire.Protocol)
	})

	t.Run("WireMsg", func(t *testing.T) {
		assert.Equal(t, directWire, directMsg.WireMsg())
	})

	t.Run("Content", func(t *testing.T) {
		assert.Equal(t, content, directMsg.Content())
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		// This now calls our mock ValidateBasic() and returns true
		assert.True(t, directMsg.ValidateBasic())

		// Use the mock's control fields for initialization
		invalidContent := &mockMessageContent{}
		validMock := common.NewMessage(directRouting, invalidContent, nil)
		assert.True(t, validMock.ValidateBasic())
	})

	t.Run("String", func(t *testing.T) {
		expectedDirect := fmt.Sprintf("Type: %s, From: %s, To: %s", directMsg.Type(), from.String(), to.String())
		assert.Equal(t, expectedDirect, directMsg.String())

		expectedBroadcast := fmt.Sprintf("Type: %s, From: %s, To: all", broadcastMsg.Type(), from.String())
		assert.Equal(t, expectedBroadcast, broadcastMsg.String())

		// With old committee flag
		routing := common.MessageRouting{From: from, IsToOldCommittee: true}
		wire := common.NewMessageWrapper(routing, content)
		msg := common.NewMessage(routing, content, wire)
		expectedWithFlag := fmt.Sprintf("Type: %s, From: %s, To: all (To Old Committee)", msg.Type(), from.String())
		assert.Equal(t, expectedWithFlag, msg.String())
	})
}

func TestMessageRouting_IsBroadcast(t *testing.T) {
	from := &common.PartyID{ID: "from"}
	to := &common.PartyID{ID: "to"}

	tests := []struct {
		name     string
		routing  *common.MessageRouting
		expected bool
	}{
		{
			name:     "broadcast with nil To",
			routing:  &common.MessageRouting{From: from, To: nil},
			expected: true,
		},
		{
			name:     "broadcast with empty To ID",
			routing:  &common.MessageRouting{From: from, To: &common.PartyID{ID: ""}},
			expected: true,
		},
		{
			name:     "direct message",
			routing:  &common.MessageRouting{From: from, To: to},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.routing.IsBroadcast())
		})
	}
}
