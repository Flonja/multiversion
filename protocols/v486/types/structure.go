package types

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// StructureSettings is a struct holding settings of a structure block. Its fields may be changed using the
// in-game UI on the client-side.
type StructureSettings struct {
	protocol.StructureSettings
}

// Marshal reads/writes StructureSettings x using IO r.
func (x *StructureSettings) Marshal(r protocol.IO) {
	r.String(&x.PaletteName)
	r.Bool(&x.IgnoreEntities)
	r.Bool(&x.IgnoreBlocks)
	r.UBlockPos(&x.Size)
	r.UBlockPos(&x.Offset)
	r.Varint64(&x.LastEditingPlayerUniqueID)
	r.Uint8(&x.Rotation)
	r.Uint8(&x.Mirror)
	r.Uint8(&x.AnimationMode)
	r.Float32(&x.AnimationDuration)
	r.Float32(&x.Integrity)
	r.Uint32(&x.Seed)
	r.Vec3(&x.Pivot)
}
