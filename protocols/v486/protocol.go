package v486

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/flonja/multiversion/internal/chunk"
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
	legacypacket "github.com/flonja/multiversion/protocols/v486/packet"
	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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
	pool[packet.IDAddPlayer] = func() packet.Packet { return &legacypacket.AddPlayer{} }
	pool[packet.IDAddVolumeEntity] = func() packet.Packet { return &legacypacket.AddVolumeEntity{} }
	pool[packet.IDNetworkChunkPublisherUpdate] = func() packet.Packet { return &legacypacket.NetworkChunkPublisherUpdate{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDRemoveVolumeEntity] = func() packet.Packet { return &legacypacket.RemoveVolumeEntity{} }
	pool[packet.IDSpawnParticleEffect] = func() packet.Packet { return &legacypacket.SpawnParticleEffect{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDUpdateAttributes] = func() packet.Packet { return &legacypacket.UpdateAttributes{} }
	return pool
}

func (p Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

var (
	// latestAirRID is the runtime ID of the air block in the latest version of the game.
	latestAirRID, _ = latest.StateToRuntimeID(blockupgrader.BlockState{Name: "minecraft:air"})
	// legacyAirRID is the runtime ID of the air block in the legacy version of the game.
	legacyAirRID, _ = mappings.StateToRuntimeID(blockupgrader.BlockState{Name: "minecraft:air"})
)

func (p Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *legacypacket.PlayerAuthInput:
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
		return []packet.Packet{
			&packet.PlayerAuthInput{
				Pitch:               pk.Pitch,
				Yaw:                 pk.Yaw,
				Position:            pk.Position,
				MoveVector:          pk.MoveVector,
				HeadYaw:             pk.HeadYaw,
				InputData:           pk.InputData,
				InputMode:           pk.InputMode,
				PlayMode:            pk.PlayMode,
				InteractionModel:    packet.InteractionModelClassic,
				GazeDirection:       pk.GazeDirection,
				Tick:                pk.Tick,
				Delta:               pk.Delta,
				ItemInteractionData: pk.ItemInteractionData,
				ItemStackRequest:    pk.ItemStackRequest,
				BlockActions:        pk.BlockActions,
				AnalogueMoveVector:  mgl32.Vec2{},
			},
		}
	}
	return []packet.Packet{pk}
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	// TODO: add UpdateAbilities
	case *packet.LevelChunk:
		count := int(pk.SubChunkCount)
		if pk.SubChunkCount == protocol.SubChunkRequestModeLimited {
			count = int(pk.HighestSubChunk)
		}
		fmt.Println(count)
		fmt.Println(pk.CacheEnabled)

		buf := bytes.NewBuffer(pk.RawPayload)
		oldFormat := conn.GameData().BaseGameVersion == "1.17.40"
		c, err := chunk.NetworkDecode(latestAirRID, buf, count, oldFormat, world.Overworld.Range())
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
	case *packet.UpdateBlock:
		pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
	case *packet.UpdateBlockSynced:
		pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
	case *packet.MobEquipment:
		pk.NewItem = downgradeItemInstance(pk.NewItem)
	case *packet.MobArmourEquipment:
		pk.Helmet = downgradeItemInstance(pk.Helmet)
		pk.Chestplate = downgradeItemInstance(pk.Chestplate)
		pk.Leggings = downgradeItemInstance(pk.Leggings)
		pk.Boots = downgradeItemInstance(pk.Boots)
	case *packet.AddItemActor:
		pk.Item = downgradeItemInstance(pk.Item)
		pk.EntityMetadata = downgradeEntityMetadata(pk.EntityMetadata)
	case *packet.CraftingEvent:
		pk.Input = lo.Map(pk.Input, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
			return downgradeItemInstance(item)
		})
		pk.Output = lo.Map(pk.Output, func(item protocol.ItemInstance, _ int) protocol.ItemInstance {
			return downgradeItemInstance(item)
		})
	// Changed encoding/decoding:
	case *packet.StartGame:
		return []packet.Packet{
			&legacypacket.StartGame{
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
				ServerBlockStateChecksum:       pk.ServerBlockStateChecksum,
			},
		}
	case *packet.NetworkChunkPublisherUpdate:
		return []packet.Packet{
			&legacypacket.NetworkChunkPublisherUpdate{
				Position: pk.Position,
				Radius:   pk.Radius,
			},
		}

	case *packet.AddActor:
		return []packet.Packet{
			&legacypacket.AddActor{
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
			},
		}
	case *packet.AddPlayer:
		return []packet.Packet{
			&legacypacket.AddPlayer{
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
				HeldItem:        downgradeItemInstance(pk.HeldItem),
				EntityMetadata:  downgradeEntityMetadata(pk.EntityMetadata),
				AdventureSettings: types.AdventureSettings{
					CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
					PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				},
				DeviceID:    pk.DeviceID,
				EntityLinks: pk.EntityLinks,
			},
		}
	case *packet.UpdateAttributes:
		return []packet.Packet{
			&legacypacket.UpdateAttributes{
				EntityRuntimeID: pk.EntityRuntimeID,
				Attributes: lo.Map(pk.Attributes, func(item protocol.Attribute, index int) types.Attribute {
					return types.Attribute{Attribute: item}
				}),
				Tick: pk.Tick,
			},
		}
	}

	return []packet.Packet{pk}
}
