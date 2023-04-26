package v486

import (
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/internal/item"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const legacyItemVersion = 71

// downgradeBlockRuntimeID downgrades latest block runtime IDs to a legacy block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := latest.RuntimeIDToState(input)
	if !ok {
		return legacyAirRID
	}
	runtimeID, ok := mappings.StateToRuntimeID(state)
	if !ok {
		return legacyAirRID
	}
	return runtimeID
}

// downgradeItem downgrades the input item stack to a legacy item stack.
func downgradeItem(input protocol.ItemStack) protocol.ItemStack {
	if input.NetworkID == int32(latestAirRID) {
		return protocol.ItemStack{}
	}

	name, _ := latest.ItemRuntimeIDToName(input.NetworkID)
	i := item.Downgrade(item.Item{
		Name:     name,
		Metadata: input.MetadataValue,
		Version:  latestItemVersion,
	}, legacyItemVersion)
	blockRuntimeId := uint32(0)
	if latestBlockState, ok := item.BlockStateFromItem(i); ok {
		rid, _ := latest.StateToRuntimeID(latestBlockState)
		blockRuntimeId = downgradeBlockRuntimeID(rid)
	}
	networkID, ok := mappings.ItemNameToRuntimeID(name)
	if !ok {
		networkID, _ = mappings.ItemNameToRuntimeID("minecraft:air")
		blockRuntimeId = legacyAirRID
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

// downgradeItemInstance downgrades the input item instance to a legacy item instance.
func downgradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance {
	return protocol.ItemInstance{
		StackNetworkID: input.StackNetworkID,
		Stack:          downgradeItem(input.Stack),
	}
}

// downgradeChunk downgrades a chunk from the latest version to the legacy equivalent.
func downgradeChunk(c *chunk.Chunk, oldFormat bool) {
	start := 0
	if oldFormat {
		start = 4
	}

	// First downgrade the blocks.
	for subInd, sub := range c.Sub()[start : len(c.Sub())-start] {
		for layerInd, layer := range sub.Layers() {
			downgradedLayer := c.Sub()[subInd].Layer(uint8(layerInd))
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					for y := uint8(0); y < 16; y++ {
						latestRuntimeID := layer.At(x, y, z)
						if latestRuntimeID == latestAirRID {
							// Don't bother with air.
							continue
						}

						downgradedLayer.Set(x, y, z, downgradeBlockRuntimeID(latestRuntimeID))
					}
				}
			}
		}
	}
}
