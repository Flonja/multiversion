package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Emote is sent by both the server and the client. When the client sends an emote, it sends this packet to
// the server, after which the server will broadcast the packet to other players online.
type Emote struct {
	// EntityRuntimeID is the entity that sent the emote. When a player sends this packet, it has this field
	// set as its own entity runtime ID.
	EntityRuntimeID uint64
	// EmoteID is the ID of the emote to send.
	EmoteID string
	// Flags is a combination of flags that change the way the Emote packet operates. When the server sends
	// this packet to other players, EmoteFlagServerSide must be present.
	Flags byte
}

// ID ...
func (*Emote) ID() uint32 {
	return packet.IDEmote
}

func (pk *Emote) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	io.String(&pk.EmoteID)
	io.Uint8(&pk.Flags)
}
