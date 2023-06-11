package latest

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte

	Block = mapping.NewBlockMapping(blockStateData)
)
