package item

import (
	"bytes"
	"embed"
	"encoding/json"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"regexp"
	"sort"
	"strconv"
)

var (
	//go:embed schemas/*.json
	schemasFS embed.FS
	//go:embed 1.12.0_item_id_to_block_id_map.json
	rawItemToBlockIdMap []byte
	//go:embed 1.12.0_to_1.18.10_blockstate_map.bin
	rawblockStateMap []byte
	// schemaIDs is a list of ids from all registered schemas.
	schemaIDs []int
	// schemas is a map of all registered item upgrade schemas.
	schemas map[uint16]schema
	// itemToBlockIdMap is a list of all registered item upgrade schemas.
	itemToBlockIdMap map[string]string
	// blockStateMap is a list of all legacy block states mapped to an upgraded version of it.
	blockStateMap map[stateHash]blockupgrader.BlockState

	filenameRegex = regexp.MustCompile(`(\d{4})_[\w.]+\.json`)
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
	schemas = make(map[uint16]schema)
	files, err := schemasFS.ReadDir("schemas")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		subMatches := filenameRegex.FindStringSubmatch(f.Name())
		if subMatches == nil {
			continue
		}
		id, err := strconv.Atoi(subMatches[1])
		if err != nil {
			continue
		}

		buf, err := schemasFS.ReadFile("schemas/" + f.Name())
		if err != nil {
			panic(err)
		}
		var s schema
		if err = json.Unmarshal(buf, &s); err != nil {
			panic(err)
		}
		schemas[uint16(id)] = s
	}
	schemaIDs = make([]int, 0, len(schemas))
	for k := range schemas {
		schemaIDs = append(schemaIDs, int(k))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(schemaIDs)))

	if err = json.Unmarshal(rawItemToBlockIdMap, &itemToBlockIdMap); err != nil {
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
