package v486

import (
	"github.com/flonja/multiversion/internal/item"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const latestItemVersion = 91

// upgradeBlockRuntimeID upgrades legacy block runtime IDs to a latest block runtime ID.
func upgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := mappings.RuntimeIDToState(input)
	if !ok {
		return latestAirRID
	}
	runtimeID, ok := latest.StateToRuntimeID(state)
	if !ok {
		return latestAirRID
	}
	return runtimeID
}

// upgradeItem upgrades the input item stack to a legacy item stack.
func upgradeItem(input protocol.ItemStack) protocol.ItemStack {
	name, _ := mappings.ItemRuntimeIDToName(input.NetworkID)
	i := item.Upgrade(item.Item{
		Name:     name,
		Metadata: input.MetadataValue,
		Version:  legacyItemVersion,
	}, latestItemVersion)
	blockRuntimeId := uint32(0)
	if latestBlockState, ok := item.BlockStateFromItem(i); ok {
		blockRuntimeId, _ = latest.StateToRuntimeID(latestBlockState)
	}
	networkID, ok := latest.ItemNameToRuntimeID(name)
	if !ok {
		networkID, _ = latest.ItemNameToRuntimeID("minecraft:air")
		blockRuntimeId = latestAirRID
	}
	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     networkID,
			MetadataValue: i.Metadata,
		},
		BlockRuntimeID: int32(blockRuntimeId),
		Count:          input.Count,
		NBTData:        input.NBTData,
		CanBePlacedOn:  input.CanBePlacedOn,
		CanBreak:       input.CanBreak,
		HasNetworkID:   input.HasNetworkID,
	}
}

// upgradeItemInstance upgrades the input item instance to a legacy item instance.
func upgradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance {
	return protocol.ItemInstance{
		StackNetworkID: input.StackNetworkID,
		Stack:          downgradeItem(input.Stack),
	}
}
