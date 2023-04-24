package types

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type AdventureSettings struct {
	// Flags is a set of flags that specify certain properties of the player, such as whether it can
	// fly and/or move through blocks.
	Flags uint32
	// CommandPermissionLevel is a set of permissions that specify what commands a player is allowed to execute.
	CommandPermissionLevel uint32
	// ActionPermissions is, much like Flags, a set of flags that specify actions that the player is allowed
	// to undertake, such as whether it is allowed to edit blocks, open doors etc.
	ActionPermissions uint32
	// PermissionLevel is the permission level of the player as it shows up in the player list built up using
	// the PlayerList packet.
	PermissionLevel uint32
	// CustomStoredPermissions ...
	CustomStoredPermissions uint32
	// PlayerUniqueID is a unique identifier of the player. It appears it is not required to fill this field
	// out with a correct value. Simply writing 0 seems to work.
	PlayerUniqueID int64
}

func (a *AdventureSettings) Marshal(r protocol.IO) {
	r.Varuint32(&a.Flags)
	r.Varuint32(&a.CommandPermissionLevel)
	r.Varuint32(&a.ActionPermissions)
	r.Varuint32(&a.PermissionLevel)
	r.Varuint32(&a.CustomStoredPermissions)
	r.Int64(&a.PlayerUniqueID)
}
