package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// LecternUpdate is sent by the client to update the server on which page was opened in a book on a lectern,
// or if the book should be removed from it.
type LecternUpdate struct {
	// Page is the page number in the book that was opened by the player on the lectern.
	Page byte
	// PageCount is the number of pages that the book opened in the lectern has.
	PageCount byte
	// Position is the position of the lectern that was updated. If no lectern is at the block position,
	// the packet should be ignored.
	Position protocol.BlockPos
	// DropBook specifies if the book currently set on display in the lectern should be dropped server-side.
	DropBook bool
}

// ID ...
func (*LecternUpdate) ID() uint32 {
	return packet.IDLecternUpdate
}

func (pk *LecternUpdate) Marshal(io protocol.IO) {
	io.Uint8(&pk.Page)
	io.Uint8(&pk.PageCount)
	io.UBlockPos(&pk.Position)
	io.Bool(&pk.DropBook)
}
