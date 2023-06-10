package item

import (
	"github.com/df-mc/worldupgrader/blockupgrader"
)

// Item ...
type Item struct {
	Name     string
	Metadata uint32
	Version  uint16
}

// Upgrade upgrades the given item using the registered item upgrade schemas.
func Upgrade(item Item, ver uint16) Item {
	version := item.Version
	for id, s := range schemas {
		if version > id || id > ver {
			continue
		}

		name := item.Name
		metadata := item.Metadata
		nameRenamed := false
		metadataRemapped := false

		if metadataCombinations, ok := s.RemappedMetas[name]; ok {
			name, metadataRemapped = metadataCombinations[item.Metadata]
			metadata = 0
		} else if name, nameRenamed = s.RenamedIDs[item.Name]; !nameRenamed {
			name = item.Name
		}

		if nameRenamed || metadataRemapped {
			item = Item{
				Name:     name,
				Metadata: metadata,
				Version:  id,
			}
		}
	}
	return item
}

// Downgrade downgrades the given item using the registered item upgrade schemas.
func Downgrade(item Item, ver uint16) Item {
	for i, id := range schemaIDs {
		s := schemas[uint16(id)]
		if uint16(id) > item.Version {
			continue
		}

		name := item.Name
		metadata := item.Metadata
		nameRenamed := false
		metadataRemapped := false

		for oldName, newName := range s.RenamedIDs {
			if newName == name {
				name = oldName
				nameRenamed = true
			}
		}

		if !nameRenamed {
			for oldName, m := range s.RemappedMetas {
				for meta, newName := range m {
					if newName == name {
						name = oldName
						metadata = meta
						nameRenamed = true
						metadataRemapped = true
					}
				}
				if metadataRemapped {
					break
				}
			}
		}

		if ver > item.Version || i == len(schemaIDs)-1 {
			break
		}

		return Downgrade(Item{
			Name:     name,
			Metadata: metadata,
			Version:  uint16(schemaIDs[i+1]),
		}, ver)
	}
	return item
}

func BlockStateFromItemName(itemName string, metadata uint32) (blockupgrader.BlockState, bool) {
	blockId, ok := itemToBlockIdMap[itemName]
	if !ok {
		return blockupgrader.BlockState{}, false
	}

	blockState, ok := blockStateMap[stateHash{
		Name:     blockId,
		Metadata: metadata,
	}]
	return blockState, ok
}

func BlockStateFromItem(item Item) (blockupgrader.BlockState, bool) {
	return BlockStateFromItemName(item.Name, item.Metadata)
}
