package translator

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/flonja/multiversion/internal/item"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/packbuilder"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type ItemTranslator interface {
	// DowngradeItemType downgrades the input item type to a legacy item type.
	DowngradeItemType(input protocol.ItemType) protocol.ItemType
	// DowngradeItemStack downgrades the input item stack to a legacy item stack.
	DowngradeItemStack(input protocol.ItemStack) protocol.ItemStack
	// DowngradeItemInstance downgrades the input item instance to a legacy item instance.
	DowngradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance
	// DowngradeItemDescriptor downgrades the input item descriptor to a legacy item descriptor.
	DowngradeItemDescriptor(input protocol.ItemDescriptor) protocol.ItemDescriptor
	// DowngradeItemDescriptorCount downgrades the input item descriptor (with count) to a legacy item descriptor (with count).
	DowngradeItemDescriptorCount(input protocol.ItemDescriptorCount) protocol.ItemDescriptorCount
	DowngradeItemPackets(pks []packet.Packet, conn *minecraft.Conn) []packet.Packet
	// UpgradeItemType upgrades the input item type to the latest item type.
	UpgradeItemType(input protocol.ItemType) protocol.ItemType
	// UpgradeItemStack upgrades the input item stack to the latest item stack.
	UpgradeItemStack(input protocol.ItemStack) protocol.ItemStack
	// UpgradeItemInstance upgrades the input item instance to the latest item instance.
	UpgradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance
	// UpgradeItemDescriptor upgrades the input item descriptor to the latest item descriptor.
	UpgradeItemDescriptor(input protocol.ItemDescriptor) protocol.ItemDescriptor
	// UpgradeItemDescriptorCount upgrades the input item descriptor (with count) to the latest item descriptor (with count).
	UpgradeItemDescriptorCount(input protocol.ItemDescriptorCount) protocol.ItemDescriptorCount
	UpgradeItemPackets(pks []packet.Packet, conn *minecraft.Conn) []packet.Packet
	// Register registers a custom item entry.
	Register(item world.CustomItem, replacement string)
	// CustomItems lists all custom items used as substitutes, with the runtime id as the key
	CustomItems() map[int32]world.CustomItem
}

type DefaultItemTranslator struct {
	mapping            mapping.Item
	latest             mapping.Item
	blockMapping       mapping.Block
	blockMappingLatest mapping.Block
	ridToCustomItem    map[int32]world.CustomItem
	originalToCustom   map[int32]int32
	customToOriginal   map[int32]int32
}

func NewItemTranslator(mapping mapping.Item, latestMapping mapping.Item, blockMapping mapping.Block, blockMappingLatest mapping.Block) *DefaultItemTranslator {
	return &DefaultItemTranslator{mapping: mapping, latest: latestMapping, blockMapping: blockMapping, blockMappingLatest: blockMappingLatest,
		ridToCustomItem: make(map[int32]world.CustomItem), originalToCustom: make(map[int32]int32), customToOriginal: make(map[int32]int32)}
}

func (t *DefaultItemTranslator) DowngradeItemType(input protocol.ItemType) protocol.ItemType {
	if input.NetworkID == t.latest.Air() || input.NetworkID == 0 {
		return protocol.ItemType{
			NetworkID: t.mapping.Air(),
		}
	}
	networkID := input.NetworkID
	metadata := input.MetadataValue

	var ok bool
	if networkID, ok = t.originalToCustom[input.NetworkID]; !ok {
		name, _ := t.latest.ItemRuntimeIDToName(input.NetworkID)
		i := item.Downgrade(item.Item{
			Name:     name,
			Metadata: input.MetadataValue,
			Version:  t.latest.ItemVersion(),
		}, t.mapping.ItemVersion())
		metadata = i.Metadata

		networkID, ok = t.mapping.ItemNameToRuntimeID(i.Name)
		if !ok {
			networkID, _ = t.mapping.ItemNameToRuntimeID("minecraft:info_update")
		}
	}

	return protocol.ItemType{
		NetworkID:     networkID,
		MetadataValue: metadata,
	}
}

func (t *DefaultItemTranslator) DowngradeItemStack(input protocol.ItemStack) protocol.ItemStack {
	input.ItemType = t.DowngradeItemType(input.ItemType)

	blockRuntimeId := uint32(0)
	if input.NetworkID != t.mapping.Air() {
		name, _ := t.mapping.ItemRuntimeIDToName(input.NetworkID)
		if latestBlockState, ok := item.BlockStateFromItemName(name, input.MetadataValue); ok {
			var found bool
			if blockRuntimeId, found = t.blockMapping.StateToRuntimeID(latestBlockState); !found {
				blockRuntimeId = t.blockMapping.Air()
			}
		}
	}
	return protocol.ItemStack{
		ItemType:       input.ItemType,
		BlockRuntimeID: int32(blockRuntimeId),
		Count:          input.Count,
		NBTData:        input.NBTData,
		CanBePlacedOn:  input.CanBePlacedOn,
		CanBreak:       input.CanBreak,
		HasNetworkID:   input.HasNetworkID,
	}
}

func (t *DefaultItemTranslator) DowngradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance {
	input.Stack = t.DowngradeItemStack(input.Stack)
	return input
}

func (t *DefaultItemTranslator) DowngradeItemDescriptor(input protocol.ItemDescriptor) protocol.ItemDescriptor {
	switch descriptor := input.(type) {
	case *protocol.InvalidItemDescriptor:
		return input
	case *protocol.DefaultItemDescriptor:
		itemType := t.DowngradeItemType(protocol.ItemType{NetworkID: int32(descriptor.NetworkID), MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.NetworkID, descriptor.MetadataValue = int16(itemType.NetworkID), int16(itemType.MetadataValue)
		return descriptor
	case *protocol.MoLangItemDescriptor:
		return input
	case *protocol.ItemTagItemDescriptor:
		return input
	case *protocol.DeferredItemDescriptor:
		rid, ok := t.latest.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			descriptor.MetadataValue = 0
			return descriptor
		}
		itemType := t.DowngradeItemType(protocol.ItemType{NetworkID: rid, MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.MetadataValue = int16(itemType.MetadataValue)
		if name, ok := t.mapping.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	case *protocol.ComplexAliasItemDescriptor:
		rid, ok := t.latest.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			return descriptor
		}
		itemType := t.DowngradeItemType(protocol.ItemType{NetworkID: rid})
		if name, ok := t.mapping.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	}
	panic("unknown item descriptor")
}

func (t *DefaultItemTranslator) DowngradeItemDescriptorCount(input protocol.ItemDescriptorCount) protocol.ItemDescriptorCount {
	input.Descriptor = t.DowngradeItemDescriptor(input.Descriptor)
	return input
}

func (t *DefaultItemTranslator) UpgradeItemType(input protocol.ItemType) protocol.ItemType {
	if input.NetworkID == t.mapping.Air() || input.NetworkID == 0 {
		return protocol.ItemType{
			NetworkID: t.latest.Air(),
		}
	}
	networkID := input.NetworkID
	metadata := input.MetadataValue

	var ok bool
	if networkID, ok = t.customToOriginal[input.NetworkID]; !ok {
		name, _ := t.mapping.ItemRuntimeIDToName(input.NetworkID)
		i := item.Upgrade(item.Item{
			Name:     name,
			Metadata: input.MetadataValue,
			Version:  t.mapping.ItemVersion(),
		}, t.latest.ItemVersion())
		networkID, ok = t.latest.ItemNameToRuntimeID(i.Name)
		if !ok {
			networkID, _ = t.latest.ItemNameToRuntimeID("minecraft:info_update")
		}
	}

	return protocol.ItemType{
		NetworkID:     networkID,
		MetadataValue: metadata,
	}
}

func (t *DefaultItemTranslator) UpgradeItemStack(input protocol.ItemStack) protocol.ItemStack {
	input.ItemType = t.UpgradeItemType(input.ItemType)

	blockRuntimeId := uint32(0)
	if input.NetworkID != t.latest.Air() {
		name, _ := t.latest.ItemRuntimeIDToName(input.NetworkID)
		if latestBlockState, ok := item.BlockStateFromItemName(name, input.MetadataValue); ok {
			blockRuntimeId, _ = t.blockMappingLatest.StateToRuntimeID(latestBlockState)
		}
	}
	return protocol.ItemStack{
		ItemType:       input.ItemType,
		BlockRuntimeID: int32(blockRuntimeId),
		Count:          input.Count,
		NBTData:        input.NBTData,
		CanBePlacedOn:  input.CanBePlacedOn,
		CanBreak:       input.CanBreak,
		HasNetworkID:   input.HasNetworkID,
	}
}

func (t *DefaultItemTranslator) UpgradeItemInstance(input protocol.ItemInstance) protocol.ItemInstance {
	input.Stack = t.UpgradeItemStack(input.Stack)
	return input
}

func (t *DefaultItemTranslator) UpgradeItemDescriptor(input protocol.ItemDescriptor) protocol.ItemDescriptor {
	switch descriptor := input.(type) {
	case *protocol.InvalidItemDescriptor:
		return input
	case *protocol.DefaultItemDescriptor:
		itemType := t.UpgradeItemType(protocol.ItemType{NetworkID: int32(descriptor.NetworkID), MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.NetworkID, descriptor.MetadataValue = int16(itemType.NetworkID), int16(itemType.MetadataValue)
		return descriptor
	case *protocol.MoLangItemDescriptor:
		return input
	case *protocol.ItemTagItemDescriptor:
		return input
	case *protocol.DeferredItemDescriptor:
		rid, ok := t.mapping.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			descriptor.MetadataValue = 0
			return descriptor
		}
		itemType := t.UpgradeItemType(protocol.ItemType{NetworkID: rid, MetadataValue: uint32(descriptor.MetadataValue)})
		descriptor.MetadataValue = int16(itemType.MetadataValue)
		if name, ok := t.latest.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	case *protocol.ComplexAliasItemDescriptor:
		rid, ok := t.mapping.ItemNameToRuntimeID(descriptor.Name)
		descriptor.Name = "minecraft:air"
		if !ok {
			return descriptor
		}
		itemType := t.UpgradeItemType(protocol.ItemType{NetworkID: rid})
		if name, ok := t.latest.ItemRuntimeIDToName(itemType.NetworkID); ok {
			descriptor.Name = name
		}
		return descriptor
	}
	panic("unknown item descriptor")
}

func (t *DefaultItemTranslator) UpgradeItemDescriptorCount(input protocol.ItemDescriptorCount) protocol.ItemDescriptorCount {
	input.Descriptor = t.UpgradeItemDescriptor(input.Descriptor)
	return input
}

func (t *DefaultItemTranslator) DowngradeItemPackets(pks []packet.Packet, _ *minecraft.Conn) (result []packet.Packet) {
	for _, pk := range pks {
		switch pk := pk.(type) {
		case *packet.MobEquipment:
			pk.NewItem = t.DowngradeItemInstance(pk.NewItem)
		case *packet.MobArmourEquipment:
			pk.Helmet = t.DowngradeItemInstance(pk.Helmet)
			pk.Chestplate = t.DowngradeItemInstance(pk.Chestplate)
			pk.Leggings = t.DowngradeItemInstance(pk.Leggings)
			pk.Boots = t.DowngradeItemInstance(pk.Boots)
		case *packet.AddItemActor:
			pk.Item = t.DowngradeItemInstance(pk.Item)
		case *packet.AddPlayer:
			pk.HeldItem = t.DowngradeItemInstance(pk.HeldItem)
		case *packet.InventorySlot:
			pk.NewItem = t.DowngradeItemInstance(pk.NewItem)
		case *packet.InventoryContent:
			pk.Content = lo.Map(pk.Content, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return t.DowngradeItemInstance(item)
			})
		case *packet.ItemStackRequest:
			for i, request := range pk.Requests {
				for i2, action := range request.Actions {
					if act, ok := action.(*protocol.CraftResultsDeprecatedStackRequestAction); ok {
						act.ResultItems = lo.Map(act.ResultItems, func(item protocol.ItemStack, _ int) protocol.ItemStack {
							return t.DowngradeItemStack(item)
						})
						pk.Requests[i].Actions[i2] = act
					}
				}
			}
		case *packet.CraftingData:
			for i, recipe := range pk.Recipes {
				switch recipe := recipe.(type) {
				case *protocol.ShapelessRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.DowngradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.DowngradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.DowngradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.DowngradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.FurnaceRecipe:
					recipe.InputType = t.DowngradeItemType(recipe.InputType)
					recipe.Output = t.DowngradeItemStack(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.FurnaceDataRecipe:
					recipe.InputType = t.DowngradeItemType(recipe.InputType)
					recipe.Output = t.DowngradeItemStack(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.ShulkerBoxRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.DowngradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.DowngradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapelessChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.DowngradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.DowngradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.DowngradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.DowngradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.SmithingTransformRecipe:
					recipe.Template = t.DowngradeItemDescriptorCount(recipe.Template)
					recipe.Base = t.DowngradeItemDescriptorCount(recipe.Base)
					recipe.Addition = t.DowngradeItemDescriptorCount(recipe.Addition)
					recipe.Result = t.DowngradeItemStack(recipe.Result)
					pk.Recipes[i] = recipe
				case *protocol.SmithingTrimRecipe:
					recipe.Template = t.DowngradeItemDescriptorCount(recipe.Template)
					recipe.Base = t.DowngradeItemDescriptorCount(recipe.Base)
					recipe.Addition = t.DowngradeItemDescriptorCount(recipe.Addition)
				}
			}
			for i, recipe := range pk.PotionRecipes {
				itemType := t.DowngradeItemType(protocol.ItemType{NetworkID: recipe.InputPotionID, MetadataValue: uint32(recipe.InputPotionMetadata)})
				recipe.InputPotionID, recipe.InputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = t.DowngradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID, MetadataValue: uint32(recipe.ReagentItemMetadata)})
				recipe.ReagentItemID, recipe.ReagentItemMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = t.DowngradeItemType(protocol.ItemType{NetworkID: recipe.OutputPotionID, MetadataValue: uint32(recipe.OutputPotionMetadata)})
				recipe.OutputPotionID, recipe.OutputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				pk.PotionRecipes[i] = recipe
			}
			for i, recipe := range pk.PotionContainerChangeRecipes {
				itemType := t.DowngradeItemType(protocol.ItemType{NetworkID: recipe.InputItemID})
				recipe.InputItemID = itemType.NetworkID
				itemType = t.DowngradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID})
				recipe.ReagentItemID = itemType.NetworkID
				itemType = t.DowngradeItemType(protocol.ItemType{NetworkID: recipe.OutputItemID})
				recipe.OutputItemID = itemType.NetworkID
				pk.PotionContainerChangeRecipes[i] = recipe
			}
			for i, recipe := range pk.MaterialReducers {
				recipe.InputItem = t.DowngradeItemType(recipe.InputItem)
				for i2, output := range recipe.Outputs {
					itemType := t.DowngradeItemType(protocol.ItemType{NetworkID: output.NetworkID})
					output.NetworkID = itemType.NetworkID
					recipe.Outputs[i2] = output
				}
				pk.MaterialReducers[i] = recipe
			}
		case *packet.CraftingEvent:
			pk.Input = lo.Map(pk.Input, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return t.DowngradeItemInstance(item)
			})
			pk.Output = lo.Map(pk.Output, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return t.DowngradeItemInstance(item)
			})
		case *packet.PlayerAuthInput:
			for i, action := range pk.ItemStackRequest.Actions {
				if act, ok := action.(*protocol.CraftResultsDeprecatedStackRequestAction); ok {
					act.ResultItems = lo.Map(act.ResultItems, func(item protocol.ItemStack, _ int) protocol.ItemStack {
						return t.DowngradeItemStack(item)
					})
					pk.ItemStackRequest.Actions[i] = act
				}
			}
			for i, action := range pk.ItemInteractionData.Actions {
				action.OldItem = t.DowngradeItemInstance(action.OldItem)
				action.NewItem = t.DowngradeItemInstance(action.NewItem)
				pk.ItemInteractionData.Actions[i] = action
			}
			pk.ItemInteractionData.HeldItem = t.DowngradeItemInstance(pk.ItemInteractionData.HeldItem)
		case *packet.CreativeContent:
			for i, creativeItem := range pk.Items {
				creativeItem.Item = t.DowngradeItemStack(creativeItem.Item)

				pk.Items[i] = creativeItem
			}
		case *packet.InventoryTransaction:
			for i, action := range pk.Actions {
				action.OldItem = t.DowngradeItemInstance(action.OldItem)
				action.NewItem = t.DowngradeItemInstance(action.NewItem)
				pk.Actions[i] = action
			}
			switch transactionData := pk.TransactionData.(type) {
			case *protocol.UseItemTransactionData:
				transactionData.HeldItem = t.DowngradeItemInstance(transactionData.HeldItem)
				for i, action := range transactionData.Actions {
					action.OldItem = t.DowngradeItemInstance(action.OldItem)
					action.NewItem = t.DowngradeItemInstance(action.NewItem)
					transactionData.Actions[i] = action
				}
			case *protocol.UseItemOnEntityTransactionData:
				transactionData.HeldItem = t.DowngradeItemInstance(transactionData.HeldItem)
			case *protocol.ReleaseItemTransactionData:
				transactionData.HeldItem = t.DowngradeItemInstance(transactionData.HeldItem)
			}
		case *packet.LevelEvent:
			if pk.EventType == packet.LevelEventParticleLegacyEvent|14 { // egg crack
				itemType := t.DowngradeItemType(protocol.ItemType{
					NetworkID:     pk.EventData >> 16,
					MetadataValue: uint32(pk.EventData & 0xf),
				})
				pk.EventData = (itemType.NetworkID << 16) | int32(itemType.MetadataValue)
			}
		case *packet.StartGame:
			for i, entry := range pk.Items {
				if !entry.ComponentBased {
					itemType := t.DowngradeItemType(protocol.ItemType{
						NetworkID:     int32(entry.RuntimeID),
						MetadataValue: 0,
					})
					if itemType.NetworkID == t.mapping.Air() {
						removeIndex(pk.Items, i)
						continue
					}
					entry.RuntimeID = int16(itemType.NetworkID)

					var ok bool
					if entry.Name, ok = t.mapping.ItemRuntimeIDToName(itemType.NetworkID); !ok {
						panic(itemType)
					}
				} else {
					t.latest.RegisterEntry(entry.Name)
					entry.RuntimeID = int16(t.mapping.RegisterEntry(entry.Name))
				}
				pk.Items[i] = entry
			}
			for rid, i := range t.CustomItems() {
				name, _ := i.EncodeItem()
				pk.Items = append(pk.Items, protocol.ItemEntry{
					Name:           name,
					RuntimeID:      int16(rid),
					ComponentBased: true,
				})
			}
		case *packet.ItemComponent:
			for _, i := range t.CustomItems() {
				name, _ := i.EncodeItem()
				pk.Items = append(pk.Items, protocol.ItemComponentEntry{
					Name: name,
					Data: packbuilder.Components(i),
				})
			}
		}
		result = append(result, pk)
	}
	return result
}

func (t *DefaultItemTranslator) UpgradeItemPackets(pks []packet.Packet, _ *minecraft.Conn) (result []packet.Packet) {
	for _, pk := range pks {
		switch pk := pk.(type) {
		case *packet.MobEquipment:
			pk.NewItem = t.UpgradeItemInstance(pk.NewItem)
		case *packet.MobArmourEquipment:
			pk.Helmet = t.UpgradeItemInstance(pk.Helmet)
			pk.Chestplate = t.UpgradeItemInstance(pk.Chestplate)
			pk.Leggings = t.UpgradeItemInstance(pk.Leggings)
			pk.Boots = t.UpgradeItemInstance(pk.Boots)
		case *packet.AddItemActor:
			pk.Item = t.UpgradeItemInstance(pk.Item)
		case *packet.AddPlayer:
			pk.HeldItem = t.UpgradeItemInstance(pk.HeldItem)
		case *packet.InventorySlot:
			pk.NewItem = t.UpgradeItemInstance(pk.NewItem)
		case *packet.InventoryContent:
			pk.Content = lo.Map(pk.Content, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return t.UpgradeItemInstance(item)
			})
		case *packet.ItemStackRequest:
			for i, request := range pk.Requests {
				for i2, action := range request.Actions {
					if act, ok := action.(*protocol.CraftResultsDeprecatedStackRequestAction); ok {
						act.ResultItems = lo.Map(act.ResultItems, func(item protocol.ItemStack, _ int) protocol.ItemStack {
							return t.UpgradeItemStack(item)
						})
						pk.Requests[i].Actions[i2] = act
					}
				}
			}
		case *packet.CraftingData:
			for i, recipe := range pk.Recipes {
				switch recipe := recipe.(type) {
				case *protocol.ShapelessRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.UpgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.UpgradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.UpgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.UpgradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.FurnaceRecipe:
					recipe.InputType = t.UpgradeItemType(recipe.InputType)
					recipe.Output = t.UpgradeItemStack(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.FurnaceDataRecipe:
					recipe.InputType = t.UpgradeItemType(recipe.InputType)
					recipe.Output = t.UpgradeItemStack(recipe.Output)
					pk.Recipes[i] = recipe
				case *protocol.ShulkerBoxRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.UpgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.UpgradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapelessChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.UpgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.UpgradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.ShapedChemistryRecipe:
					for i2, input := range recipe.Input {
						recipe.Input[i2] = t.UpgradeItemDescriptorCount(input)
					}
					for i2, stack := range recipe.Output {
						recipe.Output[i2] = t.UpgradeItemStack(stack)
					}
					pk.Recipes[i] = recipe
				case *protocol.SmithingTransformRecipe:
					recipe.Template = t.UpgradeItemDescriptorCount(recipe.Template)
					recipe.Base = t.UpgradeItemDescriptorCount(recipe.Base)
					recipe.Addition = t.UpgradeItemDescriptorCount(recipe.Addition)
					recipe.Result = t.UpgradeItemStack(recipe.Result)
					pk.Recipes[i] = recipe
				case *protocol.SmithingTrimRecipe:
					recipe.Template = t.UpgradeItemDescriptorCount(recipe.Template)
					recipe.Base = t.UpgradeItemDescriptorCount(recipe.Base)
					recipe.Addition = t.UpgradeItemDescriptorCount(recipe.Addition)
				}
			}
			for i, recipe := range pk.PotionRecipes {
				itemType := t.UpgradeItemType(protocol.ItemType{NetworkID: recipe.InputPotionID, MetadataValue: uint32(recipe.InputPotionMetadata)})
				recipe.InputPotionID, recipe.InputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = t.UpgradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID, MetadataValue: uint32(recipe.ReagentItemMetadata)})
				recipe.ReagentItemID, recipe.ReagentItemMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				itemType = t.UpgradeItemType(protocol.ItemType{NetworkID: recipe.OutputPotionID, MetadataValue: uint32(recipe.OutputPotionMetadata)})
				recipe.OutputPotionID, recipe.OutputPotionMetadata = itemType.NetworkID, int32(itemType.MetadataValue)
				pk.PotionRecipes[i] = recipe
			}
			for i, recipe := range pk.PotionContainerChangeRecipes {
				itemType := t.UpgradeItemType(protocol.ItemType{NetworkID: recipe.InputItemID})
				recipe.InputItemID = itemType.NetworkID
				itemType = t.UpgradeItemType(protocol.ItemType{NetworkID: recipe.ReagentItemID})
				recipe.ReagentItemID = itemType.NetworkID
				itemType = t.UpgradeItemType(protocol.ItemType{NetworkID: recipe.OutputItemID})
				recipe.OutputItemID = itemType.NetworkID
				pk.PotionContainerChangeRecipes[i] = recipe
			}
			for i, recipe := range pk.MaterialReducers {
				recipe.InputItem = t.UpgradeItemType(recipe.InputItem)
				for i2, output := range recipe.Outputs {
					itemType := t.UpgradeItemType(protocol.ItemType{NetworkID: output.NetworkID})
					output.NetworkID = itemType.NetworkID
					recipe.Outputs[i2] = output
				}
				pk.MaterialReducers[i] = recipe
			}
		case *packet.CraftingEvent:
			pk.Input = lo.Map(pk.Input, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return t.UpgradeItemInstance(item)
			})
			pk.Output = lo.Map(pk.Output, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
				return t.UpgradeItemInstance(item)
			})
		case *packet.PlayerAuthInput:
			for i, action := range pk.ItemStackRequest.Actions {
				if act, ok := action.(*protocol.CraftResultsDeprecatedStackRequestAction); ok {
					act.ResultItems = lo.Map(act.ResultItems, func(item protocol.ItemStack, _ int) protocol.ItemStack {
						return t.UpgradeItemStack(item)
					})
					pk.ItemStackRequest.Actions[i] = act
				}
			}
			for i, action := range pk.ItemInteractionData.Actions {
				action.OldItem = t.UpgradeItemInstance(action.OldItem)
				action.NewItem = t.UpgradeItemInstance(action.NewItem)
				pk.ItemInteractionData.Actions[i] = action
			}
			pk.ItemInteractionData.HeldItem = t.UpgradeItemInstance(pk.ItemInteractionData.HeldItem)
		case *packet.CreativeContent:
			for i, creativeItem := range pk.Items {
				creativeItem.Item = t.UpgradeItemStack(creativeItem.Item)
				pk.Items[i] = creativeItem
			}
		case *packet.InventoryTransaction:
			for i, action := range pk.Actions {
				action.OldItem = t.UpgradeItemInstance(action.OldItem)
				action.NewItem = t.UpgradeItemInstance(action.NewItem)
				pk.Actions[i] = action
			}
			switch transactionData := pk.TransactionData.(type) {
			case *protocol.UseItemTransactionData:
				transactionData.HeldItem = t.UpgradeItemInstance(transactionData.HeldItem)
				for i, action := range transactionData.Actions {
					action.OldItem = t.UpgradeItemInstance(action.OldItem)
					action.NewItem = t.UpgradeItemInstance(action.NewItem)
					transactionData.Actions[i] = action
				}
			case *protocol.UseItemOnEntityTransactionData:
				transactionData.HeldItem = t.UpgradeItemInstance(transactionData.HeldItem)
			case *protocol.ReleaseItemTransactionData:
				transactionData.HeldItem = t.UpgradeItemInstance(transactionData.HeldItem)
			}
		case *packet.LevelEvent:
			if pk.EventType == packet.LevelEventParticleLegacyEvent|14 { // egg crack
				itemType := t.UpgradeItemType(protocol.ItemType{
					NetworkID:     pk.EventData >> 16,
					MetadataValue: uint32(pk.EventData & 0xf),
				})
				pk.EventData = (itemType.NetworkID << 16) | int32(itemType.MetadataValue)
			}
		case *packet.StartGame:
			for i, entry := range pk.Items {
				if !entry.ComponentBased {
					itemType := t.UpgradeItemType(protocol.ItemType{
						NetworkID:     int32(entry.RuntimeID),
						MetadataValue: 0,
					})
					entry.RuntimeID = int16(itemType.NetworkID)

					var ok bool
					if entry.Name, ok = t.latest.ItemRuntimeIDToName(itemType.NetworkID); !ok {
						panic(itemType)
					}
				} else {
					t.latest.RegisterEntry(entry.Name)
					entry.RuntimeID = int16(t.mapping.RegisterEntry(entry.Name))
				}
				pk.Items[i] = entry
			}
			for rid, i := range t.CustomItems() {
				name, _ := i.EncodeItem()
				pk.Items = append(pk.Items, protocol.ItemEntry{
					Name:           name,
					RuntimeID:      int16(rid),
					ComponentBased: true,
				})
			}
		case *packet.ItemComponent:
			for _, i := range t.CustomItems() {
				name, _ := i.EncodeItem()
				pk.Items = append(pk.Items, protocol.ItemComponentEntry{
					Name: name,
					Data: packbuilder.Components(i),
				})
			}
		}
		result = append(result, pk)
	}
	return result
}

func (t *DefaultItemTranslator) Register(item world.CustomItem, replacement string) {
	name, _ := item.EncodeItem()
	originalRid, ok := t.latest.ItemNameToRuntimeID(replacement)
	if !ok {
		panic(fmt.Errorf("%v not found in latest items", replacement))
	}
	if _, ok := t.originalToCustom[originalRid]; ok {
		panic(fmt.Errorf("%v is already mapped", replacement))
	}

	nextRID := t.mapping.RegisterEntry(name)
	t.ridToCustomItem[nextRID] = item
	t.originalToCustom[originalRid] = nextRID
	t.customToOriginal[nextRID] = originalRid
}

func (t *DefaultItemTranslator) CustomItems() map[int32]world.CustomItem {
	return t.ridToCustomItem
}

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
