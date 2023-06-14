package v486

import (
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// downgradeBlockActorData downgrades a block actor from latest version to legacy version.
func downgradeBlockActorData(data map[string]any) map[string]any {
	switch data["id"] {
	case "Sign":
		delete(data, "BackText")
		frontRaw, ok := data["FrontText"]
		if !ok {
			frontRaw = map[string]any{"Text": ""}
		}
		front, ok := frontRaw.(map[string]any)
		if !ok {
			front = map[string]any{"Text": ""}
		}
		textRaw, ok := front["Text"]
		if !ok {
			textRaw = ""
		}
		text, ok := textRaw.(string)
		if !ok {
			text = ""
		}
		data["Text"] = text
	}
	return data
}

// downgradeEntityMetadata downgrades entity metadata from latest version to legacy version.
func downgradeEntityMetadata(data map[uint32]any) map[uint32]any {
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

func downgradeCraftingDescription(descriptor protocol.ItemDescriptor, m mapping.Item) protocol.ItemDescriptor {
	var networkId int32
	var metadata int32
	switch descriptor := descriptor.(type) {
	case *protocol.DefaultItemDescriptor:
		networkId = int32(descriptor.NetworkID)
		metadata = int32(descriptor.MetadataValue)
	case *protocol.DeferredItemDescriptor:
		if rid, ok := m.ItemNameToRuntimeID(descriptor.Name); ok {
			networkId = rid
			metadata = int32(descriptor.MetadataValue)
		}
	case *protocol.ItemTagItemDescriptor:
		/// ?????
	case *protocol.ComplexAliasItemDescriptor:
		/// ?????
	}
	return &types.DefaultItemDescriptor{
		NetworkID:     networkId,
		MetadataValue: metadata,
	}
}

// TODO: add downgrade entity flags
