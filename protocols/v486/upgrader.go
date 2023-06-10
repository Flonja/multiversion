package v486

import (
	"github.com/flonja/multiversion/internal/item"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const latestItemVersion = 101

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
	if input.NetworkID == int32(legacyAirRID) {
		return protocol.ItemStack{}
	}

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
		Stack:          upgradeItem(input.Stack),
	}
}

// TODO: add upgradeChunk

// upgradeEntityMetadata upgrades entity metadata from legacy version to latest version.
func upgradeEntityMetadata(data map[uint32]any) map[uint32]any {
	data = upgradeKey(data)

	var flag1, flag2 int64
	if v, ok := data[protocol.EntityDataKeyFlags]; ok {
		flag1 = v.(int64)
	}
	if v, ok := data[protocol.EntityDataKeyFlagsTwo]; ok {
		flag2 = v.(int64)
	}

	flag2 <<= 1
	flag2 |= (flag1 >> 63) & 1

	newFlag1 := flag1 & ^(^0 << (protocol.EntityDataFlagDash - 1))
	lastHalf := flag1 & (^0<<protocol.EntityDataFlagDash - 1)
	lastHalf <<= 1
	newFlag1 |= lastHalf

	data[protocol.EntityDataKeyFlagsTwo] = flag2
	data[protocol.EntityDataKeyFlags] = newFlag1
	return data
}

// upgradeKey upgrades the legacy key of an entity metadata map to the latest key.
func upgradeKey(data map[uint32]any) map[uint32]any {
	newData := make(map[uint32]any)
	for key, value := range data {
		switch key {
		case 120:
			key = protocol.EntityDataKeyBaseRuntimeID
		case 121:
			key = protocol.EntityDataKeyFreezingEffectStrength
		case 122:
			key = protocol.EntityDataKeyBuoyancyData
		case 123:
			key = protocol.EntityDataKeyGoatHornCount
		case 125:
			key = protocol.EntityDataKeyMovementSoundDistanceOffset
		case 126:
			key = protocol.EntityDataKeyHeartbeatIntervalTicks
		case 127:
			key = protocol.EntityDataKeyHeartbeatSoundEvent
		case 128:
			key = protocol.EntityDataKeyPlayerLastDeathPosition
		case 129:
			key = protocol.EntityDataKeyPlayerLastDeathDimension
		case 130:
			key = protocol.EntityDataKeyPlayerHasDied
		}
		newData[key] = value
	}
	return newData
}

// TODO: add upgrade entity flags
// TODO: add upgrade command params
