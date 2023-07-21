package translator

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/mapping"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type BlockTranslator interface {
	// DowngradeBlockRuntimeID downgrades the input block runtime IDs to a legacy block runtime ID.
	DowngradeBlockRuntimeID(uint32) uint32
	// DowngradeChunk downgrades the input chunk to a legacy chunk.
	DowngradeChunk(*chunk.Chunk, bool) *chunk.Chunk
	// DowngradeSubChunk downgrades the input sub chunk to a legacy sub chunk.
	DowngradeSubChunk(*chunk.SubChunk)
	// DowngradeBlockPackets downgrades the input block packets to legacy block packets.
	DowngradeBlockPackets([]packet.Packet, *minecraft.Conn) (result []packet.Packet)
	// UpgradeBlockRuntimeID upgrades the input block runtime IDs to the latest block runtime ID.
	UpgradeBlockRuntimeID(uint32) uint32
	// UpgradeChunk upgrades the input chunk to the latest chunk.
	UpgradeChunk(*chunk.Chunk, bool) *chunk.Chunk
	// UpgradeSubChunk upgrades the input sub chunk to the latest sub chunk.
	UpgradeSubChunk(*chunk.SubChunk)
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

func (t *DefaultBlockTranslator) DowngradeChunk(input *chunk.Chunk, oldFormat bool) *chunk.Chunk {
	start := 0
	r := world.Overworld.Range()
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	downgraded := chunk.New(t.latest.Air(), r)

	i := 0
	// First downgrade the blocks.
	for _, sub := range input.Sub()[start : len(input.Sub())-start] {
		t.DowngradeSubChunk(sub)
		downgraded.Sub()[i] = sub
		i += 1
	}
	i = 0
	// Then downgrade the biome ids.
	for _, sub := range input.BiomeSub()[start : len(input.BiomeSub())-start] {
		// todo
		sub.Palette().Replace(func(v uint32) uint32 {
			return 0 // at least the client doesn't crash now
		})
		downgraded.BiomeSub()[i] = sub
		i += 1
	}
	return downgraded
}

func (t *DefaultBlockTranslator) DowngradeSubChunk(input *chunk.SubChunk) {
	for _, storage := range input.Layers() {
		storage.Palette().Replace(t.DowngradeBlockRuntimeID)
	}
}

func (t *DefaultBlockTranslator) downgradeEntityMetadata(metadata map[uint32]any) map[uint32]any {
	if latestRID, ok := metadata[protocol.EntityDataKeyVariant]; ok {
		metadata[protocol.EntityDataKeyVariant] = int32(t.DowngradeBlockRuntimeID(uint32(latestRID.(int32))))
	}
	return metadata
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

func (t *DefaultBlockTranslator) UpgradeChunk(input *chunk.Chunk, oldFormat bool) *chunk.Chunk {
	start := 0
	r := world.Overworld.Range()
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	upgraded := chunk.New(t.mapping.Air(), r)

	i := 0
	// First upgrade the blocks.
	for _, sub := range input.Sub()[start : len(input.Sub())-start] {
		t.UpgradeSubChunk(sub)
		upgraded.Sub()[i] = sub
		i += 1
	}
	i = 0
	// Then upgrade the biome ids.
	for _, sub := range input.BiomeSub()[start : len(input.BiomeSub())-start] {
		// todo
		upgraded.BiomeSub()[i] = sub
		i += 1
	}
	return upgraded
}

func (t *DefaultBlockTranslator) UpgradeSubChunk(input *chunk.SubChunk) {
	for _, storage := range input.Layers() {
		storage.Palette().Replace(t.UpgradeBlockRuntimeID)
	}
}

func (t *DefaultBlockTranslator) upgradeEntityMetadata(metadata map[uint32]any) map[uint32]any {
	if latestRID, ok := metadata[protocol.EntityDataKeyVariant]; ok {
		metadata[protocol.EntityDataKeyVariant] = int32(t.UpgradeBlockRuntimeID(uint32(latestRID.(int32))))
	}
	return metadata
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
			buf := bytes.NewBuffer(pk.RawPayload)
			writeBuf := bytes.NewBuffer(nil)
			if !pk.CacheEnabled && !conn.ClientCacheEnabled() {
				r := world.Overworld.Range()
				if oldFormat {
					r = cube.Range{0, 255}
				}

				c, err := chunk.NetworkDecode(t.latest.Air(), buf, count, oldFormat, r)
				if err != nil {
					//fmt.Println(err)
					break
				}
				t.DowngradeChunk(c, oldFormat)

				payload, err := chunk.NetworkEncode(t.mapping.Air(), c, oldFormat)
				if err != nil {
					//fmt.Println(err)
					break
				}
				writeBuf.Write(payload)
				pk.SubChunkCount = uint32(len(c.Sub()))
			}
			safeBytes := buf.Bytes()

			countBorder, err := buf.ReadByte()
			if err != nil {
				pk.RawPayload = append(writeBuf.Bytes(), safeBytes...)
				break
			}
			borderBytes := make([]byte, countBorder)
			if _, err = buf.Read(borderBytes); err != nil {
				pk.RawPayload = append(writeBuf.Bytes(), safeBytes...)
				break
			}
			writeBuf.WriteByte(countBorder)
			writeBuf.Write(borderBytes)

			enc := nbt.NewEncoderWithEncoding(writeBuf, nbt.NetworkLittleEndian)
			dec := nbt.NewDecoderWithEncoding(buf, nbt.NetworkLittleEndian)
			for {
				var decNbt map[string]any
				if err = dec.Decode(&decNbt); err != nil {
					break
				}
				t.mapping.DowngradeBlockActorData(decNbt)

				if err = enc.Encode(decNbt); err != nil {
					break
				}
			}
			pk.RawPayload = append(writeBuf.Bytes(), buf.Bytes()...)
		case *packet.SubChunk:
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			for i, entry := range pk.SubChunkEntries {
				if entry.Result == protocol.SubChunkResultSuccess {
					buf := bytes.NewBuffer(entry.RawPayload)
					writeBuf := bytes.NewBuffer(nil)
					if !pk.CacheEnabled && !conn.ClientCacheEnabled() {
						ind := byte(i)
						subChunk, err := chunk.DecodeSubChunk(t.latest.Air(), r, buf, &ind, chunk.NetworkEncoding)
						if err != nil {
							//fmt.Println(err)
							continue
						}
						t.DowngradeSubChunk(subChunk)
						writeBuf.Write(chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, chunk.SubChunkVersion9, r, int(ind)))
					}

					enc := nbt.NewEncoderWithEncoding(writeBuf, nbt.NetworkLittleEndian)
					dec := nbt.NewDecoderWithEncoding(buf, nbt.NetworkLittleEndian)
					for {
						var decNbt map[string]any
						if err := dec.Decode(&decNbt); err != nil {
							break
						}
						t.mapping.DowngradeBlockActorData(decNbt)

						if err := enc.Encode(decNbt); err != nil {
							break
						}
					}

					entry.RawPayload = append(writeBuf.Bytes(), buf.Bytes()...)
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
				t.DowngradeSubChunk(subChunk)

				blob.Payload = append(chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, chunk.SubChunkVersion9, r, int(ind)), buf.Bytes()...)
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
		case *packet.AddActor:
			if pk.EntityType != "minecraft:falling_block" {
				continue
			}
			pk.EntityMetadata = t.downgradeEntityMetadata(pk.EntityMetadata)
		case *packet.SetActorData:
			pk.EntityMetadata = t.downgradeEntityMetadata(pk.EntityMetadata)
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
			buf := bytes.NewBuffer(pk.RawPayload)
			writeBuf := bytes.NewBuffer(nil)
			if !pk.CacheEnabled && !conn.ClientCacheEnabled() {
				r := world.Overworld.Range()
				if oldFormat {
					r = cube.Range{0, 255}
				}

				c, err := chunk.NetworkDecode(t.mapping.Air(), buf, count, oldFormat, r)
				if err != nil {
					//fmt.Println(err)
					break
				}
				t.UpgradeChunk(c, oldFormat)

				payload, err := chunk.NetworkEncode(t.latest.Air(), c, oldFormat)
				if err != nil {
					//fmt.Println(err)
					break
				}
				writeBuf.Write(payload)
				pk.SubChunkCount = uint32(len(c.Sub()))
			}
			safeBytes := buf.Bytes()

			countBorder, err := buf.ReadByte()
			if err != nil {
				pk.RawPayload = append(writeBuf.Bytes(), safeBytes...)
				break
			}
			borderBytes := make([]byte, countBorder)
			if _, err = buf.Read(borderBytes); err != nil {
				pk.RawPayload = append(writeBuf.Bytes(), safeBytes...)
				break
			}
			writeBuf.WriteByte(countBorder)
			writeBuf.Write(borderBytes)

			enc := nbt.NewEncoderWithEncoding(writeBuf, nbt.NetworkLittleEndian)
			dec := nbt.NewDecoderWithEncoding(buf, nbt.NetworkLittleEndian)
			for {
				var decNbt map[string]any
				if err = dec.Decode(&decNbt); err != nil {
					break
				}
				t.mapping.UpgradeBlockActorData(decNbt)

				if err = enc.Encode(decNbt); err != nil {
					break
				}
			}
			pk.RawPayload = append(writeBuf.Bytes(), buf.Bytes()...)
		case *packet.SubChunk:
			r := world.Overworld.Range()
			if oldFormat {
				r = cube.Range{0, 255}
			}

			for i, entry := range pk.SubChunkEntries {
				if entry.Result == protocol.SubChunkResultSuccess {
					buf := bytes.NewBuffer(entry.RawPayload)
					writeBuf := bytes.NewBuffer(nil)
					if !pk.CacheEnabled && !conn.ClientCacheEnabled() {
						ind := byte(i)
						subChunk, err := chunk.DecodeSubChunk(t.mapping.Air(), r, buf, &ind, chunk.NetworkEncoding)
						if err != nil {
							// Has a possibility to be a biome, ignore then
							continue
						}
						t.UpgradeSubChunk(subChunk)
						writeBuf.Write(chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, chunk.SubChunkVersion9, r, int(ind)))
					}

					enc := nbt.NewEncoderWithEncoding(writeBuf, nbt.NetworkLittleEndian)
					dec := nbt.NewDecoderWithEncoding(buf, nbt.NetworkLittleEndian)
					for {
						var decNbt map[string]any
						if err := dec.Decode(&decNbt); err != nil {
							break
						}
						t.mapping.UpgradeBlockActorData(decNbt)

						if err := enc.Encode(decNbt); err != nil {
							break
						}
					}

					entry.RawPayload = append(writeBuf.Bytes(), buf.Bytes()...)
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
				t.UpgradeSubChunk(subChunk)

				blob.Payload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, chunk.SubChunkVersion9, r, int(ind))
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
		case *packet.AddActor:
			if pk.EntityType != "minecraft:falling_block" {
				continue
			}
			pk.EntityMetadata = t.upgradeEntityMetadata(pk.EntityMetadata)
		case *packet.SetActorData:
			pk.EntityMetadata = t.upgradeEntityMetadata(pk.EntityMetadata)
		}
		result = append(result, pk)
	}
	return result
}
