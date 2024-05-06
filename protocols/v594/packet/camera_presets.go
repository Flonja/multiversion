package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// CameraPresets gives the client a list of custom camera presets.
type CameraPresets struct {
	// Data is a compound tag of the presets being set.
	Data map[string]any
}

// ID ...
func (*CameraPresets) ID() uint32 {
	return packet.IDCameraPresets
}

func (pk *CameraPresets) Marshal(io protocol.IO) {
	io.NBT(&pk.Data, nbt.NetworkLittleEndian)
}
