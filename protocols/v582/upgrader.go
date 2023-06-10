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

// upgradeItemType upgrades the input item type to the latest item type.
func upgradeItemType(input protocol.ItemType) protocol.ItemType {
	if input.NetworkID == int32(mappings.AirRID) {
		return protocol.ItemType{
			NetworkID: int32(latest.AirRID),
		}
	}

	name, _ := mappings.ItemRuntimeIDToName(input.NetworkID)
	i := item.Upgrade(item.Item{
		Name:     name,
		Metadata: input.MetadataValue,
		Version:  mappings.ItemVersion,
	}, latest.ItemVersion)
	networkID, ok := latest.ItemNameToRuntimeID(i.Name)
	if !ok {
		networkID, _ = latest.ItemNameToRuntimeID("minecraft:air")
	}
	return protocol.ItemType{
		NetworkID:     networkID,
		MetadataValue: i.Metadata,
	}
}

// upgradeItem upgrades the input item stack to the latest item stack.
func upgradeItem(input protocol.ItemStack) protocol.ItemStack {
	input.ItemType = upgradeItemType(input.ItemType)

	blockRuntimeId := uint32(0)
	hasNetworkId := true
	if input.NetworkID != int32(latest.AirRID) {
		name, _ := latest.ItemRuntimeIDToName(input.NetworkID)
		if latestBlockState, ok := item.BlockStateFromItemName(name, input.MetadataValue); ok {
			blockRuntimeId, _ = latest.StateToRuntimeID(latestBlockState)
		}
	} else {
		blockRuntimeId = latest.AirRID
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

// upgradeItemInstance upgrades the input item instance to the latest item instance.
func upgradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance {
	input.Stack = upgradeItem(input.Stack)
	return input
}

// upgradeItemDescriptor upgrades the input item descriptor to the latest item descriptor.
func upgradeItemDescriptor(input protocol.ItemDescriptor) protocol.ItemDescriptor {
	switch descriptor := input.(type) {
	case *protocol.InvalidItemDescriptor:
		return input
	case *protocol.DefaultItemDescriptor:
		itemType := upgradeItemType(protocol.ItemType{NetworkID: int32(descriptor.NetworkID), MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.NetworkID, descriptor.MetadataValue = int16(itemType.NetworkID), int16(itemType.MetadataValue)
		return descriptor
	case *protocol.MoLangItemDescriptor:
		return input
	case *protocol.ItemTagItemDescriptor:
		return input
	case *protocol.DeferredItemDescriptor:
		rid, ok := mappings.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			descriptor.MetadataValue = 0
			return descriptor
		}
		itemType := upgradeItemType(protocol.ItemType{NetworkID: rid, MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.MetadataValue = int16(itemType.MetadataValue)
		if name, ok := latest.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	case *protocol.ComplexAliasItemDescriptor:
		rid, ok := mappings.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			return descriptor
		}
		itemType := upgradeItemType(protocol.ItemType{NetworkID: rid})
		if name, ok := latest.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	}
	panic("unknown item descriptor")
}

// upgradeItemDescriptorCount upgrades the input item descriptor (with count) to the latest item descriptor (with count).
func upgradeItemDescriptorCount(input protocol.ItemDescriptorCount) protocol.ItemDescriptorCount {
	input.Descriptor = upgradeItemDescriptor(input.Descriptor)
	return input
}

// upgradeBlockRuntimeID upgrades the input block runtime ID to the latest block runtime ID.
func upgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := mappings.RuntimeIDToState(input)
	if !ok {
		return latest.AirRID
	}
	runtimeID, ok := latest.StateToRuntimeID(state)
	if !ok {
		return latest.AirRID
	}
	return runtimeID
}

// upgradeChunk upgrades the input chunk to the latest chunk.
func upgradeChunk(c *chunk.Chunk, oldFormat bool) {
	start := 0
	r := world.Overworld.Range()
	if oldFormat {
		start = 4
		r = cube.Range{0, 255}
	}
	upgraded := chunk.New(latest.AirRID, r)

	i := 0
	// First upgrade the blocks.
	for _, sub := range c.Sub()[start : len(c.Sub())-start] {
		upgraded.Sub()[i] = upgradeSubChunk(sub)
		i += 1
	}
}

// upgradeSubChunk upgrades the input sub chunk to the latest sub chunk.
func upgradeSubChunk(sub *chunk.SubChunk) *chunk.SubChunk {
	upgraded := chunk.NewSubChunk(latest.AirRID)

	for layerInd, layer := range sub.Layers() {
		upgradedLayer := upgraded.Layer(uint8(layerInd))
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				for y := uint8(0); y < 16; y++ {
					legacyRuntimeID := layer.At(x, y, z)
					if legacyRuntimeID == mappings.AirRID {
						// Don't bother with air.
						continue
					}

					upgradedLayer.Set(x, y, z, upgradeBlockRuntimeID(legacyRuntimeID))
				}
			}
		}
	}

	return upgraded
}

func upgradeItemPackets(pks []packet.Packet) (result []packet.Packet) {
	for _, pk := range pks {
		switch pk := pk.(type) {
		case *packet.MobEquipment:
			pk.NewItem = upgradeItemInstance(pk.NewItem)
		case *packet.MobArmourEquipment:
			pk.Helmet = upgradeItemInstance(pk.Helmet)
			pk.Chestplate = upgradeItemInstance(pk.Chestplate)
			pk.Leggings = upgradeItemInstance(pk.Leggings)
			pk.Boots = upgradeItemInstance(pk.Boots)
		case *packet.AddItemActor:
			pk.Item = upgradeItemInstance(pk.Item)
		case *packet.AddPlayer:
			pk.HeldItem = upgradeItemInstance(pk.HeldItem)
		case *packet.InventorySlot:
			pk.NewItem = upgradeItemInstance(pk.NewItem)
		case *packet.InventoryContent:
			pk.Content = lo.Map(pk.Content, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return upgradeItemInstance(item)
			})
		case *packet.CraftingData:
			for i, recipe := range pk.Recipes {
				switch recipe := recipe.(type) {
				case *protocol.ShapelessRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = upgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = upgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = upgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = upgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.FurnaceRecipe:
					recipe.InputType = upgradeItemType(recipe.InputType)
					recipe.Output = upgradeItem(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.FurnaceDataRecipe:
					recipe.InputType = upgradeItemType(recipe.InputType)
					recipe.Output = upgradeItem(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.ShulkerBoxRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = upgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = upgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapelessChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = upgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = upgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = upgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = upgradeItem(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.SmithingTransformRecipe:
					recipe.Template = upgradeItemDescriptorCount(recipe.Template)
					recipe.Base = upgradeItemDescriptorCount(recipe.Base)
					recipe.Addition = upgradeItemDescriptorCount(recipe.Addition)
					recipe.Result = upgradeItem(recipe.Result)
					pk.Recipes[i] = recipe
				case *protocol.SmithingTrimRecipe:
					recipe.Template = upgradeItemDescriptorCount(recipe.Template)
					recipe.Base = upgradeItemDescriptorCount(recipe.Base)
					recipe.Addition = upgradeItemDescriptorCount(recipe.Addition)
				}
			}
			for i, recipe := range pk.PotionRecipes {
				itemType := upgradeItemType(protocol.ItemType{NetworkID: recipe.InputPotionID, MetadataValue: uint32(recipe.InputPotionMetadata)})
				recipe.InputPotionID, recipe.InputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = upgradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID, MetadataValue: uint32(recipe.ReagentItemMetadata)})
				recipe.ReagentItemID, recipe.ReagentItemMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = upgradeItemType(protocol.ItemType{NetworkID: recipe.OutputPotionID, MetadataValue: uint32(recipe.OutputPotionMetadata)})
				recipe.OutputPotionID, recipe.OutputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				pk.PotionRecipes[i] = recipe
			}
			for i, recipe := range pk.PotionContainerChangeRecipes {
				itemType := upgradeItemType(protocol.ItemType{NetworkID: recipe.InputItemID})
				recipe.InputItemID = itemType.NetworkID
				itemType = upgradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID})
				recipe.ReagentItemID = itemType.NetworkID
				itemType = upgradeItemType(protocol.ItemType{NetworkID: recipe.OutputItemID})
				recipe.OutputItemID = itemType.NetworkID
				pk.PotionContainerChangeRecipes[i] = recipe
			}
			for i, recipe := range pk.MaterialReducers {
				recipe.InputItem = upgradeItemType(recipe.InputItem)
				for i2, output := range recipe.Outputs {
					itemType := upgradeItemType(protocol.ItemType{NetworkID: output.NetworkID})
					output.NetworkID = itemType.NetworkID
					recipe.Outputs[i2] = output
				}
				pk.MaterialReducers[i] = recipe
			}
		case *packet.CraftingEvent:
			pk.Input = lo.Map(pk.Input, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return upgradeItemInstance(item)
			})
			pk.Output = lo.Map(pk.Output, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return upgradeItemInstance(item)
			})
		case *packet.PlayerAuthInput:
			for i, action := range pk.ItemStackRequest.Actions {
				if act, ok := action.(*protocol.CraftResultsDeprecatedStackRequestAction); ok {
					lo.Map(act.ResultItems, func(item protocol.ItemStack, _ int) protocol.ItemStack {
						return upgradeItem(item)
					})
					pk.ItemStackRequest.Actions[i] = act
				}
			}
			for i, action := range pk.ItemInteractionData.Actions {
				action.OldItem = upgradeItemInstance(action.OldItem)
				action.NewItem = upgradeItemInstance(action.NewItem)
				pk.ItemInteractionData.Actions[i] = action
			}
			pk.ItemInteractionData.HeldItem = upgradeItemInstance(pk.ItemInteractionData.HeldItem)
		case *packet.CreativeContent:
			for i, creativeItem := range pk.Items {
				creativeItem.Item = upgradeItem(creativeItem.Item)
				pk.Items[i] = creativeItem
			}
		case *packet.InventoryTransaction:
			for i, action := range pk.Actions {
				action.OldItem = upgradeItemInstance(action.OldItem)
				action.NewItem = upgradeItemInstance(action.NewItem)
				pk.Actions[i] = action
			}
			switch transactionData := pk.TransactionData.(type) {
			case *protocol.UseItemTransactionData:
				transactionData.HeldItem = upgradeItemInstance(transactionData.HeldItem)
				for i, action := range transactionData.Actions {
					action.OldItem = upgradeItemInstance(action.OldItem)
					action.NewItem = upgradeItemInstance(action.NewItem)
					transactionData.Actions[i] = action
				}
			case *protocol.UseItemOnEntityTransactionData:
				transactionData.HeldItem = upgradeItemInstance(transactionData.HeldItem)
			case *protocol.ReleaseItemTransactionData:
				transactionData.HeldItem = upgradeItemInstance(transactionData.HeldItem)
			}
		case *packet.LevelEvent:
			if pk.EventType == packet.LevelEventParticleLegacyEvent|14 { // egg crack
				itemType := upgradeItemType(protocol.ItemType{
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

func upgradeWorldPackets(pks []packet.Packet, data minecraft.GameData, cache bool) (result []packet.Packet) {
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
			c, err := chunk.NetworkDecode(mappings.AirRID, buf, count, oldFormat, r)
			if err != nil {
				fmt.Println(err)
				continue
			}
			upgradeChunk(c, oldFormat)

			payload, err := chunk.NetworkEncode(latest.AirRID, c, oldFormat)
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
					subChunk, err := chunk.DecodeSubChunk(mappings.AirRID, r, buf, &ind, chunk.NetworkEncoding)
					if err != nil {
						fmt.Println(err)
						continue
					}
					subChunk = upgradeSubChunk(subChunk)
					entry.RawPayload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
					pk.SubChunkEntries[ind] = entry
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
				subChunk, err := chunk.DecodeSubChunk(mappings.AirRID, r, buf, &ind, chunk.NetworkEncoding)
				if err != nil {
					fmt.Println(err)
					continue
				}
				subChunk = upgradeSubChunk(subChunk)

				blob.Payload = chunk.EncodeSubChunk(subChunk, chunk.NetworkEncoding, r, ind)
				pk.Blobs[i] = blob
			}
		case *packet.UpdateSubChunkBlocks:
			for i, block := range pk.Blocks {
				block.BlockRuntimeID = upgradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
			for i, block := range pk.Extra {
				block.BlockRuntimeID = upgradeBlockRuntimeID(block.BlockRuntimeID)
				pk.Blocks[i] = block
			}
		case *packet.UpdateBlock:
			pk.NewBlockRuntimeID = upgradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.UpdateBlockSynced:
			pk.NewBlockRuntimeID = upgradeBlockRuntimeID(pk.NewBlockRuntimeID)
		case *packet.InventoryTransaction:
			if transactionData, ok := pk.TransactionData.(*protocol.UseItemTransactionData); ok {
				transactionData.BlockRuntimeID = upgradeBlockRuntimeID(transactionData.BlockRuntimeID)
				pk.TransactionData = transactionData
			}
		case *packet.LevelEvent:
			switch pk.EventType {
			case packet.LevelEventParticleLegacyEvent | 20: // terrain
				fallthrough
			case packet.LevelEventParticlesDestroyBlock:
				fallthrough
			case packet.LevelEventParticlesDestroyBlockNoSound:
				pk.EventData = int32(upgradeBlockRuntimeID(uint32(pk.EventData)))
			case packet.LevelEventParticlesCrackBlock:
				face := pk.EventData >> 24
				rid := upgradeBlockRuntimeID(uint32(pk.EventData & 0xffff))
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
				pk.ExtraData = int32(upgradeBlockRuntimeID(uint32(pk.ExtraData)))
			}
		}
		result = append(result, pk)
	}
	return result
}
