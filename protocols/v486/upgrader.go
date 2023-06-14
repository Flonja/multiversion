package v486

import (
	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// upgradeBlockActorData upgrades a block actor from a legacy version to the latest version.
func upgradeBlockActorData(data map[string]any) map[string]any {
	switch data["id"] {
	case "Sign":
		textRaw, ok := data["Text"]
		if !ok {
			textRaw = ""
		}
		text, ok := textRaw.(string)
		if !ok {
			text = ""
		}
		data["FrontText"] = map[string]any{"Text": text}
		data["BackText"] = map[string]any{"Text": ""}
	}
	return data
}

// upgradeEntityMetadata upgrades entity metadata from legacy version to latest version.
func upgradeEntityMetadata(data map[uint32]any) map[uint32]any {
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

func upgradeCraftingDescription(descriptor *types.DefaultItemDescriptor) protocol.ItemDescriptor {
	return &protocol.DefaultItemDescriptor{
		NetworkID:     int16(descriptor.NetworkID),
		MetadataValue: int16(descriptor.MetadataValue),
	}
}

// TODO: add upgrade entity flags
