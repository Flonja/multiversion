package item

import (
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/df-mc/worldupgrader/itemupgrader"
)

func BlockStateFromItemName(itemName string, metadata uint32) (blockupgrader.BlockState, bool) {
	blockId, ok := itemToBlockIdMap[itemName]
	if !ok {
		return blockupgrader.BlockState{}, false
	}

	blockState, ok := blockStateMap[stateHash{
		Name:     blockId,
		Metadata: metadata,
	}]
	if !ok {
		blockState, ok = blockStateMap[stateHash{
			Name: blockId,
		}]
	}
	return blockState, ok
}

func BlockStateFromItem(item itemupgrader.ItemMeta) (blockupgrader.BlockState, bool) {
	return BlockStateFromItemName(item.Name, uint32(item.Meta))
}
