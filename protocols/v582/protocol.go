package v582

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/packbuilder"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v582/items"
	legacypacket "github.com/flonja/multiversion/protocols/v582/packet"
	"github.com/flonja/multiversion/translator"
	"github.com/sandertv/gophertunnel/minecraft"
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
	itemMapping := mapping.NewItemMapping(itemRuntimeIDData, 111)
	blockMapping := mapping.NewBlockMapping(blockStateData)
	latestBlockMapping := latest.NewBlockMapping()

	itemTranslator := translator.NewItemTranslator(itemMapping, latest.NewItemMapping(), blockMapping, latestBlockMapping)
	itemTranslator.Register(items.DiscRelic{}, "minecraft:music_disc_relic")
	return &Protocol{itemMapping: itemMapping, blockMapping: blockMapping,
		itemTranslator:  itemTranslator,
		blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ResourcePack() *resource.Pack {
	resourcePack, ok := packbuilder.BuildResourcePack(maps.Values(p.itemTranslator.CustomItems()), ver)
	if !ok {
		panic("couldn't create resource pack")
	}
	return resourcePack
}

func (Protocol) ID() int32 {
	return 582
}

const ver = "1.19.83"

func (Protocol) Ver() string {
	return ver
}

func (Protocol) Packets() packet.Pool {
	pool := packet.NewPool()
	pool[packet.IDEmote] = func() packet.Packet { return &legacypacket.Emote{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDUnlockedRecipes] = func() packet.Packet { return &legacypacket.UnlockedRecipes{} }
	return pool
}

func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.ClientCacheStatus:
		pk.Enabled = false
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
		var pks []packet.Packet
		// todo: figure out what to do when there are no custom items
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
		pks = append(pks, &packet.StartGame{
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
	return p.blockTranslator.UpgradeBlockPackets(p.itemTranslator.UpgradeItemPackets([]packet.Packet{pk}, conn), conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets(p.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn), conn)

	for _, pk := range result {
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
			var pks []packet.Packet
			// todo: figure out what to do when there are no custom items
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
			pks = append(pks, &legacypacket.StartGame{
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
			})
			return pks
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
	}
	return result
}
