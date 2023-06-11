package v486

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

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
