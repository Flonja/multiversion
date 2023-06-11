package latest

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte
)

func NewBlockMapping() *mapping.DefaultBlockMapping {
	return mapping.NewBlockMapping(blockStateData)
}
