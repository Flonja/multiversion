package v486

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/internal/item"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

// downgradeBlockRuntimeID downgrades latest block runtime IDs to a legacy block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := latest.RuntimeIDToState(input)
	if !ok {
		return mappings.AirRID
	}
	runtimeID, ok := mappings.StateToRuntimeID(state)
	if !ok {
		return mappings.AirRID
	}
	return runtimeID
}

// downgradeItem downgrades the input item stack to a legacy item stack.
func downgradeItem(input protocol.ItemStack) protocol.ItemStack {
	if input.NetworkID == int32(latest.AirRID) {
		return protocol.ItemStack{}
	}

	name, _ := latest.ItemRuntimeIDToName(input.NetworkID)
	i := item.Downgrade(item.Item{
		Name:     name,
		Metadata: input.MetadataValue,
		Version:  latest.ItemVersion,
	}, mappings.ItemVersion)
	blockRuntimeId := uint32(0)
	if latestBlockState, ok := item.BlockStateFromItem(i); ok {
		rid, _ := latest.StateToRuntimeID(latestBlockState)
		blockRuntimeId = downgradeBlockRuntimeID(rid)
	}
	networkID, ok := mappings.ItemNameToRuntimeID(name)
	if !ok {
		networkID, _ = mappings.ItemNameToRuntimeID("minecraft:air")
		blockRuntimeId = mappings.AirRID
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
	r := cube.Range{-64, 319}
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	downgraded := chunk.New(mappings.AirRID, r)

	i := 0
	// First downgrade the blocks.
	for _, sub := range c.Sub()[start : len(c.Sub())-start] {
		downgraded.Sub()[i] = downgradeSubChunk(sub)
		i += 1
	}
}

// downgradeSubChunk downgrades a subchunk from the latest version to the legacy equivalent.
func downgradeSubChunk(sub *chunk.SubChunk) *chunk.SubChunk {
	downgraded := chunk.NewSubChunk(mappings.AirRID)

	for layerInd, layer := range sub.Layers() {
		downgradedLayer := downgraded.Layer(uint8(layerInd))
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				for y := uint8(0); y < 16; y++ {
					latestRuntimeID := layer.At(x, y, z)
					if latestRuntimeID == latest.AirRID {
						// Don't bother with air.
						continue
					}

					//downgradedLayer.Set(x, y, z, legacyAirRID)
					downgradedLayer.Set(x, y, z, downgradeBlockRuntimeID(latestRuntimeID))
				}
			}
		}
	}

	return downgraded
}

// downgradeEntityMetadata downgrades entity metadata from latest version to legacy version.
func downgradeEntityMetadata(data map[uint32]any) map[uint32]any {
	data = downgradeKey(data)

	var flag1, flag2 int64
	if v, ok := data[protocol.EntityDataKeyFlags]; ok {
		flag1 = v.(int64)
	}
	if v, ok := data[protocol.EntityDataKeyFlagsTwo]; ok {
		flag2 = v.(int64)
	}
	if flag1 == 0 && flag2 == 0 {
		return data
	}

	newFlag1 := flag1 & ^(^0 << (protocol.EntityDataFlagDash - 1))
	lastHalf := flag1 & (^0 << protocol.EntityDataFlagDash)
	lastHalf >>= 1
	lastHalf &= math.MaxInt64

	newFlag1 |= lastHalf

	if flag2 != 0 {
		newFlag1 ^= (flag2 & 1) << 63
		flag2 >>= 1
		flag2 &= math.MaxInt64

		data[protocol.EntityDataKeyFlagsTwo] = flag2
	}

	data[protocol.EntityDataKeyFlags] = newFlag1
	return data
}

// downgradeKey downgrades the latest key of an entity metadata map to the legacy key.
func downgradeKey(data map[uint32]any) map[uint32]any {
	newData := make(map[uint32]any)
	for key, value := range data {
		switch key {
		case protocol.EntityDataKeyBaseRuntimeID:
			key = 120
		case protocol.EntityDataKeyFreezingEffectStrength:
			key = 121
		case protocol.EntityDataKeyBuoyancyData:
			key = 122
		case protocol.EntityDataKeyGoatHornCount:
			key = 123
		case protocol.EntityDataKeyMovementSoundDistanceOffset:
			key = 125
		case protocol.EntityDataKeyHeartbeatIntervalTicks:
			key = 126
		case protocol.EntityDataKeyHeartbeatSoundEvent:
			key = 127
		case protocol.EntityDataKeyPlayerLastDeathPosition:
			key = 128
		case protocol.EntityDataKeyPlayerLastDeathDimension:
			key = 129
		case protocol.EntityDataKeyPlayerHasDied:
			key = 130
		}
		newData[key] = value
	}
	return newData
}

// TODO: add downgrade entity flags
// TODO: add downgrade command params
