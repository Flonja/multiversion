package v486

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

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
