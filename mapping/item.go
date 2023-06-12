package mapping

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type Item interface {
	// ItemRuntimeIDToName converts an item runtime ID to a string ID.
	ItemRuntimeIDToName(int32) (string, bool)
	// ItemNameToRuntimeID converts a string ID to an item runtime ID.
	ItemNameToRuntimeID(string) (int32, bool)
	RegisterEntry(name string) int32
	Air() int32
	ItemVersion() uint16
}

type DefaultItemMapping struct {
	// itemRuntimeIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemRuntimeIDsToNames map[int32]string
	// itemNamesToRuntimeIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToRuntimeIDs map[string]int32
	airRID                int32
	itemVersion           uint16
}

func NewItemMapping(raw []byte, itemVersion uint16) *DefaultItemMapping {
	itemRuntimeIDsToNames := make(map[int32]string)
	itemNamesToRuntimeIDs := make(map[string]int32)
	var airRID *int32

	var items map[string]int32
	if err := nbt.Unmarshal(raw, &items); err != nil {
		panic(err)
	}
	for name, rid := range items {
		if name == "minecraft:air" {
			airRID = &rid
		}

		itemNamesToRuntimeIDs[name] = rid
		itemRuntimeIDsToNames[rid] = name
	}
	if airRID == nil {
		panic("couldn't find air")
	}

	return &DefaultItemMapping{itemRuntimeIDsToNames: itemRuntimeIDsToNames, itemNamesToRuntimeIDs: itemNamesToRuntimeIDs,
		itemVersion: itemVersion}
}

func (m *DefaultItemMapping) ItemRuntimeIDToName(runtimeID int32) (name string, found bool) {
	name, ok := m.itemRuntimeIDsToNames[runtimeID]
	return name, ok
}

func (m *DefaultItemMapping) ItemNameToRuntimeID(name string) (runtimeID int32, found bool) {
	rid, ok := m.itemNamesToRuntimeIDs[name]
	return rid, ok
}

func (m *DefaultItemMapping) RegisterEntry(name string) int32 {
	nextRID := int32(len(m.itemRuntimeIDsToNames))
	m.itemNamesToRuntimeIDs[name] = nextRID
	m.itemRuntimeIDsToNames[nextRID] = name
	return nextRID
}

func (m *DefaultItemMapping) Air() int32 {
	return m.airRID
}

func (m *DefaultItemMapping) ItemVersion() uint16 {
	return m.itemVersion
}
