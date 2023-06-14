package types

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerListEntry is an entry found in the PlayerList packet. It represents a single player using the UUID
// found in the entry, and contains several properties such as the skin.
type PlayerListEntry struct {
	protocol.PlayerListEntry
}

// Marshal encodes/decodes a PlayerListEntry.
func (x *PlayerListEntry) Marshal(r protocol.IO) {
	r.UUID(&x.UUID)
	r.Varint64(&x.EntityUniqueID)
	r.String(&x.Username)
	r.String(&x.XUID)
	r.String(&x.PlatformChatID)
	r.Int32(&x.BuildPlatform)
	protocol.Single(r, &Skin{x.Skin})
	r.Bool(&x.Teacher)
	r.Bool(&x.Host)
}
