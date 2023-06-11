package latest

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
)

const ItemVersion = 121

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte
)

func NewItemMapping() *mapping.DefaultItemMapping {
	return mapping.NewItemMapping(itemRuntimeIDData, ItemVersion)
}
