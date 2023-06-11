package translator

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/mapping"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type BlockTranslator interface {
	// DowngradeBlockRuntimeID downgrades the input block runtime IDs to a legacy block runtime ID.
	DowngradeBlockRuntimeID(uint32) uint32
	// DowngradeChunk downgrades the input chunk to a legacy chunk.
	DowngradeChunk(*chunk.Chunk, bool)
	// DowngradeSubChunk downgrades the input sub chunk to a legacy sub chunk.
	DowngradeSubChunk(*chunk.SubChunk) *chunk.SubChunk
	// DowngradeBlockPackets downgrades the input block packets to legacy block packets.
	DowngradeBlockPackets([]packet.Packet, *minecraft.Conn) (result []packet.Packet)
	// UpgradeBlockRuntimeID upgrades the input block runtime IDs to the latest block runtime ID.
	UpgradeBlockRuntimeID(uint32) uint32
	// UpgradeChunk upgrades the input chunk to the latest chunk.
	UpgradeChunk(*chunk.Chunk, bool)
	// UpgradeSubChunk upgrades the input sub chunk to the latest sub chunk.
	UpgradeSubChunk(*chunk.SubChunk) *chunk.SubChunk
	// UpgradeBlockPackets upgrades the input block packets to the latest block packets.
	UpgradeBlockPackets([]packet.Packet, *minecraft.Conn) (result []packet.Packet)
}

type DefaultBlockTranslator struct {
	mapping mapping.Block
	latest  mapping.Block
}

func NewBlockTranslator(mapping mapping.Block, latestMapping mapping.Block) *DefaultBlockTranslator {
	return &DefaultBlockTranslator{mapping: mapping, latest: latestMapping}
}

func (t *DefaultBlockTranslator) DowngradeBlockRuntimeID(input uint32) uint32 {
	state, ok := t.latest.RuntimeIDToState(input)
	if !ok {
		return t.mapping.Air()
	}
	runtimeID, ok := t.mapping.StateToRuntimeID(state)
	if !ok {
		return t.mapping.Air()
	}
	return runtimeID
}

func (t *DefaultBlockTranslator) DowngradeChunk(input *chunk.Chunk, oldFormat bool) {
	start := 0
	r := world.Overworld.Range()
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	downgraded := chunk.New(t.mapping.Air(), r)

	i := 0
	// First downgrade the blocks.
	for _, sub := range input.Sub()[start : len(input.Sub())-start] {
		downgraded.Sub()[i] = t.DowngradeSubChunk(sub)
		i += 1
	}
}

func (t *DefaultBlockTranslator) DowngradeSubChunk(input *chunk.SubChunk) *chunk.SubChunk {
	downgraded := chunk.NewSubChunk(t.mapping.Air())

	for layerInd, layer := range input.Layers() {
		downgradedLayer := downgraded.Layer(uint8(layerInd))
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				for y := uint8(0); y < 16; y++ {
					latestRuntimeID := layer.At(x, y, z)
					downgradedLayer.Set(x, y, z, t.DowngradeBlockRuntimeID(latestRuntimeID))
				}
			}
		}
	}

	return downgraded
}

func (t *DefaultBlockTranslator) UpgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := t.mapping.RuntimeIDToState(input)
	if !ok {
		return t.latest.Air()
	}
	runtimeID, ok := t.latest.StateToRuntimeID(state)
	if !ok {
		return t.latest.Air()
	}
	return runtimeID
}

func (t *DefaultBlockTranslator) UpgradeChunk(c *chunk.Chunk, oldFormat bool) {
	start := 0
	r := world.Overworld.Range()
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	upgraded := chunk.New(t.latest.Air(), r)

	i := 0
	// First upgrade the blocks.
	for _, sub := range c.Sub()[start : len(c.Sub())-start] {
		upgraded.Sub()[i] = t.UpgradeSubChunk(sub)
		i += 1
	}
}

func (t *DefaultBlockTranslator) UpgradeSubChunk(sub *chunk.SubChunk) *chunk.SubChunk {
	upgraded := chunk.NewSubChunk(t.latest.Air())

	for layerInd, layer := range sub.Layers() {
		upgradedLayer := upgraded.Layer(uint8(layerInd))
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				for y := uint8(0); y < 16; y++ {
					legacyRuntimeID := layer.At(x, y, z)
					upgradedLayer.Set(x, y, z, t.UpgradeBlockRuntimeID(legacyRuntimeID))
				}
			}
		}
	}

	return upgraded
}

func (t *DefaultBlockTranslator) DowngradeBlockPackets(pks []packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	oldFormat := conn.GameData().BaseGameVersion == "1.17.40"
	for _, pk := range pks {
		switch pk := pk.(type) {
		case *packet.LevelChunk:
			count := int(pk.SubChunkCount)
			if count == protocol.SubChunkRequestModeLimitless || count == protocol.SubChunkRequestModeLimited {
				break
			}
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			buf := bytes.NewBuffer(pk.RawPayload)
			c, err := chunk.NetworkDecode(t.latest.Air(), buf, count, oldFormat, r)
			if err != nil {
				fmt.Println(err)
				continue
			}
			t.DowngradeChunk(c, oldFormat)

			payload, err := chunk.NetworkEncode(t.mapping.Air(), c, oldFormat)
			if err != nil {
				fmt.Println(err)
				continue
			}
			pk.RawPayload = payload
		case *packet.SubChunk:
			if conn.ClientCacheEnabled() {
				break
			}
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			for i, entry := range pk.SubChunkEntries {
				if entry.Result == protocol.SubChunkResultSuccess {
					buf := bytes.NewBuffer(entry.RawPayload)
					ind := byte(i)
					subChunk, err := chunk.DecodeSubChunk(t.latest.Air(), r, buf, &ind, chunk.NetworkEncoding)
					if err != nil {
						fmt.Println(err)
						continue
					}
					subChunk = t.DowngradeSubChunk(subChunk)
					entry.RawPayload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
					pk.SubChunkEntries[i] = entry
				}
			}
		case *packet.ClientCacheMissResponse:
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			for i, blob := range pk.Blobs {
				buf := bytes.NewBuffer(blob.Payload)
				ind := byte(0)
				subChunk, err := chunk.DecodeSubChunk(t.latest.Air(), r, buf, &ind, chunk.NetworkEncoding)
				if err != nil {
					// Has a possibility to be a biome, ignore then
					continue
				}
				subChunk = t.DowngradeSubChunk(subChunk)

				blob.Payload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
				pk.Blobs[i] = blob
			}
		case *packet.UpdateSubChunkBlocks:
			for i, block := range pk.Blocks {
				block.BlockRuntimeID = t.DowngradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
			for i, block := range pk.Extra {
				block.BlockRuntimeID = t.DowngradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Extra[i] = block
			}
		case *packet.UpdateBlock:
			pk.NewBlockRuntimeID = t.DowngradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.UpdateBlockSynced:
			pk.NewBlockRuntimeID = t.DowngradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.InventoryTransaction:
			if transactionData, ok := pk.TransactionData.(*protocol.UseItemTransactionData); ok {
				transactionData.BlockRuntimeID = t.DowngradeBlockRuntimeID(transactionData.BlockRuntimeID)
				pk.TransactionData = transactionData
			}
		case *packet.LevelEvent:
			switch pk.EventType {
			case packet.LevelEventParticleLegacyEvent | 20: // terrain
				fallthrough
			case packet.LevelEventParticlesDestroyBlock:
				fallthrough
			case packet.LevelEventParticlesDestroyBlockNoSound:
				pk.EventData = int32(t.DowngradeBlockRuntimeID(uint32(pk.EventData)))
			case packet.LevelEventParticlesCrackBlock:
				face := pk.EventData >> 24
				rid := t.DowngradeBlockRuntimeID(uint32(pk.EventData & 0xffff))
				pk.EventData = int32(rid) | (face << 24)
			}
		case *packet.LevelSoundEvent:
			switch pk.SoundType {
			case packet.SoundEventBreak:
				fallthrough
			case packet.SoundEventPlace:
				fallthrough
			case packet.SoundEventHit:
				fallthrough
			case packet.SoundEventLand:
				fallthrough
			case packet.SoundEventItemUseOn:
				pk.ExtraData = int32(t.DowngradeBlockRuntimeID(uint32(pk.ExtraData)))
			}
		}
		result = append(result, pk)
	}
	return result
}

func (t *DefaultBlockTranslator) UpgradeBlockPackets(pks []packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	oldFormat := conn.GameData().BaseGameVersion == "1.17.40"
	for _, pk := range pks {
		switch pk := pk.(type) {
		case *packet.LevelChunk:
			count := int(pk.SubChunkCount)
			if count == protocol.SubChunkRequestModeLimitless || count == protocol.SubChunkRequestModeLimited {
				break
			}
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			buf := bytes.NewBuffer(pk.RawPayload)
			c, err := chunk.NetworkDecode(t.mapping.Air(), buf, count, oldFormat, r)
			if err != nil {
				fmt.Println(err)
				continue
			}
			t.UpgradeChunk(c, oldFormat)

			payload, err := chunk.NetworkEncode(t.latest.Air(), c, oldFormat)
			if err != nil {
				fmt.Println(err)
				continue
			}
			pk.RawPayload = payload
		case *packet.SubChunk:
			if conn.ClientCacheEnabled() {
				break
			}
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			for i, entry := range pk.SubChunkEntries {
				if entry.Result == protocol.SubChunkResultSuccess {
					buf := bytes.NewBuffer(entry.RawPayload)
					ind := byte(i)
					subChunk, err := chunk.DecodeSubChunk(t.mapping.Air(), r, buf, &ind, chunk.NetworkEncoding)
					if err != nil {
						// Has a possibility to be a biome, ignore then
						continue
					}
					subChunk = t.UpgradeSubChunk(subChunk)
					entry.RawPayload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
					pk.SubChunkEntries[i] = entry
				}
			}
		case *packet.ClientCacheMissResponse:
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			for i, blob := range pk.Blobs {
				buf := bytes.NewBuffer(blob.Payload)
				ind := byte(0)
				subChunk, err := chunk.DecodeSubChunk(t.mapping.Air(), r, buf, &ind, chunk.NetworkEncoding)
				if err != nil {
					fmt.Println(err)
					continue
				}
				subChunk = t.UpgradeSubChunk(subChunk)

				blob.Payload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
				pk.Blobs[i] = blob
			}
		case *packet.UpdateSubChunkBlocks:
			for i, block := range pk.Blocks {
				block.BlockRuntimeID = t.UpgradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
			for i, block := range pk.Extra {
				block.BlockRuntimeID = t.UpgradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
		case *packet.UpdateBlock:
			pk.NewBlockRuntimeID = t.UpgradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.UpdateBlockSynced:
			pk.NewBlockRuntimeID = t.UpgradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.InventoryTransaction:
			if transactionData, ok := pk.TransactionData.(*protocol.UseItemTransactionData); ok {
				transactionData.BlockRuntimeID = t.UpgradeBlockRuntimeID(transactionData.BlockRuntimeID)
				pk.TransactionData = transactionData
			}
		case *packet.LevelEvent:
			switch pk.EventType {
			case packet.LevelEventParticleLegacyEvent | 20: // terrain
				fallthrough
			case packet.LevelEventParticlesDestroyBlock:
				fallthrough
			case packet.LevelEventParticlesDestroyBlockNoSound:
				pk.EventData = int32(t.UpgradeBlockRuntimeID(uint32(pk.EventData)))
			case packet.LevelEventParticlesCrackBlock:
				face := pk.EventData >> 24
				rid := t.UpgradeBlockRuntimeID(uint32(pk.EventData & 0xffff))
				pk.EventData = int32(rid) | (face << 24)
			}
		case *packet.LevelSoundEvent:
			switch pk.SoundType {
			case packet.SoundEventBreak:
				fallthrough
			case packet.SoundEventPlace:
				fallthrough
			case packet.SoundEventHit:
				fallthrough
			case packet.SoundEventLand:
				fallthrough
			case packet.SoundEventItemUseOn:
				pk.ExtraData = int32(t.UpgradeBlockRuntimeID(uint32(pk.ExtraData)))
			}
		}
		result = append(result, pk)
	}
	return result
}
