package wasm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	clienttypes "github.com/line/lbm-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/line/lbm-sdk/x/ibc/core/04-channel/types"
	wasmvmtypes "github.com/line/wasmvm/types"
)

func TestMapToWasmVMIBCPacket(t *testing.T) {
	var myTimestamp uint64 = 1
	specs := map[string]struct {
		src channeltypes.Packet
		exp wasmvmtypes.IBCPacket
	}{
		"with height timeout": {
			src: IBCPacketFixture(),
			exp: wasmvmtypes.IBCPacket{
				Data:     []byte("myData"),
				Src:      wasmvmtypes.IBCEndpoint{PortID: "srcPort", ChannelID: "channel-1"},
				Dest:     wasmvmtypes.IBCEndpoint{PortID: "destPort", ChannelID: "channel-2"},
				Sequence: 1,
				Timeout:  wasmvmtypes.IBCTimeout{Block: &wasmvmtypes.IBCTimeoutBlock{Height: 1, Revision: 2}},
			},
		},
		"with time timeout": {
			src: IBCPacketFixture(func(p *channeltypes.Packet) {
				p.TimeoutTimestamp = myTimestamp
				p.TimeoutHeight = clienttypes.Height{}
			}),
			exp: wasmvmtypes.IBCPacket{
				Data:     []byte("myData"),
				Src:      wasmvmtypes.IBCEndpoint{PortID: "srcPort", ChannelID: "channel-1"},
				Dest:     wasmvmtypes.IBCEndpoint{PortID: "destPort", ChannelID: "channel-2"},
				Sequence: 1,
				Timeout:  wasmvmtypes.IBCTimeout{Timestamp: myTimestamp},
			},
		}, "with time and height timeout": {
			src: IBCPacketFixture(func(p *channeltypes.Packet) {
				p.TimeoutTimestamp = myTimestamp
			}),
			exp: wasmvmtypes.IBCPacket{
				Data:     []byte("myData"),
				Src:      wasmvmtypes.IBCEndpoint{PortID: "srcPort", ChannelID: "channel-1"},
				Dest:     wasmvmtypes.IBCEndpoint{PortID: "destPort", ChannelID: "channel-2"},
				Sequence: 1,
				Timeout: wasmvmtypes.IBCTimeout{
					Block:     &wasmvmtypes.IBCTimeoutBlock{Height: 1, Revision: 2},
					Timestamp: myTimestamp,
				},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := newIBCPacket(spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func IBCPacketFixture(mutators ...func(p *channeltypes.Packet)) channeltypes.Packet {
	r := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "srcPort",
		SourceChannel:      "channel-1",
		DestinationPort:    "destPort",
		DestinationChannel: "channel-2",
		Data:               []byte("myData"),
		TimeoutHeight: clienttypes.Height{
			RevisionHeight: 1,
			RevisionNumber: 2,
		},
		TimeoutTimestamp: 0,
	}
	for _, m := range mutators {
		m(&r)
	}
	return r
}
