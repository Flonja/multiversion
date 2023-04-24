package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RemoveVolumeEntity indicates a volume entity to be removed from server to client.
type RemoveVolumeEntity struct {
	// EntityRuntimeID ...
	EntityRuntimeID uint64
}

// ID ...
func (*RemoveVolumeEntity) ID() uint32 {
	return packet.IDRemoveVolumeEntity
}

func (pk *RemoveVolumeEntity) Marshal(r protocol.IO) {
	r.Uint64(&pk.EntityRuntimeID)
}
