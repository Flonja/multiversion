package mapping

import (
	"github.com/df-mc/worldupgrader/itemupgrader"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type Item interface {
	// ItemRuntimeIDToName converts an item runtime ID to a string ID.
	ItemRuntimeIDToName(int32) (itemupgrader.ItemMeta, bool)
	// ItemNameToRuntimeID converts a string ID to an item runtime ID.
	ItemNameToRuntimeID(itemupgrader.ItemMeta) (int32, bool)
	RegisterEntry(string) int32
	Air() int32
}

type DefaultItemMapping struct {
	// itemRuntimeIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemRuntimeIDsToNames map[int32]itemupgrader.ItemMeta
	// itemNamesToRuntimeIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToRuntimeIDs map[itemupgrader.ItemMeta]int32
	airRID                int32
}

func NewItemMapping(raw []byte) *DefaultItemMapping {
	itemRuntimeIDsToNames := make(map[int32]itemupgrader.ItemMeta)
	itemNamesToRuntimeIDs := make(map[itemupgrader.ItemMeta]int32)
	var airRID *int32

	var items map[string]int32
	if err := nbt.Unmarshal(raw, &items); err != nil {
		panic(err)
	}
	for name, rid := range items {
		if name == "minecraft:air" {
			airRID = &rid
		}

		itemMeta := itemupgrader.Upgrade(itemupgrader.ItemMeta{Name: name})
		itemNamesToRuntimeIDs[itemMeta] = rid
		itemRuntimeIDsToNames[rid] = itemMeta
	}
	if airRID == nil {
		panic("couldn't find air")
	}

	return &DefaultItemMapping{itemRuntimeIDsToNames: itemRuntimeIDsToNames, itemNamesToRuntimeIDs: itemNamesToRuntimeIDs}
}

func (m *DefaultItemMapping) ItemRuntimeIDToName(runtimeID int32) (itemMeta itemupgrader.ItemMeta, found bool) {
	itemMeta, ok := m.itemRuntimeIDsToNames[runtimeID]
	return itemMeta, ok
}

func (m *DefaultItemMapping) ItemNameToRuntimeID(itemMeta itemupgrader.ItemMeta) (runtimeID int32, found bool) {
	rid, ok := m.itemNamesToRuntimeIDs[itemupgrader.Upgrade(itemMeta)]
	return rid, ok
}

func (m *DefaultItemMapping) RegisterEntry(name string) int32 {
	nextRID := int32(len(m.itemRuntimeIDsToNames))
	itemMeta := itemupgrader.Upgrade(itemupgrader.ItemMeta{Name: name})
	m.itemNamesToRuntimeIDs[itemMeta] = nextRID
	m.itemRuntimeIDsToNames[nextRID] = itemMeta
	return nextRID
}

func (m *DefaultItemMapping) Air() int32 {
	return m.airRID
}
