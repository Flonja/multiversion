package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// AddVolumeEntity sends a volume entity's definition and metadata from server to client.
type AddVolumeEntity struct {
	// EntityRuntimeID is the runtime ID of the volume. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the volume functions or looks.
	EntityMetadata map[string]any
	// EncodingIdentifier is the unique identifier for the volume. It must be of the form 'namespace:name', where
	// namespace cannot be 'minecraft'.
	EncodingIdentifier string
	// InstanceIdentifier is the identifier of a fog definition.
	InstanceIdentifier string
	// EngineVersion is the engine version the entity is using, for example, '1.17.0'.
	EngineVersion string
}

// ID ...
func (*AddVolumeEntity) ID() uint32 {
	return packet.IDAddVolumeEntity
}

func (pk *AddVolumeEntity) Marshal(r protocol.IO) {
	r.Uint64(&pk.EntityRuntimeID)
	r.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	r.String(&pk.EncodingIdentifier)
	r.String(&pk.InstanceIdentifier)
	r.String(&pk.EngineVersion)
}
