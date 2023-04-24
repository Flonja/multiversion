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
