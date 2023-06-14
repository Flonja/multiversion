package types

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// DefaultItemDescriptor represents an item descriptor for regular items. This is used for the significant majority of
// items.
type DefaultItemDescriptor struct {
	// NetworkID is the numerical network ID of the item. This is sometimes a positive ID, and sometimes a
	// negative ID, depending on what item it concerns.
	NetworkID int32
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int32
}

// Marshal ...
func (x *DefaultItemDescriptor) Marshal(r protocol.IO) {
	r.Varint32(&x.NetworkID)
	if x.NetworkID != 0 {
		r.Varint32(&x.MetadataValue)
	}
}
