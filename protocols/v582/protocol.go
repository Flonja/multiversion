package v582

import (
	_ "embed"
	"github.com/df-mc/worldupgrader/itemupgrader"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/packbuilder"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v582/items"
	legacypacket "github.com/flonja/multiversion/protocols/v582/packet"
	legacypacket_v589 "github.com/flonja/multiversion/protocols/v589/packet"
	"github.com/flonja/multiversion/protocols/v589/types"
	"github.com/flonja/multiversion/translator"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"golang.org/x/exp/maps"
)

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte
	//go:embed block_states.nbt
	blockStateData []byte
)

type Protocol struct {
	itemMapping     mapping.Item
	blockMapping    mapping.Block
	itemTranslator  translator.ItemTranslator
	blockTranslator translator.BlockTranslator
}

func New() *Protocol {
	itemMapping := mapping.NewItemMapping(itemRuntimeIDData)
	blockMapping := mapping.NewBlockMapping(blockStateData)
	latestBlockMapping := latest.NewBlockMapping()

	itemTranslator := translator.NewItemTranslator(itemMapping, latest.NewItemMapping(), blockMapping, latestBlockMapping)
	itemTranslator.Register(items.DiscRelic{}, itemupgrader.ItemMeta{Name: "minecraft:music_disc_relic"})
	return &Protocol{itemMapping: itemMapping, blockMapping: blockMapping,
		itemTranslator:  itemTranslator,
		blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ResourcePack(ver string) *resource.Pack {
	resourcePack, ok := packbuilder.BuildResourcePack(maps.Values(p.itemTranslator.CustomItems()), ver)
	if !ok {
		panic("couldn't create resource pack")
	}
	return resourcePack
}

func (Protocol) ID() int32 {
	return 582
}

func (Protocol) Ver() string {
	return "1.19.83"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDAvailableCommands] = func() packet.Packet { return &legacypacket_v589.AvailableCommands{} }
	pool[packet.IDEmote] = func() packet.Packet { return &legacypacket.Emote{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDUnlockedRecipes] = func() packet.Packet { return &legacypacket.UnlockedRecipes{} }
	return pool
}

func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return protocol.NewReader(r, shieldID, enableLimits)
}

func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return protocol.NewWriter(w, shieldID)
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	var newPks []packet.Packet
	switch pk := pk.(type) {
	case *packet.ClientCacheStatus:
		pk.Enabled = false
		newPks = append(newPks, pk)
	case *legacypacket.Emote:
		newPks = append(newPks, &packet.Emote{
			EntityRuntimeID: pk.EntityRuntimeID,
			EmoteID:         pk.EmoteID,
			XUID:            conn.IdentityData().XUID,
			PlatformID:      conn.ClientData().PlatformOnlineID,
			Flags:           pk.Flags,
		})
	case *legacypacket.StartGame:
		// todon't: figure out what to do when there are no custom items
		//if len(lo.Filter(pk.Items, func(item protocol.ItemEntry, _ int) bool {
		//	return item.ComponentBased
		//})) == 0 {
		//	pks = append(pks, &packet.ItemComponent{
		//		Items: func() (entries []protocol.ItemComponentEntry) {
		//			for _, item := range p.itemMapping.CustomItems() {
		//				name, _ := item.EncodeItem()
		//				entries = append(entries, protocol.ItemComponentEntry{
		//					Name: name,
		//					Data: packbuilder.Components(item),
		//				})
		//			}
		//			return entries
		//		}(),
		//	})
		//}
		editorWorldType := int32(packet.EditorWorldTypeNotEditor)
		if pk.EditorWorld {
			editorWorldType = packet.EditorWorldTypeProject
		}
		newPks = append(newPks, &packet.StartGame{
			EntityUniqueID:                 pk.EntityUniqueID,
			EntityRuntimeID:                pk.EntityRuntimeID,
			PlayerGameMode:                 pk.PlayerGameMode,
			PlayerPosition:                 pk.PlayerPosition,
			Pitch:                          pk.Pitch,
			Yaw:                            pk.Yaw,
			WorldSeed:                      pk.WorldSeed,
			SpawnBiomeType:                 pk.SpawnBiomeType,
			UserDefinedBiomeName:           pk.UserDefinedBiomeName,
			Dimension:                      pk.Dimension,
			Generator:                      pk.Generator,
			WorldGameMode:                  pk.WorldGameMode,
			Difficulty:                     pk.Difficulty,
			WorldSpawn:                     pk.WorldSpawn,
			AchievementsDisabled:           pk.AchievementsDisabled,
			EditorWorldType:                editorWorldType,
			CreatedInEditor:                pk.CreatedInEditor,
			ExportedFromEditor:             pk.ExportedFromEditor,
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
			PersonaDisabled:                pk.PersonaDisabled,
			CustomSkinsDisabled:            pk.CustomSkinsDisabled,
			EmoteChatMuted:                 pk.EmoteChatMuted,
			BaseGameVersion:                pk.BaseGameVersion,
			LimitedWorldWidth:              pk.LimitedWorldWidth,
			LimitedWorldDepth:              pk.LimitedWorldDepth,
			NewNether:                      pk.NewNether,
			EducationSharedResourceURI:     pk.EducationSharedResourceURI,
			ForceExperimentalGameplay:      pk.ForceExperimentalGameplay,
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
			PropertyData:                   pk.PropertyData,
			ServerBlockStateChecksum:       pk.ServerBlockStateChecksum,
			ClientSideGeneration:           pk.ClientSideGeneration,
			WorldTemplateID:                pk.WorldTemplateID,
			ChatRestrictionLevel:           pk.ChatRestrictionLevel,
			DisablePlayerInteractions:      pk.DisablePlayerInteractions,
			UseBlockNetworkIDHashes:        pk.UseBlockNetworkIDHashes,
			ServerAuthoritativeSound:       false,
		})
	case *legacypacket.UnlockedRecipes:
		unlockType := packet.UnlockedRecipesTypeInitiallyUnlocked
		if pk.NewUnlocks {
			unlockType = packet.UnlockedRecipesTypeNewlyUnlocked
		}

		newPks = append(newPks,
			&packet.UnlockedRecipes{
				UnlockType: packet.UnlockedRecipesTypeRemoveAllUnlocked,
			},
			&packet.UnlockedRecipes{
				UnlockType: uint32(unlockType),
				Recipes:    pk.Recipes,
			})
	case *legacypacket_v589.AvailableCommands:
		newPks = append(newPks, &packet.AvailableCommands{
			EnumValues: pk.EnumValues,
			Suffixes:   pk.Suffixes,
			Enums:      pk.Enums,
			Commands: lo.Map(pk.Commands, func(item types.Command, _ int) protocol.Command {
				return protocol.Command{
					Name:            item.Name,
					Description:     item.Description,
					Flags:           item.Flags,
					PermissionLevel: item.PermissionLevel,
					AliasesOffset:   item.AliasesOffset,
					Overloads: lo.Map(item.Overloads, func(item types.CommandOverload, _ int) protocol.CommandOverload {
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
	default:
		newPks = append(newPks, pk)
	}
	return p.blockTranslator.UpgradeBlockPackets(p.itemTranslator.UpgradeItemPackets(newPks, conn), conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets(p.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn), conn)

	for i, pk := range result {
		switch pk := pk.(type) {
		case *packet.Emote:
			result[i] = &legacypacket.Emote{
				EntityRuntimeID: pk.EntityRuntimeID,
				EmoteID:         pk.EmoteID,
				Flags:           pk.Flags,
			}
		case *packet.StartGame:
			// todon't: figure out what to do when there are no custom items
			//if len(lo.Filter(pk.Items, func(item protocol.ItemEntry, _ int) bool {
			//	return item.ComponentBased
			//})) == 0 {
			//pks = append(pks, &packet.ItemComponent{
			//	Items: func() (entries []protocol.ItemComponentEntry) {
			//		for _, item := range p.itemMapping.CustomItems() {
			//			name, _ := item.EncodeItem()
			//			entries = append(entries, protocol.ItemComponentEntry{
			//				Name: name,
			//				Data: packbuilder.Components(item),
			//			})
			//		}
			//		return entries
			//	}(),
			//})
			//}
			editorWorld := false
			if pk.EditorWorldType > packet.EditorWorldTypeNotEditor {
				editorWorld = true
			}
			result[i] = &legacypacket.StartGame{
				EntityUniqueID:                 pk.EntityUniqueID,
				EntityRuntimeID:                pk.EntityRuntimeID,
				PlayerGameMode:                 pk.PlayerGameMode,
				PlayerPosition:                 pk.PlayerPosition,
				Pitch:                          pk.Pitch,
				Yaw:                            pk.Yaw,
				WorldSeed:                      pk.WorldSeed,
				SpawnBiomeType:                 pk.SpawnBiomeType,
				UserDefinedBiomeName:           pk.UserDefinedBiomeName,
				Dimension:                      pk.Dimension,
				Generator:                      pk.Generator,
				WorldGameMode:                  pk.WorldGameMode,
				Difficulty:                     pk.Difficulty,
				WorldSpawn:                     pk.WorldSpawn,
				AchievementsDisabled:           pk.AchievementsDisabled,
				EditorWorld:                    editorWorld,
				CreatedInEditor:                pk.CreatedInEditor,
				ExportedFromEditor:             pk.ExportedFromEditor,
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
				PersonaDisabled:                pk.PersonaDisabled,
				CustomSkinsDisabled:            pk.CustomSkinsDisabled,
				EmoteChatMuted:                 pk.EmoteChatMuted,
				BaseGameVersion:                pk.BaseGameVersion,
				LimitedWorldWidth:              pk.LimitedWorldWidth,
				LimitedWorldDepth:              pk.LimitedWorldDepth,
				NewNether:                      pk.NewNether,
				EducationSharedResourceURI:     pk.EducationSharedResourceURI,
				ForceExperimentalGameplay:      pk.ForceExperimentalGameplay,
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
				PropertyData:                   pk.PropertyData,
				ServerBlockStateChecksum:       pk.ServerBlockStateChecksum,
				ClientSideGeneration:           pk.ClientSideGeneration,
				WorldTemplateID:                pk.WorldTemplateID,
				ChatRestrictionLevel:           pk.ChatRestrictionLevel,
				DisablePlayerInteractions:      pk.DisablePlayerInteractions,
				UseBlockNetworkIDHashes:        pk.UseBlockNetworkIDHashes,
			}
		case *packet.UnlockedRecipes:
			newUnlocks := false
			if pk.UnlockType == packet.UnlockedRecipesTypeInitiallyUnlocked || pk.UnlockType == packet.UnlockedRecipesTypeNewlyUnlocked {
				newUnlocks = true
			}
			result[i] = &legacypacket.UnlockedRecipes{
				NewUnlocks: newUnlocks,
				Recipes:    pk.Recipes,
			}
		case *packet.AvailableCommands:
			result[i] = &legacypacket_v589.AvailableCommands{
				EnumValues: pk.EnumValues,
				Suffixes:   pk.Suffixes,
				Enums:      pk.Enums,
				Commands: lo.Map(pk.Commands, func(item protocol.Command, _ int) types.Command {
					return types.Command{
						Name:            item.Name,
						Description:     item.Description,
						Flags:           item.Flags,
						PermissionLevel: item.PermissionLevel,
						AliasesOffset:   item.AliasesOffset,
						Overloads: lo.Map(item.Overloads, func(item protocol.CommandOverload, _ int) types.CommandOverload {
							return types.CommandOverload{
								Parameters: item.Parameters,
							}
						}),
					}
				}),
				DynamicEnums: pk.DynamicEnums,
				Constraints:  pk.Constraints,
			}
		}
	}
	return result
}
