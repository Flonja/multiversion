package v486

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Protocol struct {
}

func (p Protocol) ID() int32 {
	return 486
}

func (p Protocol) Ver() string {
	return "1.18.12"
}

func (p Protocol) Packets() packet.Pool {
	pool := packet.NewPool()
	return pool
}

func (p Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	//TODO implement me
	panic("implement me")
}

var (
	// latestAirRID is the runtime ID of the air block in the latest version of the game.
	latestAirRID, _ = latest.StateToRuntimeID(blockupgrader.BlockState{Name: "minecraft:air"})
	// legacyAirRID is the runtime ID of the air block in the legacy version of the game.
	legacyAirRID, _ = mappings.StateToRuntimeID(blockupgrader.BlockState{Name: "minecraft:air"})
)

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.LevelChunk:
		buf := bytes.NewBuffer(pk.RawPayload)
		oldFormat := conn.GameData().BaseGameVersion == "1.17.40"
		c, err := chunk.NetworkDecode(latestAirRID, buf, int(pk.SubChunkCount), oldFormat, world.Overworld.Range())
		if err != nil {
			fmt.Println(err)
			return nil
		}
		downgradeChunk(c, oldFormat)

		payload, err := chunk.NetworkEncode(legacyAirRID, c, oldFormat)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		return []packet.Packet{
			&packet.LevelChunk{
				Position:        pk.Position,
				HighestSubChunk: pk.HighestSubChunk,
				SubChunkCount:   pk.SubChunkCount,
				CacheEnabled:    pk.CacheEnabled,
				BlobHashes:      pk.BlobHashes,
				RawPayload:      payload,
			},
		}
	}

	return []packet.Packet{pk}
}

// downgradeBlockRuntimeID downgrades latest block runtime IDs to a legacy block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := latest.RuntimeIDToState(input)
	if !ok {
		return legacyAirRID
	}
	runtimeID, ok := mappings.StateToRuntimeID(state)
	if !ok {
		return legacyAirRID
	}
	return runtimeID
}

// upgradeBlockRuntimeID upgrades legacy block runtime IDs to a latest block runtime ID.
func upgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := mappings.RuntimeIDToState(input)
	if !ok {
		return latestAirRID
	}
	runtimeID, ok := latest.StateToRuntimeID(state)
	if !ok {
		return latestAirRID
	}
	return runtimeID
}

// downgradeChunk downgrades a chunk from the latest version to the legacy equivalent.
func downgradeChunk(c *chunk.Chunk, oldFormat bool) {
	start := 0
	if oldFormat {
		start = 4
	}

	// First downgrade the blocks.
	for subInd, sub := range c.Sub()[start : len(c.Sub())-start] {
		for layerInd, layer := range sub.Layers() {
			downgradedLayer := c.Sub()[subInd].Layer(uint8(layerInd))
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					for y := uint8(0); y < 16; y++ {
						latestRuntimeID := layer.At(x, y, z)
						if latestRuntimeID == latestAirRID {
							// Don't bother with air.
							continue
						}

						downgradedLayer.Set(x, y, z, downgradeBlockRuntimeID(latestRuntimeID))
					}
				}
			}
		}
	}
}
