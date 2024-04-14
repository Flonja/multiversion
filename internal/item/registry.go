package item

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

var (
	//go:embed 1.12.0_item_id_to_block_id_map.json
	rawItemToBlockIdMap []byte
	//go:embed 1.12.0_to_1.18.10_blockstate_map.bin
	rawblockStateMap []byte
	// itemToBlockIdMap is a list of all registered item upgrade schemas.
	itemToBlockIdMap map[string]string
	// blockStateMap is a list of all legacy block states mapped to an upgraded version of it.
	blockStateMap map[stateHash]blockupgrader.BlockState
)

// stateHash is a struct that may be used as a map key for block states.
type stateHash struct {
	// Name is the name of the block. It looks like 'minecraft:stone'.
	Name string
	// Metadata is the metadata value of the block. A lot of blocks only have 0 as data value, but some blocks
	// carry specific variants or properties encoded in the metadata.
	Metadata uint32
}

// init ...
func init() {
	if err := json.Unmarshal(rawItemToBlockIdMap, &itemToBlockIdMap); err != nil {
		panic(err)
	}

	blockStateMap = make(map[stateHash]blockupgrader.BlockState)
	buf := protocol.NewReader(bytes.NewBuffer(rawblockStateMap), 0, false)
	var length uint32
	buf.Varuint32(&length)
	for i := uint32(0); i < length; i++ {
		var legacyStringId string
		buf.String(&legacyStringId)

		var pairs uint32
		buf.Varuint32(&pairs)
		for y := uint32(0); y < pairs; y++ {
			var meta uint32
			buf.Varuint32(&meta)

			var blockStateRaw map[string]any
			buf.NBT(&blockStateRaw, nbt.LittleEndian)
			latestBlockState := blockupgrader.Upgrade(blockupgrader.BlockState{
				Name:       blockStateRaw["name"].(string),
				Properties: blockStateRaw["states"].(map[string]any),
				Version:    blockStateRaw["version"].(int32),
			})
			blockStateMap[stateHash{
				Name:     legacyStringId,
				Metadata: meta,
			}] = latestBlockState
		}
	}
}
