package latest

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
)

const ItemVersion = 121

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte

	Item = mapping.NewItemMapping(itemRuntimeIDData, ItemVersion)
)
