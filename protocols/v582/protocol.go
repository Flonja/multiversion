package v582

import (
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v582/mappings"
	legacypacket "github.com/flonja/multiversion/protocols/v582/packet"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Protocol struct {
}

func (p Protocol) ID() int32 {
	return 582
}

func (p Protocol) Ver() string {
	return "1.19.80"
}

func (p Protocol) Packets() packet.Pool {
	pool := packet.NewPool()
	pool[packet.IDEmote] = func() packet.Packet { return &legacypacket.Emote{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDUnlockedRecipes] = func() packet.Packet { return &legacypacket.UnlockedRecipes{} }
	return pool
}

func (p Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *legacypacket.Emote:
		return []packet.Packet{
			&packet.Emote{
				EntityRuntimeID: pk.EntityRuntimeID,
				EmoteID:         pk.EmoteID,
				XUID:            conn.IdentityData().XUID,
				PlatformID:      conn.ClientData().PlatformOnlineID,
				Flags:           pk.Flags,
			},
		}
	case *legacypacket.StartGame:
		return []packet.Packet{
			&packet.StartGame{
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
				EditorWorld:                    pk.EditorWorld,
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
				Items: lo.Map(pk.Items, func(item protocol.ItemEntry, _ int) protocol.ItemEntry {
					if !item.ComponentBased {
						itemType := upgradeItemType(protocol.ItemType{
							NetworkID:     int32(item.RuntimeID),
							MetadataValue: 0,
						})
						item.RuntimeID = int16(itemType.NetworkID)

						var ok bool
						if item.Name, ok = latest.ItemRuntimeIDToName(itemType.NetworkID); !ok {
							panic(itemType)
						}
					}
					return item
				}),
				MultiPlayerCorrelationID:     pk.MultiPlayerCorrelationID,
				ServerAuthoritativeInventory: pk.ServerAuthoritativeInventory,
				GameVersion:                  pk.GameVersion,
				PropertyData:                 pk.PropertyData,
				ServerBlockStateChecksum:     pk.ServerBlockStateChecksum,
				ClientSideGeneration:         pk.ClientSideGeneration,
				WorldTemplateID:              pk.WorldTemplateID,
				ChatRestrictionLevel:         pk.ChatRestrictionLevel,
				DisablePlayerInteractions:    pk.DisablePlayerInteractions,
				UseBlockNetworkIDHashes:      pk.UseBlockNetworkIDHashes,
				ServerAuthoritativeSound:     false,
			},
		}
	case *legacypacket.UnlockedRecipes:
		unlockType := packet.UnlockedRecipesTypeInitiallyUnlocked
		if pk.NewUnlocks {
			unlockType = packet.UnlockedRecipesTypeNewlyUnlocked
		}
		return []packet.Packet{
			&packet.UnlockedRecipes{
				UnlockType: packet.UnlockedRecipesTypeRemoveAllUnlocked,
			},
			&packet.UnlockedRecipes{
				UnlockType: uint32(unlockType),
				Recipes:    pk.Recipes,
			},
		}
	}
	return upgradeWorldPackets(upgradeItemPackets([]packet.Packet{pk}), conn.GameData(), conn.ClientCacheEnabled())
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.Emote:
		return []packet.Packet{
			&legacypacket.Emote{
				EntityRuntimeID: pk.EntityRuntimeID,
				EmoteID:         pk.EmoteID,
				Flags:           pk.Flags,
			},
		}
	case *packet.StartGame:
		return []packet.Packet{
			&legacypacket.StartGame{
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
				EditorWorld:                    pk.EditorWorld,
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
				Items: lo.Map(pk.Items, func(item protocol.ItemEntry, _ int) protocol.ItemEntry {
					if !item.ComponentBased {
						itemType := downgradeItemType(protocol.ItemType{
							NetworkID:     int32(item.RuntimeID),
							MetadataValue: 0,
						})
						item.RuntimeID = int16(itemType.NetworkID)

						var ok bool
						if item.Name, ok = mappings.ItemRuntimeIDToName(itemType.NetworkID); !ok {
							panic(itemType)
						}
					}
					return item
				}),
				MultiPlayerCorrelationID:     pk.MultiPlayerCorrelationID,
				ServerAuthoritativeInventory: pk.ServerAuthoritativeInventory,
				GameVersion:                  pk.GameVersion,
				PropertyData:                 pk.PropertyData,
				ServerBlockStateChecksum:     pk.ServerBlockStateChecksum,
				ClientSideGeneration:         pk.ClientSideGeneration,
				WorldTemplateID:              pk.WorldTemplateID,
				ChatRestrictionLevel:         pk.ChatRestrictionLevel,
				DisablePlayerInteractions:    pk.DisablePlayerInteractions,
				UseBlockNetworkIDHashes:      pk.UseBlockNetworkIDHashes,
			},
		}
	case *packet.UnlockedRecipes:
		newUnlocks := false
		if pk.UnlockType == packet.UnlockedRecipesTypeInitiallyUnlocked || pk.UnlockType == packet.UnlockedRecipesTypeNewlyUnlocked {
			newUnlocks = true
		}

		return []packet.Packet{
			&legacypacket.UnlockedRecipes{
				NewUnlocks: newUnlocks,
				Recipes:    pk.Recipes,
			},
		}
	}

	return downgradeWorldPackets(downgradeItemPackets([]packet.Packet{pk}), conn.GameData(), conn.ClientCacheEnabled())
}
