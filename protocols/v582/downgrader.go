package v582

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/internal/item"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v582/mappings"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// downgradeItemType downgrades the input item type to a legacy item type.
func downgradeItemType(input protocol.ItemType) protocol.ItemType {
	if input.NetworkID == int32(latest.AirRID) || input.NetworkID == 0 {
		return protocol.ItemType{
			NetworkID: int32(mappings.AirRID),
		}
	}

	name, _ := latest.ItemRuntimeIDToName(input.NetworkID)
	i := item.Downgrade(item.Item{
		Name:     name,
		Metadata: input.MetadataValue,
		Version:  latest.ItemVersion,
	}, mappings.ItemVersion)
	networkID, ok := mappings.ItemNameToRuntimeID(i.Name)
	if !ok {
		// TODO: add substitute for unknown items
		networkID, _ = mappings.ItemNameToRuntimeID("minecraft:air")
	}
	return protocol.ItemType{
		NetworkID:     networkID,
		MetadataValue: i.Metadata,
	}
}

// downgradeItem downgrades the input item stack to a legacy item stack.
func downgradeItem(input protocol.ItemStack) protocol.ItemStack {
	input.ItemType = downgradeItemType(input.ItemType)

	blockRuntimeId := uint32(0)
	hasNetworkId := true
	if input.NetworkID != int32(mappings.AirRID) {
		name, _ := mappings.ItemRuntimeIDToName(input.NetworkID)
		if latestBlockState, ok := item.BlockStateFromItemName(name, input.MetadataValue); ok {
			rid, _ := latest.StateToRuntimeID(latestBlockState)
			blockRuntimeId = downgradeBlockRuntimeID(rid)
		}
	} else {
		blockRuntimeId = mappings.AirRID
		hasNetworkId = false
	}
	return protocol.ItemStack{
		ItemType:       input.ItemType,
		BlockRuntimeID: int32(blockRuntimeId),
		Count:          input.Count,
		NBTData:        input.NBTData,
		CanBePlacedOn:  input.CanBePlacedOn,
		CanBreak:       input.CanBreak,
		HasNetworkID:   hasNetworkId,
	}
}

// downgradeItemInstance downgrades the input item instance to a legacy item instance.
func downgradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance {
	input.Stack = downgradeItem(input.Stack)
	return input
}

// downgradeItemDescriptor downgrades the input item descriptor to a legacy item descriptor.
func downgradeItemDescriptor(input protocol.ItemDescriptor) protocol.ItemDescriptor {
	switch descriptor := input.(type) {
	case *protocol.InvalidItemDescriptor:
		return input
	case *protocol.DefaultItemDescriptor:
		itemType := downgradeItemType(protocol.ItemType{NetworkID: int32(descriptor.NetworkID), MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.NetworkID, descriptor.MetadataValue = int16(itemType.NetworkID), int16(itemType.MetadataValue)
		return descriptor
	case *protocol.MoLangItemDescriptor:
		return input
	case *protocol.ItemTagItemDescriptor:
		return input
	case *protocol.DeferredItemDescriptor:
		rid, ok := latest.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			descriptor.MetadataValue = 0
			return descriptor
		}
		itemType := downgradeItemType(protocol.ItemType{NetworkID: rid, MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.MetadataValue = int16(itemType.MetadataValue)
		if name, ok := mappings.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	case *protocol.ComplexAliasItemDescriptor:
		rid, ok := latest.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			return descriptor
		}
		itemType := downgradeItemType(protocol.ItemType{NetworkID: rid})
		if name, ok := mappings.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	}
	panic("unknown item descriptor")
}

// downgradeItemDescriptorCount downgrades the input item descriptor (with count) to a legacy item descriptor (with count).
func downgradeItemDescriptorCount(input protocol.ItemDescriptorCount) protocol.ItemDescriptorCount {
	input.Descriptor = downgradeItemDescriptor(input.Descriptor)
	return input
}

// downgradeBlockRuntimeID downgrades the input block runtime IDs to a legacy block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := latest.RuntimeIDToState(input)
	if !ok {
		return mappings.AirRID
	}
	runtimeID, ok := mappings.StateToRuntimeID(state)
	if !ok {
		return mappings.AirRID
	}
	return runtimeID
}

// downgradeChunk downgrades the input chunk to a legacy chunk.
func downgradeChunk(input *chunk.Chunk, oldFormat bool) {
	start := 0
	r := world.Overworld.Range()
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	downgraded := chunk.New(mappings.AirRID, r)

	i := 0
	// First downgrade the blocks.
	for _, sub := range input.Sub()[start : len(input.Sub())-start] {
		downgraded.Sub()[i] = downgradeSubChunk(sub)
		i += 1
	}
}

// downgradeSubChunk downgrades the input sub chunk to a legacy sub chunk.
func downgradeSubChunk(input *chunk.SubChunk) *chunk.SubChunk {
	downgraded := chunk.NewSubChunk(mappings.AirRID)

	for layerInd, layer := range input.Layers() {
		downgradedLayer := downgraded.Layer(uint8(layerInd))
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				for y := uint8(0); y < 16; y++ {
					latestRuntimeID := layer.At(x, y, z)
					if latestRuntimeID == latest.AirRID {
						// Don't bother with air.
						continue
					}

					downgradedLayer.Set(x, y, z, downgradeBlockRuntimeID(latestRuntimeID))
				}
			}
		}
	}

	return downgraded
}

func downgradeItemPackets(pks []packet.Packet) (result []packet.Packet) {
	for _, pk := range pks {
		switch pk := pk.(type) {
		case *packet.MobEquipment:
			pk.NewItem = downgradeItemInstance(pk.NewItem)
		case *packet.MobArmourEquipment:
			pk.Helmet = downgradeItemInstance(pk.Helmet)
			pk.Chestplate = downgradeItemInstance(pk.Chestplate)
			pk.Leggings = downgradeItemInstance(pk.Leggings)
			pk.Boots = downgradeItemInstance(pk.Boots)
		case *packet.AddItemActor:
			pk.Item = downgradeItemInstance(pk.Item)
		case *packet.AddPlayer:
			pk.HeldItem = downgradeItemInstance(pk.HeldItem)
		case *packet.InventorySlot:
			pk.NewItem = downgradeItemInstance(pk.NewItem)
		case *packet.InventoryContent:
			pk.Content = lo.Map(pk.Content, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return downgradeItemInstance(item)
			})
		case *packet.CraftingData:
			for i, recipe := range pk.Recipes {
				switch recipe := recipe.(type) {
				case *protocol.ShapelessRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = downgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = downgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = downgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = downgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.FurnaceRecipe:
					recipe.InputType = downgradeItemType(recipe.InputType)
					recipe.Output = downgradeItem(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.FurnaceDataRecipe:
					recipe.InputType = downgradeItemType(recipe.InputType)
					recipe.Output = downgradeItem(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.ShulkerBoxRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = downgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = downgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapelessChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = downgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = downgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = downgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = downgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.SmithingTransformRecipe:
					recipe.Template = downgradeItemDescriptorCount(recipe.Template)
					recipe.Base = downgradeItemDescriptorCount(recipe.Base)
					recipe.Addition = downgradeItemDescriptorCount(recipe.Addition)
					recipe.Result = downgradeItem(recipe.Result)
					pk.Recipes[i] = recipe
				case *protocol.SmithingTrimRecipe:
					recipe.Template = downgradeItemDescriptorCount(recipe.Template)
					recipe.Base = downgradeItemDescriptorCount(recipe.Base)
					recipe.Addition = downgradeItemDescriptorCount(recipe.Addition)
				}
			}
			for i, recipe := range pk.PotionRecipes {
				itemType := downgradeItemType(protocol.ItemType{NetworkID: recipe.InputPotionID, MetadataValue: uint32(recipe.InputPotionMetadata)})
				recipe.InputPotionID, recipe.InputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = downgradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID, MetadataValue: uint32(recipe.ReagentItemMetadata)})
				recipe.ReagentItemID, recipe.ReagentItemMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = downgradeItemType(protocol.ItemType{NetworkID: recipe.OutputPotionID, MetadataValue: uint32(recipe.OutputPotionMetadata)})
				recipe.OutputPotionID, recipe.OutputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				pk.PotionRecipes[i] = recipe
			}
			for i, recipe := range pk.PotionContainerChangeRecipes {
				itemType := downgradeItemType(protocol.ItemType{NetworkID: recipe.InputItemID})
				recipe.InputItemID = itemType.NetworkID
				itemType = downgradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID})
				recipe.ReagentItemID = itemType.NetworkID
				itemType = downgradeItemType(protocol.ItemType{NetworkID: recipe.OutputItemID})
				recipe.OutputItemID = itemType.NetworkID
				pk.PotionContainerChangeRecipes[i] = recipe
			}
			for i, recipe := range pk.MaterialReducers {
				recipe.InputItem = downgradeItemType(recipe.InputItem)
				for i2, output := range recipe.Outputs {
					itemType := downgradeItemType(protocol.ItemType{NetworkID: output.NetworkID})
					output.NetworkID = itemType.NetworkID
					recipe.Outputs[i2] = output
				}
				pk.MaterialReducers[i] = recipe
			}
		case *packet.CraftingEvent:
			pk.Input = lo.Map(pk.Input, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return downgradeItemInstance(item)
			})
			pk.Output = lo.Map(pk.Output, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return downgradeItemInstance(item)
			})
		case *packet.PlayerAuthInput:
			for i, action := range pk.ItemStackRequest.Actions {
				if act, ok := action.(*protocol.CraftResultsDeprecatedStackRequestAction); ok {
					lo.Map(act.ResultItems, func(item protocol.ItemStack, _ int) protocol.ItemStack {
						return downgradeItem(item)
					})
					pk.ItemStackRequest.Actions[i] = act
				}
			}
			for i, action := range pk.ItemInteractionData.Actions {
				action.OldItem = downgradeItemInstance(action.OldItem)
				action.NewItem = downgradeItemInstance(action.NewItem)
				pk.ItemInteractionData.Actions[i] = action
			}
			pk.ItemInteractionData.HeldItem = downgradeItemInstance(pk.ItemInteractionData.HeldItem)
		case *packet.CreativeContent:
			for i, creativeItem := range pk.Items {
				creativeItem.Item = downgradeItem(creativeItem.Item)

				pk.Items[i] = creativeItem
			}
		case *packet.InventoryTransaction:
			for i, action := range pk.Actions {
				action.OldItem = downgradeItemInstance(action.OldItem)
				action.NewItem = downgradeItemInstance(action.NewItem)
				pk.Actions[i] = action
			}
			switch transactionData := pk.TransactionData.(type) {
			case *protocol.UseItemTransactionData:
				transactionData.HeldItem = downgradeItemInstance(transactionData.HeldItem)
				for i, action := range transactionData.Actions {
					action.OldItem = downgradeItemInstance(action.OldItem)
					action.NewItem = downgradeItemInstance(action.NewItem)
					transactionData.Actions[i] = action
				}
			case *protocol.UseItemOnEntityTransactionData:
				transactionData.HeldItem = downgradeItemInstance(transactionData.HeldItem)
			case *protocol.ReleaseItemTransactionData:
				transactionData.HeldItem = downgradeItemInstance(transactionData.HeldItem)
			}
		case *packet.LevelEvent:
			if pk.EventType == packet.LevelEventParticleLegacyEvent|14 { // egg crack
				itemType := downgradeItemType(protocol.ItemType{
					NetworkID:     pk.EventData >> 16,
					MetadataValue: uint32(pk.EventData & 0xf),
				})
				pk.EventData = (itemType.NetworkID << 16) | int32(itemType.MetadataValue)
			}
		}
		result = append(result, pk)
	}
	return result
}

func downgradeWorldPackets(pks []packet.Packet, data minecraft.GameData, cache bool) (result []packet.Packet) {
	oldFormat := data.BaseGameVersion == "1.17.40"
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
			c, err := chunk.NetworkDecode(latest.AirRID, buf, count, oldFormat, r)
			if err != nil {
				fmt.Println(err)
				continue
			}
			downgradeChunk(c, oldFormat)

			payload, err := chunk.NetworkEncode(mappings.AirRID, c, oldFormat)
			if err != nil {
				fmt.Println(err)
				continue
			}
			pk.RawPayload = payload
		case *packet.SubChunk:
			if cache {
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
					subChunk, err := chunk.DecodeSubChunk(latest.AirRID, r, buf, &ind, chunk.NetworkEncoding)
					if err != nil {
						fmt.Println(err)
						continue
					}
					subChunk = downgradeSubChunk(subChunk)
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
				subChunk, err := chunk.DecodeSubChunk(latest.AirRID, r, buf, &ind, chunk.NetworkEncoding)
				if err != nil {
					fmt.Println(err)
					continue
				}
				subChunk = downgradeSubChunk(subChunk)

				blob.Payload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
				pk.Blobs[i] = blob
			}
		case *packet.UpdateSubChunkBlocks:
			for i, block := range pk.Blocks {
				block.BlockRuntimeID = downgradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
			for i, block := range pk.Extra {
				block.BlockRuntimeID = downgradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
		case *packet.UpdateBlock:
			pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.UpdateBlockSynced:
			pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.InventoryTransaction:
			if transactionData, ok := pk.TransactionData.(*protocol.UseItemTransactionData); ok {
				transactionData.BlockRuntimeID = downgradeBlockRuntimeID(transactionData.BlockRuntimeID)
				pk.TransactionData = transactionData
			}
		case *packet.LevelEvent:
			switch pk.EventType {
			case packet.LevelEventParticleLegacyEvent | 20: // terrain
				fallthrough
			case packet.LevelEventParticlesDestroyBlock:
				fallthrough
			case packet.LevelEventParticlesDestroyBlockNoSound:
				pk.EventData = int32(downgradeBlockRuntimeID(uint32(pk.EventData)))
			case packet.LevelEventParticlesCrackBlock:
				face := pk.EventData >> 24
				rid := downgradeBlockRuntimeID(uint32(pk.EventData & 0xffff))
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
				pk.ExtraData = int32(downgradeBlockRuntimeID(uint32(pk.ExtraData)))
			}
		}
		result = append(result, pk)
	}
	return result
}
