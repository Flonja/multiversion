package v486

import (
	_ "embed"
	"encoding/json"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v486/packet"
	"github.com/flonja/multiversion/protocols/v486/types"
	legacypacket_v582 "github.com/flonja/multiversion/protocols/v582/packet"
	legacypacket_v589 "github.com/flonja/multiversion/protocols/v589/packet"
	types_v589 "github.com/flonja/multiversion/protocols/v589/types"
	"github.com/flonja/multiversion/translator"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte
	//go:embed block_states.nbt
	blockStateData []byte
)

// Protocol Deprecated due to Mojang not supporting <1.20
type Protocol struct {
	itemMapping     mapping.Item
	blockMapping    mapping.Block
	itemTranslator  translator.ItemTranslator
	blockTranslator translator.BlockTranslator
}

func New() *Protocol {
	// TODOn't: add custom block/item replacements (aka make it cool)

	itemMapping := mapping.NewItemMapping(itemRuntimeIDData)
	blockMapping := mapping.NewBlockMapping(blockStateData).WithBlockActorRemapper(downgradeBlockActorData, upgradeBlockActorData)
	latestBlockMapping := latest.NewBlockMapping()
	return &Protocol{itemMapping: itemMapping, blockMapping: blockMapping,
		itemTranslator:  translator.NewItemTranslator(itemMapping, latest.NewItemMapping(), blockMapping, latestBlockMapping),
		blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ID() int32 {
	return 486
}

func (p Protocol) Ver() string {
	return "1.18.12"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDAddActor] = func() packet.Packet { return &legacypacket.AddActor{} }
	pool[packet.IDAddPlayer] = func() packet.Packet { return &legacypacket.AddPlayer{} }
	pool[packet.IDAddVolumeEntity] = func() packet.Packet { return &legacypacket.AddVolumeEntity{} }
	pool[packet.IDAvailableCommands] = func() packet.Packet { return &legacypacket_v589.AvailableCommands{} }
	pool[packet.IDEmote] = func() packet.Packet { return &legacypacket_v582.Emote{} }
	pool[packet.IDCommandRequest] = func() packet.Packet { return &legacypacket.CommandRequest{} }
	pool[packet.IDNetworkChunkPublisherUpdate] = func() packet.Packet { return &legacypacket.NetworkChunkPublisherUpdate{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDPlayerList] = func() packet.Packet { return &legacypacket.PlayerList{} }
	pool[packet.IDPlayerSkin] = func() packet.Packet { return &legacypacket.PlayerSkin{} }
	pool[packet.IDRemoveVolumeEntity] = func() packet.Packet { return &legacypacket.RemoveVolumeEntity{} }
	pool[packet.IDRequestChunkRadius] = func() packet.Packet { return &legacypacket.RequestChunkRadius{} }
	pool[packet.IDSpawnParticleEffect] = func() packet.Packet { return &legacypacket.SpawnParticleEffect{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDStructureBlockUpdate] = func() packet.Packet { return &legacypacket.StructureBlockUpdate{} }
	pool[packet.IDStructureTemplateDataRequest] = func() packet.Packet { return &legacypacket.StructureTemplateDataRequest{} }
	pool[packet.IDUpdateAttributes] = func() packet.Packet { return &legacypacket.UpdateAttributes{} }
	pool[packet.IDItemStackRequest] = func() packet.Packet { return &legacypacket.ItemStackRequest{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	return pool
}

func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return NewReader(protocol.NewReader(r, shieldID, enableLimits))
}

func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return NewWriter(protocol.NewWriter(w, shieldID))
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	var newPks []packet.Packet
	switch pk := pk.(type) {
	case *packet.ClientCacheStatus:
		pk.Enabled = false
		newPks = append(newPks, pk)
	case *legacypacket.AddActor:
		newPks = append(newPks, &packet.AddActor{
			EntityMetadata:   upgradeEntityMetadata(pk.EntityMetadata),
			EntityRuntimeID:  pk.EntityRuntimeID,
			EntityType:       pk.EntityType,
			EntityUniqueID:   pk.EntityUniqueID,
			HeadYaw:          pk.HeadYaw,
			Pitch:            pk.Pitch,
			Position:         pk.Position,
			Velocity:         pk.Velocity,
			Yaw:              pk.Yaw,
			Attributes:       pk.Attributes,
			EntityLinks:      pk.EntityLinks,
			EntityProperties: protocol.EntityProperties{},
		})
	case *legacypacket.AddPlayer:
		newPks = append(newPks, &packet.AddPlayer{
			UUID:             pk.UUID,
			Username:         pk.Username,
			EntityRuntimeID:  pk.EntityRuntimeID,
			PlatformChatID:   pk.PlatformChatID,
			Position:         pk.Position,
			Velocity:         pk.Velocity,
			Pitch:            pk.Pitch,
			Yaw:              pk.Yaw,
			HeadYaw:          pk.HeadYaw,
			HeldItem:         pk.HeldItem,
			EntityMetadata:   upgradeEntityMetadata(pk.EntityMetadata),
			DeviceID:         pk.DeviceID,
			EntityLinks:      pk.EntityLinks,
			GameType:         packet.GameTypeSurvival,
			EntityProperties: protocol.EntityProperties{},
			AbilityData: protocol.AbilityData{
				EntityUniqueID:     pk.EntityUniqueID,
				PlayerPermissions:  byte(pk.AdventureSettings.PermissionLevel),
				CommandPermissions: byte(pk.AdventureSettings.CommandPermissionLevel),
				Layers: []protocol.AbilityLayer{{
					Type:      protocol.AbilityLayerTypeBase,
					Abilities: protocol.AbilityCount - 1,
				}},
			},
			BuildPlatform: int32(protocol.DeviceAndroid),
		})
	case *legacypacket.AddVolumeEntity:
		newPks = append(newPks, &packet.AddVolumeEntity{
			EntityRuntimeID:    pk.EntityRuntimeID,
			EntityMetadata:     pk.EntityMetadata,
			EncodingIdentifier: pk.EncodingIdentifier,
			InstanceIdentifier: pk.InstanceIdentifier,
			EngineVersion:      pk.EngineVersion,
			Bounds:             [2]protocol.BlockPos{},
			Dimension:          0,
		})
	case *legacypacket_v589.AvailableCommands:
		for ind1, command := range pk.Commands {
			for ind2, overload := range command.Overloads {
				for ind3, parameter := range overload.Parameters {
					parameterType := uint32(0)
					switch parameter.Type {
					case 7:
						parameterType = protocol.CommandArgTypeTarget
					case 8:
						parameterType = protocol.CommandArgTypeWildcardTarget
					case 16:
						parameterType = protocol.CommandArgTypeFilepath
					case 32:
						parameterType = protocol.CommandArgTypeString
					case 40:
						parameterType = protocol.CommandArgTypePosition
					case 44:
						parameterType = protocol.CommandArgTypeMessage
					case 46:
						parameterType = protocol.CommandArgTypeRawText
					case 50:
						parameterType = protocol.CommandArgTypeJSON
					case 63:
						parameterType = protocol.CommandArgTypeCommand
					}
					parameter.Type = parameterType
					pk.Commands[ind1].Overloads[ind2].Parameters[ind3] = parameter
				}
			}
		}
		newPks = append(newPks, &packet.AvailableCommands{
			EnumValues: pk.EnumValues,
			Suffixes:   pk.Suffixes,
			Enums:      pk.Enums,
			Commands: lo.Map(pk.Commands, func(item types_v589.Command, _ int) protocol.Command {
				return protocol.Command{
					Name:            item.Name,
					Description:     item.Description,
					Flags:           item.Flags,
					PermissionLevel: item.PermissionLevel,
					AliasesOffset:   item.AliasesOffset,
					Overloads: lo.Map(item.Overloads, func(item types_v589.CommandOverload, _ int) protocol.CommandOverload {
						return protocol.CommandOverload{
							Parameters: item.Parameters,
							Chaining:   false,
						}
					}),
				}
			}),
			DynamicEnums: pk.DynamicEnums,
			Constraints:  pk.Constraints,
		})
	case *packet.BlockActorData:
		pk.NBTData = downgradeBlockActorData(pk.NBTData)
		newPks = append(newPks, pk)
	case *legacypacket.CommandRequest:
		newPks = append(newPks, &packet.CommandRequest{
			CommandLine:   pk.CommandLine,
			CommandOrigin: pk.CommandOrigin,
			Internal:      pk.Internal,
		})
	case *packet.CraftingData:
		for i, recipe := range pk.Recipes {
			switch recipe := recipe.(type) {
			case *protocol.ShapelessRecipe:
				recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
					item.Descriptor = upgradeCraftingDescription(item.Descriptor.(*types.DefaultItemDescriptor))
					return item
				})
				pk.Recipes[i] = recipe
			case *protocol.ShapedRecipe:
				recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
					item.Descriptor = upgradeCraftingDescription(item.Descriptor.(*types.DefaultItemDescriptor))
					return item
				})
				pk.Recipes[i] = recipe
			case *protocol.ShulkerBoxRecipe:
				recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
					item.Descriptor = upgradeCraftingDescription(item.Descriptor.(*types.DefaultItemDescriptor))
					return item
				})
				pk.Recipes[i] = recipe
			case *protocol.ShapelessChemistryRecipe:
				recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
					item.Descriptor = upgradeCraftingDescription(item.Descriptor.(*types.DefaultItemDescriptor))
					return item
				})
				pk.Recipes[i] = recipe
			case *protocol.ShapedChemistryRecipe:
				recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
					item.Descriptor = upgradeCraftingDescription(item.Descriptor.(*types.DefaultItemDescriptor))
					return item
				})
				pk.Recipes[i] = recipe
			case *protocol.SmithingTransformRecipe:
				recipe.Template.Descriptor = upgradeCraftingDescription(recipe.Template.Descriptor.(*types.DefaultItemDescriptor))
				recipe.Base.Descriptor = upgradeCraftingDescription(recipe.Base.Descriptor.(*types.DefaultItemDescriptor))
				recipe.Addition.Descriptor = upgradeCraftingDescription(recipe.Addition.Descriptor.(*types.DefaultItemDescriptor))
				pk.Recipes[i] = recipe
			case *protocol.SmithingTrimRecipe:
				recipe.Template.Descriptor = upgradeCraftingDescription(recipe.Template.Descriptor.(*types.DefaultItemDescriptor))
				recipe.Base.Descriptor = upgradeCraftingDescription(recipe.Base.Descriptor.(*types.DefaultItemDescriptor))
				recipe.Addition.Descriptor = upgradeCraftingDescription(recipe.Addition.Descriptor.(*types.DefaultItemDescriptor))
			}
		}
		newPks = append(newPks, pk)
	case *packet.InventoryTransaction:
		pk.LegacySetItemSlots = lo.Map(pk.LegacySetItemSlots, func(item protocol.LegacySetItemSlot, _ int) protocol.LegacySetItemSlot {
			if item.ContainerID >= 21 { // RECIPE_BOOK
				item.ContainerID += 1
			}
			return item
		})
		newPks = append(newPks, pk)
	case *packet.ItemStackResponse:
		for i2, respons := range pk.Responses {
			for i3, info := range respons.ContainerInfo {
				if info.ContainerID >= 21 { // RECIPE_BOOK
					info.ContainerID += 1
				}
				pk.Responses[i2].ContainerInfo[i3] = info
			}
		}
		newPks = append(newPks, pk)
	case *legacypacket.ItemStackRequest:
		newPks = append(newPks, &packet.ItemStackRequest{
			Requests: lo.Map(pk.Requests, func(item types.ItemStackRequest, _ int) protocol.ItemStackRequest {
				return protocol.ItemStackRequest{
					RequestID: item.RequestID,
					Actions: lo.Map(item.Actions, func(item protocol.StackRequestAction, _ int) protocol.StackRequestAction {
						switch action := item.(type) {
						case *types.TakeStackRequestAction:
							return &action.TakeStackRequestAction
						case *types.PlaceStackRequestAction:
							return &action.PlaceStackRequestAction
						case *types.SwapStackRequestAction:
							return &action.SwapStackRequestAction
						case *types.DropStackRequestAction:
							return &action.DropStackRequestAction
						case *types.DestroyStackRequestAction:
							return &action.DestroyStackRequestAction
						case *types.ConsumeStackRequestAction:
							return &action.DestroyStackRequestAction
						case *types.PlaceInContainerStackRequestAction:
							return &action.PlaceInContainerStackRequestAction
						case *types.TakeOutContainerStackRequestAction:
							return &action.TakeOutContainerStackRequestAction
						case *types.AutoCraftRecipeStackRequestAction:
							return &action.AutoCraftRecipeStackRequestAction
						}
						return item
					}),
					FilterStrings: item.FilterStrings,
				}
			}),
		})
	case *legacypacket.ModalFormResponse:
		responseData := protocol.Optional[[]byte]{}
		cancelReason := protocol.Optional[uint8]{}
		if string(pk.ResponseData) == "null" {
			var cancelReasonType uint8 = packet.ModalFormCancelReasonUserClosed
			cancelReason = protocol.Option(cancelReasonType)
		} else {
			responseData = protocol.Option(pk.ResponseData)
		}
		newPks = append(newPks, &packet.ModalFormResponse{
			FormID:       pk.FormID,
			ResponseData: responseData,
			CancelReason: cancelReason,
		})
	case *legacypacket.NetworkChunkPublisherUpdate:
		newPks = append(newPks, &packet.NetworkChunkPublisherUpdate{
			Position:    pk.Position,
			Radius:      pk.Radius,
			SavedChunks: []protocol.ChunkPos{},
		})
	case *legacypacket.PlayerAction:
		newPks = append(newPks, &packet.PlayerAction{
			EntityRuntimeID: pk.EntityRuntimeID,
			ActionType:      pk.ActionType,
			BlockPosition:   pk.BlockPosition,
			ResultPosition:  pk.BlockPosition,
			BlockFace:       pk.BlockFace,
		})
	case *legacypacket.PlayerAuthInput:
		newPks = append(newPks, &packet.PlayerAuthInput{
			Pitch:         pk.Pitch,
			Yaw:           pk.Yaw,
			Position:      pk.Position,
			MoveVector:    pk.MoveVector,
			HeadYaw:       pk.HeadYaw,
			InputData:     pk.InputData,
			InputMode:     pk.InputMode,
			PlayMode:      pk.PlayMode,
			GazeDirection: pk.GazeDirection,
			Tick:          pk.Tick,
			Delta:         pk.Delta,
			ItemInteractionData: func(data protocol.UseItemTransactionData) protocol.UseItemTransactionData {
				data.LegacySetItemSlots = lo.Map(data.LegacySetItemSlots, func(item protocol.LegacySetItemSlot, _ int) protocol.LegacySetItemSlot {
					if item.ContainerID >= 21 { // RECIPE_BOOK
						item.ContainerID += 1
					}
					return item
				})
				return data
			}(pk.ItemInteractionData),
			ItemStackRequest: protocol.ItemStackRequest{
				RequestID: pk.ItemStackRequest.RequestID,
				Actions: lo.Map(pk.ItemStackRequest.Actions, func(item protocol.StackRequestAction, _ int) protocol.StackRequestAction {
					switch action := item.(type) {
					case *types.TakeStackRequestAction:
						return &action.TakeStackRequestAction
					case *types.PlaceStackRequestAction:
						return &action.PlaceStackRequestAction
					case *types.SwapStackRequestAction:
						return &action.SwapStackRequestAction
					case *types.DropStackRequestAction:
						return &action.DropStackRequestAction
					case *types.DestroyStackRequestAction:
						return &action.DestroyStackRequestAction
					case *types.ConsumeStackRequestAction:
						return &action.DestroyStackRequestAction
					case *types.PlaceInContainerStackRequestAction:
						return &action.PlaceInContainerStackRequestAction
					case *types.TakeOutContainerStackRequestAction:
						return &action.TakeOutContainerStackRequestAction
					case *types.AutoCraftRecipeStackRequestAction:
						return &action.AutoCraftRecipeStackRequestAction
					}
					return item
				}),
				FilterStrings: pk.ItemStackRequest.FilterStrings,
			},
			BlockActions:       pk.BlockActions,
			AnalogueMoveVector: pk.MoveVector,
		})
	case *legacypacket.PlayerList:
		newPks = append(newPks, &packet.PlayerList{
			ActionType: pk.ActionType,
			Entries: lo.Map(pk.Entries, func(item types.PlayerListEntry, _ int) protocol.PlayerListEntry {
				return item.PlayerListEntry
			}),
		})
	case *legacypacket.PlayerSkin:
		newPks = append(newPks, &packet.PlayerSkin{
			UUID:        pk.UUID,
			Skin:        pk.Skin.Skin,
			NewSkinName: pk.NewSkinName,
			OldSkinName: pk.OldSkinName,
		})
	case *legacypacket.RemoveVolumeEntity:
		newPks = append(newPks, &packet.RemoveVolumeEntity{
			EntityRuntimeID: pk.EntityRuntimeID,
			Dimension:       0,
		})
	case *legacypacket.RequestChunkRadius:
		newPks = append(newPks, &packet.RequestChunkRadius{
			ChunkRadius:    pk.ChunkRadius,
			MaxChunkRadius: pk.ChunkRadius,
		})
	case *legacypacket.SetActorData:
		newPks = append(newPks, &packet.SetActorData{
			EntityRuntimeID:  pk.EntityRuntimeID,
			EntityMetadata:   upgradeEntityMetadata(pk.EntityMetadata),
			EntityProperties: protocol.EntityProperties{},
			Tick:             pk.Tick,
		})
	case *legacypacket.SpawnParticleEffect:
		newPks = append(newPks, &packet.SpawnParticleEffect{
			Dimension:       pk.Dimension,
			EntityUniqueID:  pk.EntityUniqueID,
			Position:        pk.Position,
			ParticleName:    pk.ParticleName,
			MoLangVariables: protocol.Optional[[]byte]{},
		})
	case *legacypacket.StartGame:
		newPks = append(newPks, &packet.StartGame{
			EntityUniqueID:                 pk.EntityUniqueID,
			EntityRuntimeID:                pk.EntityRuntimeID,
			PlayerGameMode:                 pk.PlayerGameMode,
			PlayerPosition:                 pk.PlayerPosition,
			Pitch:                          pk.Pitch,
			Yaw:                            pk.Yaw,
			WorldSeed:                      int64(pk.WorldSeed),
			SpawnBiomeType:                 pk.SpawnBiomeType,
			UserDefinedBiomeName:           pk.UserDefinedBiomeName,
			Dimension:                      pk.Dimension,
			Generator:                      pk.Generator,
			WorldGameMode:                  pk.WorldGameMode,
			Difficulty:                     pk.Difficulty,
			WorldSpawn:                     pk.WorldSpawn,
			AchievementsDisabled:           pk.AchievementsDisabled,
			DayCycleLockTime:               pk.DayCycleLockTime,
			EducationEditionOffer:          pk.EducationEditionOffer,
			EducationFeaturesEnabled:       pk.EducationFeaturesEnabled,
			EducationProductID:             pk.EducationProductID,
			RainLevel:                      pk.RainLevel,
			LightningLevel:                 pk.LightningLevel,
			ConfirmedPlatformLockedContent: pk.ConfirmedPlatformLockedContent,
			MultiPlayerGame:                pk.MultiPlayerGame,
			LANBroadcastEnabled:            pk.LANBroadcastEnabled,
			XBLBroadcastMode:               pk.XBLBroadcastMode,
			PlatformBroadcastMode:          pk.PlatformBroadcastMode,
			CommandsEnabled:                pk.CommandsEnabled,
			TexturePackRequired:            pk.TexturePackRequired,
			GameRules:                      pk.GameRules,
			Experiments:                    pk.Experiments,
			ExperimentsPreviouslyToggled:   pk.ExperimentsPreviouslyToggled,
			BonusChestEnabled:              pk.BonusChestEnabled,
			StartWithMapEnabled:            pk.StartWithMapEnabled,
			PlayerPermissions:              pk.PlayerPermissions,
			ServerChunkTickRadius:          pk.ServerChunkTickRadius,
			HasLockedBehaviourPack:         pk.HasLockedBehaviourPack,
			HasLockedTexturePack:           pk.HasLockedTexturePack,
			FromLockedWorldTemplate:        pk.FromLockedWorldTemplate,
			MSAGamerTagsOnly:               pk.MSAGamerTagsOnly,
			FromWorldTemplate:              pk.FromWorldTemplate,
			WorldTemplateSettingsLocked:    pk.WorldTemplateSettingsLocked,
			OnlySpawnV1Villagers:           pk.OnlySpawnV1Villagers,
			BaseGameVersion:                pk.BaseGameVersion,
			LimitedWorldWidth:              pk.LimitedWorldWidth,
			LimitedWorldDepth:              pk.LimitedWorldDepth,
			NewNether:                      pk.NewNether,
			EducationSharedResourceURI:     pk.EducationSharedResourceURI,
			ForceExperimentalGameplay:      protocol.Option(pk.ForceExperimentalGameplay),
			LevelID:                        pk.LevelID,
			WorldName:                      pk.WorldName,
			TemplateContentIdentity:        pk.TemplateContentIdentity,
			Trial:                          pk.Trial,
			PlayerMovementSettings:         pk.PlayerMovementSettings,
			Time:                           pk.Time,
			EnchantmentSeed:                pk.EnchantmentSeed,
			Blocks:                         pk.Blocks,
			Items:                          pk.Items,
			MultiPlayerCorrelationID:       pk.MultiPlayerCorrelationID,
			ServerAuthoritativeInventory:   pk.ServerAuthoritativeInventory,
			GameVersion:                    pk.GameVersion,
			ServerBlockStateChecksum:       pk.ServerBlockStateChecksum,
		})
	case *legacypacket.StructureBlockUpdate:
		newPks = append(newPks, &packet.StructureBlockUpdate{
			Position:           pk.Position,
			StructureName:      pk.StructureName,
			DataField:          pk.DataField,
			IncludePlayers:     pk.IncludePlayers,
			ShowBoundingBox:    pk.ShowBoundingBox,
			StructureBlockType: pk.StructureBlockType,
			Settings:           pk.Settings.StructureSettings,
			RedstoneSaveMode:   pk.RedstoneSaveMode,
			ShouldTrigger:      pk.ShouldTrigger,
			Waterlogged:        pk.Waterlogged,
		})
	case *legacypacket.StructureTemplateDataRequest:
		newPks = append(newPks, &packet.StructureTemplateDataRequest{
			StructureName: pk.StructureName,
			Position:      pk.Position,
			Settings:      pk.Settings.StructureSettings,
			RequestType:   pk.RequestType,
		})
	case *packet.SetActorData:
		pk.EntityMetadata = upgradeEntityMetadata(pk.EntityMetadata)
		newPks = append(newPks, pk)
	case *legacypacket.UpdateAttributes:
		newPks = append(newPks, &packet.UpdateAttributes{
			EntityRuntimeID: pk.EntityRuntimeID,
			Attributes: lo.Map(pk.Attributes, func(item types.Attribute, _ int) protocol.Attribute {
				return item.Attribute
			}),
			Tick: pk.Tick,
		})
	case *packet.AdventureSettings:
		handleFlag := func(flags uint32, secondFlag bool) uint32 {
			layerMapping := map[uint32]uint32{
				packet.AdventureFlagAllowFlight:  protocol.AbilityMayFly,
				packet.AdventureFlagNoClip:       protocol.AbilityNoClip,
				packet.AdventureFlagWorldBuilder: protocol.AbilityWorldBuilder,
				packet.AdventureFlagFlying:       protocol.AbilityFlying,
				packet.AdventureFlagMuted:        protocol.AbilityMuted,
			}
			if secondFlag {
				layerMapping = map[uint32]uint32{
					packet.ActionPermissionMine:             protocol.AbilityMine,
					packet.ActionPermissionDoorsAndSwitches: protocol.AbilityDoorsAndSwitches,
					packet.ActionPermissionOpenContainers:   protocol.AbilityOpenContainers,
					packet.ActionPermissionAttackPlayers:    protocol.AbilityAttackPlayers,
					packet.ActionPermissionAttackMobs:       protocol.AbilityAttackMobs,
					packet.ActionPermissionOperator:         protocol.AbilityOperatorCommands,
					packet.ActionPermissionBuild:            protocol.AbilityBuild,
				}
			}

			out := uint32(0)
			for flag, mapped := range layerMapping {
				if (flags & flag) != 0 {
					out |= mapped
				}
			}
			return out
		}

		_ = handleFlag
		//newPks = append(newPks, &packet.UpdateAbilities{
		//	AbilityData: protocol.AbilityData{
		//		EntityUniqueID:     pk.PlayerUniqueID,
		//		PlayerPermissions:  byte(pk.PermissionLevel),
		//		CommandPermissions: byte(pk.CommandPermissionLevel),
		//		Layers: []protocol.AbilityLayer{
		//			{
		//				Type:      protocol.AbilityLayerTypeBase,
		//				Abilities: protocol.AbilityCount - 1,
		//				Values:    handleFlag(pk.Flags, false) | handleFlag(pk.ActionPermissions, true),
		//				FlySpeed:  protocol.AbilityBaseFlySpeed,
		//				WalkSpeed: protocol.AbilityBaseWalkSpeed,
		//			},
		//		},
		//	},
		//})
	case *legacypacket_v582.Emote:
		newPks = append(newPks, &packet.Emote{
			EntityRuntimeID: pk.EntityRuntimeID,
			EmoteID:         pk.EmoteID,
			XUID:            conn.IdentityData().XUID,
			PlatformID:      conn.ClientData().PlatformOnlineID,
			Flags:           pk.Flags,
		})
	default:
		newPks = append(newPks, pk)
	}

	return p.blockTranslator.UpgradeBlockPackets(p.itemTranslator.UpgradeItemPackets(newPks, conn), conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets(p.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn), conn)

	for i, pk := range result {
		switch pk := pk.(type) {
		case *packet.AddActor:
			result[i] = &legacypacket.AddActor{
				EntityMetadata:  downgradeEntityMetadata(pk.EntityMetadata),
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityType:      pk.EntityType,
				EntityUniqueID:  pk.EntityUniqueID,
				HeadYaw:         pk.HeadYaw,
				Pitch:           pk.Pitch,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Yaw:             pk.Yaw,
				Attributes:      pk.Attributes,
				EntityLinks:     pk.EntityLinks,
			}
		case *packet.AddPlayer:
			result[i] = &legacypacket.AddPlayer{
				UUID:            pk.UUID,
				Username:        pk.Username,
				EntityUniqueID:  pk.AbilityData.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				PlatformChatID:  pk.PlatformChatID,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Pitch:           pk.Pitch,
				Yaw:             pk.Yaw,
				HeadYaw:         pk.HeadYaw,
				HeldItem:        pk.HeldItem,
				EntityMetadata:  downgradeEntityMetadata(pk.EntityMetadata),
				AdventureSettings: packet.AdventureSettings{
					CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
					PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
					PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
				},
				DeviceID:    pk.DeviceID,
				EntityLinks: pk.EntityLinks,
			}
		case *packet.AddVolumeEntity:
			result[i] = &legacypacket.AddVolumeEntity{
				EntityRuntimeID:    pk.EntityRuntimeID,
				EntityMetadata:     pk.EntityMetadata,
				EncodingIdentifier: pk.EncodingIdentifier,
				InstanceIdentifier: pk.InstanceIdentifier,
				EngineVersion:      pk.EngineVersion,
			}
		case *packet.AvailableCommands:
			for ind1, command := range pk.Commands {
				for ind2, overload := range command.Overloads {
					for ind3, parameter := range overload.Parameters {
						parameterType := uint32(parameter.Type) | protocol.CommandArgValid

						switch parameter.Type | protocol.CommandArgValid {
						case protocol.CommandArgTypeCompareOperator:
							parameterType = protocol.CommandArgTypeOperator
						case protocol.CommandArgTypeTarget:
							parameterType = 7
						case protocol.CommandArgTypeWildcardTarget:
							parameterType = 8
						case protocol.CommandArgTypeFilepath:
							parameterType = 16
						case protocol.CommandArgTypeString:
							parameterType = 32
						case protocol.CommandArgTypeBlockPosition:
							fallthrough
						case protocol.CommandArgTypePosition:
							parameterType = 40
						case protocol.CommandArgTypeMessage:
							parameterType = 44
						case protocol.CommandArgTypeRawText:
							parameterType = 46
						case protocol.CommandArgTypeJSON:
							parameterType = 50
						case protocol.CommandArgTypeCommand:
							parameterType = 63
						}
						parameter.Type = parameterType | protocol.CommandArgValid
						pk.Commands[ind1].Overloads[ind2].Parameters[ind3] = parameter
					}
				}
			}
			result[i] = &legacypacket_v589.AvailableCommands{
				EnumValues: pk.EnumValues,
				Suffixes:   pk.Suffixes,
				Enums:      pk.Enums,
				Commands: lo.Map(pk.Commands, func(item protocol.Command, _ int) types_v589.Command {
					return types_v589.Command{
						Name:            item.Name,
						Description:     item.Description,
						Flags:           item.Flags,
						PermissionLevel: item.PermissionLevel,
						AliasesOffset:   item.AliasesOffset,
						Overloads: lo.Map(item.Overloads, func(item protocol.CommandOverload, _ int) types_v589.CommandOverload {
							return types_v589.CommandOverload{
								Parameters: item.Parameters,
							}
						}),
					}
				}),
				DynamicEnums: pk.DynamicEnums,
				Constraints:  pk.Constraints,
			}
		case *packet.BlockActorData:
			pk.NBTData = downgradeBlockActorData(pk.NBTData)
			result[i] = pk
		case *packet.CommandRequest:
			result[i] = &legacypacket.CommandRequest{
				CommandLine:   pk.CommandLine,
				CommandOrigin: pk.CommandOrigin,
				Internal:      pk.Internal,
			}
		case *packet.CraftingData:
			for i, recipe := range pk.Recipes {
				switch recipe := recipe.(type) {
				case *protocol.ShapelessRecipe:
					recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
						item.Descriptor = downgradeCraftingDescription(item.Descriptor, p.itemMapping)
						return item
					})
					pk.Recipes[i] = recipe
				case *protocol.ShapedRecipe:
					recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
						item.Descriptor = downgradeCraftingDescription(item.Descriptor, p.itemMapping)
						return item
					})
					pk.Recipes[i] = recipe
				case *protocol.ShulkerBoxRecipe:
					recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
						item.Descriptor = downgradeCraftingDescription(item.Descriptor, p.itemMapping)
						return item
					})
					pk.Recipes[i] = recipe
				case *protocol.ShapelessChemistryRecipe:
					recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
						item.Descriptor = downgradeCraftingDescription(item.Descriptor, p.itemMapping)
						return item
					})
					pk.Recipes[i] = recipe
				case *protocol.ShapedChemistryRecipe:
					recipe.Input = lo.Map(recipe.Input, func(item protocol.ItemDescriptorCount, _ int) protocol.ItemDescriptorCount {
						item.Descriptor = downgradeCraftingDescription(item.Descriptor, p.itemMapping)
						return item
					})
					pk.Recipes[i] = recipe
				case *protocol.SmithingTransformRecipe:
					recipe.Template.Descriptor = downgradeCraftingDescription(recipe.Template.Descriptor, p.itemMapping)
					recipe.Base.Descriptor = downgradeCraftingDescription(recipe.Base.Descriptor, p.itemMapping)
					recipe.Addition.Descriptor = downgradeCraftingDescription(recipe.Addition.Descriptor, p.itemMapping)
					pk.Recipes[i] = recipe
				case *protocol.SmithingTrimRecipe:
					recipe.Template.Descriptor = downgradeCraftingDescription(recipe.Template.Descriptor, p.itemMapping)
					recipe.Base.Descriptor = downgradeCraftingDescription(recipe.Base.Descriptor, p.itemMapping)
					recipe.Addition.Descriptor = downgradeCraftingDescription(recipe.Addition.Descriptor, p.itemMapping)
				}
			}
			result[i] = pk
		case *packet.InventoryTransaction:
			pk.LegacySetItemSlots = lo.Map(pk.LegacySetItemSlots, func(item protocol.LegacySetItemSlot, _ int) protocol.LegacySetItemSlot {
				if item.ContainerID > 21 { // RECIPE_BOOK
					item.ContainerID -= 1
				}
				return item
			})
			result[i] = pk
		case *packet.ItemStackResponse:
			for i2, respons := range pk.Responses {
				for i3, info := range respons.ContainerInfo {
					if info.ContainerID > 21 { // RECIPE_BOOK
						info.ContainerID -= 1
					}
					pk.Responses[i2].ContainerInfo[i3] = info
				}
			}
			result[i] = pk
		case *packet.ItemStackRequest:
			result[i] = &legacypacket.ItemStackRequest{
				Requests: lo.Map(pk.Requests, func(item protocol.ItemStackRequest, _ int) types.ItemStackRequest {
					item.Actions = lo.Map(item.Actions, func(item protocol.StackRequestAction, _ int) protocol.StackRequestAction {
						switch action := item.(type) {
						case *protocol.TakeStackRequestAction:
							return &types.TakeStackRequestAction{TakeStackRequestAction: *action}
						case *protocol.PlaceStackRequestAction:
							return &types.PlaceStackRequestAction{PlaceStackRequestAction: *action}
						case *protocol.SwapStackRequestAction:
							return &types.SwapStackRequestAction{SwapStackRequestAction: *action}
						case *protocol.DropStackRequestAction:
							return &types.DropStackRequestAction{DropStackRequestAction: *action}
						case *protocol.DestroyStackRequestAction:
							return &types.DestroyStackRequestAction{DestroyStackRequestAction: *action}
						case *protocol.ConsumeStackRequestAction:
							return &types.ConsumeStackRequestAction{DestroyStackRequestAction: action.DestroyStackRequestAction}
						case *protocol.PlaceInContainerStackRequestAction:
							return &types.PlaceInContainerStackRequestAction{PlaceInContainerStackRequestAction: *action}
						case *protocol.TakeOutContainerStackRequestAction:
							return &types.TakeOutContainerStackRequestAction{TakeOutContainerStackRequestAction: *action}
						case *protocol.AutoCraftRecipeStackRequestAction:
							return &types.AutoCraftRecipeStackRequestAction{AutoCraftRecipeStackRequestAction: *action}
						}
						return item
					})
					return types.ItemStackRequest{ItemStackRequest: item}
				}),
			}
		case *packet.ModalFormResponse:
			var responseData []byte
			if val, ok := pk.ResponseData.Value(); ok {
				responseData = val
			}
			if _, cancelled := pk.CancelReason.Value(); cancelled {
				if resp, err := json.Marshal(nil); err == nil {
					responseData = resp
				}
			}
			result[i] = &legacypacket.ModalFormResponse{
				FormID:       pk.FormID,
				ResponseData: responseData,
			}
		case *packet.NetworkChunkPublisherUpdate:
			result[i] = &legacypacket.NetworkChunkPublisherUpdate{
				Position: pk.Position,
				Radius:   pk.Radius,
			}
		case *packet.PlayerAction:
			result[i] = &legacypacket.PlayerAction{
				EntityRuntimeID: pk.EntityRuntimeID,
				ActionType:      pk.ActionType,
				BlockPosition:   pk.BlockPosition,
				BlockFace:       pk.BlockFace,
			}
		case *packet.PlayerAuthInput:
			result[i] = &legacypacket.PlayerAuthInput{
				Pitch:         pk.Pitch,
				Yaw:           pk.Yaw,
				Position:      pk.Position,
				MoveVector:    pk.MoveVector,
				HeadYaw:       pk.HeadYaw,
				InputData:     pk.InputData,
				InputMode:     pk.InputMode,
				PlayMode:      pk.PlayMode,
				GazeDirection: pk.GazeDirection,
				Tick:          pk.Tick,
				Delta:         pk.Delta,
				ItemInteractionData: func(data protocol.UseItemTransactionData) protocol.UseItemTransactionData {
					data.LegacySetItemSlots = lo.Map(data.LegacySetItemSlots, func(item protocol.LegacySetItemSlot, _ int) protocol.LegacySetItemSlot {
						if item.ContainerID > 21 { // RECIPE_BOOK
							item.ContainerID -= 1
						}
						return item
					})
					return data
				}(pk.ItemInteractionData),
				ItemStackRequest: types.ItemStackRequest{ItemStackRequest: pk.ItemStackRequest},
				BlockActions:     pk.BlockActions,
			}
		case *packet.PlayerList:
			result[i] = &legacypacket.PlayerList{
				ActionType: pk.ActionType,
				Entries: lo.Map(pk.Entries, func(item protocol.PlayerListEntry, _ int) types.PlayerListEntry {
					return types.PlayerListEntry{PlayerListEntry: item}
				}),
			}
		case *packet.PlayerSkin:
			result[i] = &legacypacket.PlayerSkin{
				UUID:        pk.UUID,
				Skin:        types.Skin{Skin: pk.Skin},
				NewSkinName: pk.NewSkinName,
				OldSkinName: pk.OldSkinName,
			}
		case *packet.RemoveVolumeEntity:
			result[i] = &legacypacket.RemoveVolumeEntity{
				EntityRuntimeID: pk.EntityRuntimeID,
			}
		case *packet.RequestChunkRadius:
			result[i] = &legacypacket.RequestChunkRadius{
				ChunkRadius: pk.ChunkRadius,
			}
		case *packet.SetActorData:
			result[i] = &legacypacket.SetActorData{
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityMetadata:  downgradeEntityMetadata(pk.EntityMetadata),
				Tick:            pk.Tick,
			}
		case *packet.SpawnParticleEffect:
			result[i] = &legacypacket.SpawnParticleEffect{
				Dimension:      pk.Dimension,
				EntityUniqueID: pk.EntityUniqueID,
				Position:       pk.Position,
				ParticleName:   pk.ParticleName,
			}
		case *packet.StartGame:
			_, enabled := pk.ForceExperimentalGameplay.Value()
			result[i] = &legacypacket.StartGame{
				EntityUniqueID:                 pk.EntityUniqueID,
				EntityRuntimeID:                pk.EntityRuntimeID,
				PlayerGameMode:                 pk.PlayerGameMode,
				PlayerPosition:                 pk.PlayerPosition,
				Pitch:                          pk.Pitch,
				Yaw:                            pk.Yaw,
				WorldSeed:                      int32(pk.WorldSeed),
				SpawnBiomeType:                 pk.SpawnBiomeType,
				UserDefinedBiomeName:           pk.UserDefinedBiomeName,
				Dimension:                      pk.Dimension,
				Generator:                      pk.Generator,
				WorldGameMode:                  pk.WorldGameMode,
				Difficulty:                     pk.Difficulty,
				WorldSpawn:                     pk.WorldSpawn,
				AchievementsDisabled:           pk.AchievementsDisabled,
				DayCycleLockTime:               pk.DayCycleLockTime,
				EducationEditionOffer:          pk.EducationEditionOffer,
				EducationFeaturesEnabled:       pk.EducationFeaturesEnabled,
				EducationProductID:             pk.EducationProductID,
				RainLevel:                      pk.RainLevel,
				LightningLevel:                 pk.LightningLevel,
				ConfirmedPlatformLockedContent: pk.ConfirmedPlatformLockedContent,
				MultiPlayerGame:                pk.MultiPlayerGame,
				LANBroadcastEnabled:            pk.LANBroadcastEnabled,
				XBLBroadcastMode:               pk.XBLBroadcastMode,
				PlatformBroadcastMode:          pk.PlatformBroadcastMode,
				CommandsEnabled:                pk.CommandsEnabled,
				TexturePackRequired:            pk.TexturePackRequired,
				GameRules:                      pk.GameRules,
				Experiments:                    pk.Experiments,
				ExperimentsPreviouslyToggled:   pk.ExperimentsPreviouslyToggled,
				BonusChestEnabled:              pk.BonusChestEnabled,
				StartWithMapEnabled:            pk.StartWithMapEnabled,
				PlayerPermissions:              pk.PlayerPermissions,
				ServerChunkTickRadius:          pk.ServerChunkTickRadius,
				HasLockedBehaviourPack:         pk.HasLockedBehaviourPack,
				HasLockedTexturePack:           pk.HasLockedTexturePack,
				FromLockedWorldTemplate:        pk.FromLockedWorldTemplate,
				MSAGamerTagsOnly:               pk.MSAGamerTagsOnly,
				FromWorldTemplate:              pk.FromWorldTemplate,
				WorldTemplateSettingsLocked:    pk.WorldTemplateSettingsLocked,
				OnlySpawnV1Villagers:           pk.OnlySpawnV1Villagers,
				BaseGameVersion:                pk.BaseGameVersion,
				LimitedWorldWidth:              pk.LimitedWorldWidth,
				LimitedWorldDepth:              pk.LimitedWorldDepth,
				NewNether:                      pk.NewNether,
				EducationSharedResourceURI:     pk.EducationSharedResourceURI,
				ForceExperimentalGameplay:      enabled,
				LevelID:                        pk.LevelID,
				WorldName:                      pk.WorldName,
				TemplateContentIdentity:        pk.TemplateContentIdentity,
				Trial:                          pk.Trial,
				PlayerMovementSettings:         pk.PlayerMovementSettings,
				Time:                           pk.Time,
				EnchantmentSeed:                pk.EnchantmentSeed,
				Blocks:                         pk.Blocks,
				Items:                          pk.Items,
				MultiPlayerCorrelationID:       pk.MultiPlayerCorrelationID,
				ServerAuthoritativeInventory:   pk.ServerAuthoritativeInventory,
				GameVersion:                    pk.GameVersion,
				ServerBlockStateChecksum:       pk.ServerBlockStateChecksum,
			}
		case *packet.UpdateAttributes:
			result[i] = &legacypacket.UpdateAttributes{
				EntityRuntimeID: pk.EntityRuntimeID,
				Attributes: lo.Map(pk.Attributes, func(item protocol.Attribute, _ int) types.Attribute {
					return types.Attribute{Attribute: item}
				}),
				Tick: pk.Tick,
			}
		case *packet.StructureBlockUpdate:
			result[i] = &legacypacket.StructureBlockUpdate{
				Position:           pk.Position,
				StructureName:      pk.StructureName,
				DataField:          pk.DataField,
				IncludePlayers:     pk.IncludePlayers,
				ShowBoundingBox:    pk.ShowBoundingBox,
				StructureBlockType: pk.StructureBlockType,
				Settings:           types.StructureSettings{StructureSettings: pk.Settings},
				RedstoneSaveMode:   pk.RedstoneSaveMode,
				ShouldTrigger:      pk.ShouldTrigger,
				Waterlogged:        pk.Waterlogged,
			}
		case *packet.StructureTemplateDataRequest:
			result[i] = &legacypacket.StructureTemplateDataRequest{
				StructureName: pk.StructureName,
				Position:      pk.Position,
				Settings:      types.StructureSettings{StructureSettings: pk.Settings},
				RequestType:   pk.RequestType,
			}
		case *packet.UpdateAbilities:
			handleFlag := func(layers []protocol.AbilityLayer, secondFlag bool) uint32 {
				layerMapping := map[uint32]uint32{
					protocol.AbilityAttackPlayers: packet.AdventureSettingsFlagsNoPvM,
					protocol.AbilityMayFly:        packet.AdventureFlagAllowFlight,
					protocol.AbilityNoClip:        packet.AdventureFlagNoClip,
					protocol.AbilityWorldBuilder:  packet.AdventureFlagWorldBuilder,
					protocol.AbilityFlying:        packet.AdventureFlagFlying,
					protocol.AbilityMuted:         packet.AdventureFlagMuted,
				}
				if secondFlag {
					layerMapping = map[uint32]uint32{
						protocol.AbilityMine:             packet.ActionPermissionMine,
						protocol.AbilityDoorsAndSwitches: packet.ActionPermissionDoorsAndSwitches,
						protocol.AbilityOpenContainers:   packet.ActionPermissionOpenContainers,
						protocol.AbilityAttackPlayers:    packet.ActionPermissionAttackPlayers,
						protocol.AbilityAttackMobs:       packet.ActionPermissionAttackMobs,
						protocol.AbilityOperatorCommands: packet.ActionPermissionOperator,
						protocol.AbilityBuild:            packet.ActionPermissionBuild,
					}
				}

				out := uint32(0)
				for _, layer := range layers {
					for flag, mapped := range layerMapping {
						if (layer.Values & flag) != 0 {
							out |= mapped
						}
					}
				}
				return out
			}

			result[i] = &packet.AdventureSettings{
				Flags:                  handleFlag(pk.AbilityData.Layers, false),
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				ActionPermissions:      handleFlag(pk.AbilityData.Layers, true),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
			}
		case *packet.Emote:
			result[i] = &legacypacket_v582.Emote{
				EntityRuntimeID: pk.EntityRuntimeID,
				EmoteID:         pk.EmoteID,
				Flags:           pk.Flags,
			}
		}
	}

	return result
}
