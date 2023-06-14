package types

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ItemStackRequest represents a single request present in an ItemStackRequest packet sent by the client to
// change an item in an inventory.
// Item stack requests are either approved or rejected by the server using the ItemStackResponse packet.
type ItemStackRequest struct {
	protocol.ItemStackRequest
}

// Marshal encodes/decodes an ItemStackRequest.
func (x *ItemStackRequest) Marshal(r protocol.IO) {
	r.Varint32(&x.RequestID)
	protocol.FuncSlice(r, &x.Actions, r.StackRequestAction)
	protocol.FuncSlice(r, &x.FilterStrings, r.String)
}
