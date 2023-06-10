package mappings

import (
	_ "embed"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const ItemVersion = 111

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte
	// itemRuntimeIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemRuntimeIDsToNames = map[int32]string{}
	// itemNamesToRuntimeIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToRuntimeIDs = map[string]int32{}
)

// init initializes the items.
func init() {
	var items map[string]int32
	if err := nbt.Unmarshal(itemRuntimeIDData, &items); err != nil {
		panic(err)
	}
	for name, rid := range items {
		itemNamesToRuntimeIDs[name] = rid
		itemRuntimeIDsToNames[rid] = name
	}
}

// ItemRuntimeIDToName converts an item runtime ID to a string ID.
func ItemRuntimeIDToName(runtimeID int32) (name string, found bool) {
	name, ok := itemRuntimeIDsToNames[runtimeID]
	return name, ok
}

// ItemNameToRuntimeID converts a string ID to an item runtime ID.
func ItemNameToRuntimeID(name string) (runtimeID int32, found bool) {
	rid, ok := itemNamesToRuntimeIDs[name]
	return rid, ok
}
