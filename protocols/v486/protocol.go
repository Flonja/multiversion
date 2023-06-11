package v486

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v486/packet"
	"github.com/flonja/multiversion/protocols/v486/types"
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

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	return p.blockTranslator.UpgradeBlockPackets(p.itemTranslator.UpgradeItemPackets([]packet.Packet{pk}, conn), conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets(p.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn), conn)

	for i, pk := range result {
		switch pk := pk.(type) {
		// TODO: add UpdateAbilities
		case *packet.SetActorData:
			pk.EntityMetadata = downgradeEntityMetadata(pk.EntityMetadata)
			result[i] = pk
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
					HeldItem:        pk.HeldItem,
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
	}

	return result
}
