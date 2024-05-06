package v649

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v630/packet"
	"github.com/flonja/multiversion/protocols/v630/types"
	legacypacket_v649 "github.com/flonja/multiversion/protocols/v649/packet"
	v662 "github.com/flonja/multiversion/protocols/v662"
	legacypacket_v662 "github.com/flonja/multiversion/protocols/v662/packet"
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
	itemMapping := mapping.NewItemMapping(itemRuntimeIDData)
	blockMapping := mapping.NewBlockMapping(blockStateData)
	latestBlockMapping := latest.NewBlockMapping()
	return &Protocol{itemMapping: itemMapping, blockMapping: blockMapping,
		itemTranslator:  translator.NewItemTranslator(itemMapping, latest.NewItemMapping(), blockMapping, latestBlockMapping),
		blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ID() int32 {
	return 630
}

func (p Protocol) Ver() string {
	return "1.20.51"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDCorrectPlayerMovePrediction] = func() packet.Packet { return &legacypacket.CorrectPlayerMovePrediction{} }
	pool[packet.IDLevelChunk] = func() packet.Packet { return &legacypacket.LevelChunk{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDPlayerList] = func() packet.Packet { return &legacypacket.PlayerList{} }

	// v649
	pool[packet.IDLecternUpdate] = func() packet.Packet { return &legacypacket_v649.LecternUpdate{} }
	pool[packet.IDMobEffect] = func() packet.Packet { return &legacypacket_v649.MobEffect{} }
	pool[packet.IDResourcePacksInfo] = func() packet.Packet { return &legacypacket_v649.ResourcePacksInfo{} }
	pool[packet.IDActorEvent] = func() packet.Packet { return &legacypacket_v649.SetActorMotion{} }

	// v662
	pool[packet.IDClientBoundDebugRenderer] = func() packet.Packet { return &legacypacket_v662.ClientBoundDebugRenderer{} }
	pool[packet.IDResourcePackStack] = func() packet.Packet { return &legacypacket_v662.ResourcePackStack{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket_v662.StartGame{} }
	pool[packet.IDUpdateBlockSynced] = func() packet.Packet { return &legacypacket_v662.UpdateBlockSynced{} }
	pool[packet.IDUpdatePlayerGameType] = func() packet.Packet { return &legacypacket_v662.UpdatePlayerGameType{} }
	return pool
}

func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return packet.NewCTREncryption(key[:])
}

func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return v662.NewReader(protocol.NewReader(r, shieldID, enableLimits))
}

func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return v662.NewWriter(protocol.NewWriter(w, shieldID))
}

func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	var newPks []packet.Packet

	switch pk := pk.(type) {
	case *legacypacket.CorrectPlayerMovePrediction:
		newPks = append(newPks, &packet.CorrectPlayerMovePrediction{
			Position: pk.Position,
			Delta:    pk.Delta,
			OnGround: pk.OnGround,
			Tick:     pk.Tick,
		})
	case *legacypacket.LevelChunk:
		newPks = append(newPks, &packet.LevelChunk{
			Position:        pk.Position,
			HighestSubChunk: pk.HighestSubChunk,
			SubChunkCount:   pk.SubChunkCount,
			CacheEnabled:    pk.CacheEnabled,
			BlobHashes:      pk.BlobHashes,
			RawPayload:      pk.RawPayload,
		})
	case *legacypacket.PlayerAuthInput:
		newPks = append(newPks, &packet.PlayerAuthInput{
			Pitch:               pk.Pitch,
			Yaw:                 pk.Yaw,
			Position:            pk.Position,
			MoveVector:          pk.MoveVector,
			HeadYaw:             pk.HeadYaw,
			InputData:           pk.InputData,
			InputMode:           pk.InputMode,
			PlayMode:            pk.PlayMode,
			InteractionModel:    uint32(pk.InteractionModel),
			GazeDirection:       pk.GazeDirection,
			Tick:                pk.Tick,
			Delta:               pk.Delta,
			ItemInteractionData: pk.ItemInteractionData,
			ItemStackRequest:    pk.ItemStackRequest,
			BlockActions:        pk.BlockActions,
			AnalogueMoveVector:  pk.AnalogueMoveVector,
		})
	case *legacypacket.PlayerList:
		newPks = append(newPks, &packet.PlayerList{
			ActionType: pk.ActionType,
			Entries: lo.Map(pk.Entries, func(item types.PlayerListEntry, index int) protocol.PlayerListEntry {
				return protocol.PlayerListEntry{
					UUID:           item.UUID,
					EntityUniqueID: item.EntityUniqueID,
					Username:       item.Username,
					XUID:           item.XUID,
					PlatformChatID: item.PlatformChatID,
					BuildPlatform:  item.BuildPlatform,
					Skin:           item.Skin,
					Teacher:        item.Teacher,
					Host:           item.Host,
				}
			}),
		})
	case *legacypacket_v649.LecternUpdate:
		if pk.DropBook {
			break
		}
		newPks = append(newPks, &packet.LecternUpdate{
			Page:      pk.Page,
			PageCount: pk.PageCount,
			Position:  pk.Position,
		})
	case *legacypacket_v649.MobEffect:
		newPks = append(newPks, &packet.MobEffect{
			EntityRuntimeID: pk.EntityRuntimeID,
			Operation:       pk.Operation,
			EffectType:      pk.EffectType,
			Amplifier:       pk.Amplifier,
			Particles:       pk.Particles,
			Duration:        pk.Duration,
		})
	case *legacypacket_v649.ResourcePacksInfo:
		newPks = append(newPks, &packet.ResourcePacksInfo{
			TexturePackRequired: pk.TexturePackRequired,
			HasScripts:          pk.HasScripts,
			BehaviourPacks:      pk.BehaviourPacks,
			TexturePacks:        pk.TexturePacks,
			ForcingServerPacks:  pk.ForcingServerPacks,
			PackURLs:            pk.PackURLs,
		})
	case *legacypacket_v649.SetActorMotion:
		newPks = append(newPks, &packet.SetActorMotion{
			EntityRuntimeID: pk.EntityRuntimeID,
			Velocity:        pk.Velocity,
		})
	case *legacypacket_v662.ClientBoundDebugRenderer:
		newPks = append(newPks, &packet.ClientBoundDebugRenderer{
			Type:     pk.Type,
			Text:     pk.Text,
			Position: pk.Position,
			Red:      pk.Red,
			Green:    pk.Green,
			Blue:     pk.Blue,
			Alpha:    pk.Alpha,
			Duration: pk.Duration,
		})
	case *legacypacket_v662.ResourcePackStack:
		newPks = append(newPks, &packet.ResourcePackStack{
			TexturePackRequired:          pk.TexturePackRequired,
			BehaviourPacks:               pk.BehaviourPacks,
			TexturePacks:                 pk.TexturePacks,
			BaseGameVersion:              pk.BaseGameVersion,
			Experiments:                  pk.Experiments,
			ExperimentsPreviouslyToggled: pk.ExperimentsPreviouslyToggled,
		})
	case *legacypacket_v662.StartGame:
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
			EditorWorldType:                pk.EditorWorldType,
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
			ServerAuthoritativeSound:       pk.ServerAuthoritativeSound,
		})
	case *legacypacket_v662.UpdateBlockSynced:
		newPks = append(newPks, &packet.UpdateBlockSynced{
			Position:          pk.Position,
			NewBlockRuntimeID: pk.NewBlockRuntimeID,
			Flags:             pk.Flags,
			Layer:             pk.Layer,
			EntityUniqueID:    uint64(pk.EntityUniqueID),
			TransitionType:    pk.TransitionType,
		})
	case *legacypacket_v662.UpdatePlayerGameType:
		newPks = append(newPks, &packet.UpdatePlayerGameType{
			GameType:       pk.GameType,
			PlayerUniqueID: pk.PlayerUniqueID,
		})
	case *packet.AvailableCommands:
		for ind1, command := range pk.Commands {
			for ind2, overload := range command.Overloads {
				for ind3, parameter := range overload.Parameters {
					parameterType := uint32(parameter.Type) | protocol.CommandArgValid

					switch parameter.Type | protocol.CommandArgValid {
					case 43:
						parameterType = protocol.CommandArgTypeEquipmentSlots
					case 44:
						parameterType = protocol.CommandArgTypeString
					case 52:
						parameterType = protocol.CommandArgTypeBlockPosition
					case 53:
						parameterType = protocol.CommandArgTypePosition
					case 55:
						parameterType = protocol.CommandArgTypeMessage
					case 58:
						parameterType = protocol.CommandArgTypeRawText
					case 62:
						parameterType = protocol.CommandArgTypeJSON
					case 71:
						parameterType = protocol.CommandArgTypeBlockStates
					case 74:
						parameterType = protocol.CommandArgTypeCommand
					}
					parameter.Type = parameterType | protocol.CommandArgValid
					pk.Commands[ind1].Overloads[ind2].Parameters[ind3] = parameter
				}
			}
		}
	case *packet.LevelEvent:
		if pk.EventType&packet.LevelEventParticleLegacyEvent != 0 {
			particleId := pk.EventType ^ packet.LevelEventParticleLegacyEvent
			if particleId >= 90 { // wind explosion
				particleId--
			}
			pk.EventType = packet.LevelEventParticleLegacyEvent | particleId
		}
	case *packet.ClientCacheStatus:
		pk.Enabled = false
	default:
		newPks = append(newPks, pk)
	}

	return p.blockTranslator.UpgradeBlockPackets(p.itemTranslator.UpgradeItemPackets(newPks, conn), conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets(p.itemTranslator.DowngradeItemPackets([]packet.Packet{pk}, conn), conn)

	for i, pk := range result {
		switch pk := pk.(type) {
		case *packet.CorrectPlayerMovePrediction:
			result[i] = &legacypacket.CorrectPlayerMovePrediction{
				Position: pk.Position,
				Delta:    pk.Delta,
				OnGround: pk.OnGround,
				Tick:     pk.Tick,
			}
		case *packet.LevelChunk:
			result[i] = &legacypacket.LevelChunk{
				Position:        pk.Position,
				HighestSubChunk: pk.HighestSubChunk,
				SubChunkCount:   pk.SubChunkCount,
				CacheEnabled:    pk.CacheEnabled,
				BlobHashes:      pk.BlobHashes,
				RawPayload:      pk.RawPayload,
			}
		case *packet.PlayerAuthInput:
			result[i] = &legacypacket.PlayerAuthInput{
				Pitch:               pk.Pitch,
				Yaw:                 pk.Yaw,
				Position:            pk.Position,
				MoveVector:          pk.MoveVector,
				HeadYaw:             pk.HeadYaw,
				InputData:           pk.InputData,
				InputMode:           pk.InputMode,
				PlayMode:            pk.PlayMode,
				InteractionModel:    int32(pk.InteractionModel),
				GazeDirection:       pk.GazeDirection,
				Tick:                pk.Tick,
				Delta:               pk.Delta,
				ItemInteractionData: pk.ItemInteractionData,
				ItemStackRequest:    pk.ItemStackRequest,
				BlockActions:        pk.BlockActions,
				AnalogueMoveVector:  pk.AnalogueMoveVector,
			}
		case *packet.PlayerList:
			result[i] = &legacypacket.PlayerList{
				ActionType: pk.ActionType,
				Entries: lo.Map(pk.Entries, func(item protocol.PlayerListEntry, index int) types.PlayerListEntry {
					return types.PlayerListEntry{
						UUID:           item.UUID,
						EntityUniqueID: item.EntityUniqueID,
						Username:       item.Username,
						XUID:           item.XUID,
						PlatformChatID: item.PlatformChatID,
						BuildPlatform:  item.BuildPlatform,
						Skin:           item.Skin,
						Teacher:        item.Teacher,
						Host:           item.Host,
					}
				}),
			}
		case *packet.LecternUpdate:
			result[i] = &legacypacket_v649.LecternUpdate{
				Page:      pk.Page,
				PageCount: pk.PageCount,
				Position:  pk.Position,
			}
		case *packet.MobEffect:
			result[i] = &legacypacket_v649.MobEffect{
				EntityRuntimeID: pk.EntityRuntimeID,
				Operation:       pk.Operation,
				EffectType:      pk.EffectType,
				Amplifier:       pk.Amplifier,
				Particles:       pk.Particles,
				Duration:        pk.Duration,
			}
		case *packet.ResourcePacksInfo:
			result[i] = &legacypacket_v649.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				BehaviourPacks:      pk.BehaviourPacks,
				TexturePacks:        pk.TexturePacks,
				ForcingServerPacks:  pk.ForcingServerPacks,
				PackURLs:            pk.PackURLs,
			}
		case *packet.SetActorMotion:
			result[i] = &legacypacket_v649.SetActorMotion{
				EntityRuntimeID: pk.EntityRuntimeID,
				Velocity:        pk.Velocity,
			}
		case *packet.ClientBoundDebugRenderer:
			result[i] = &legacypacket_v662.ClientBoundDebugRenderer{
				Type:     pk.Type,
				Text:     pk.Text,
				Position: pk.Position,
				Red:      pk.Red,
				Green:    pk.Green,
				Blue:     pk.Blue,
				Alpha:    pk.Alpha,
				Duration: pk.Duration,
			}
		case *packet.ResourcePackStack:
			result[i] = &legacypacket_v662.ResourcePackStack{
				TexturePackRequired:          pk.TexturePackRequired,
				BehaviourPacks:               pk.BehaviourPacks,
				TexturePacks:                 pk.TexturePacks,
				BaseGameVersion:              pk.BaseGameVersion,
				Experiments:                  pk.Experiments,
				ExperimentsPreviouslyToggled: pk.ExperimentsPreviouslyToggled,
			}
		case *packet.StartGame:
			result[i] = &legacypacket_v662.StartGame{
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
				EditorWorldType:                pk.EditorWorldType,
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
				ServerAuthoritativeSound:       pk.ServerAuthoritativeSound,
			}
		case *packet.UpdateBlockSynced:
			result[i] = &legacypacket_v662.UpdateBlockSynced{
				Position:          pk.Position,
				NewBlockRuntimeID: pk.NewBlockRuntimeID,
				Flags:             pk.Flags,
				Layer:             pk.Layer,
				EntityUniqueID:    int64(pk.EntityUniqueID),
				TransitionType:    pk.TransitionType,
			}
		case *packet.UpdatePlayerGameType:
			result[i] = &legacypacket_v662.UpdatePlayerGameType{
				GameType:       pk.GameType,
				PlayerUniqueID: pk.PlayerUniqueID,
			}
		case *packet.AvailableCommands:
			for ind1, command := range pk.Commands {
				for ind2, overload := range command.Overloads {
					for ind3, parameter := range overload.Parameters {
						parameterType := uint32(parameter.Type) | protocol.CommandArgValid

						switch parameter.Type | protocol.CommandArgValid {
						case protocol.CommandArgTypeEquipmentSlots:
							parameterType = 43
						case protocol.CommandArgTypeString:
							parameterType = 44
						case protocol.CommandArgTypeBlockPosition:
							parameterType = 52
						case protocol.CommandArgTypePosition:
							parameterType = 53
						case protocol.CommandArgTypeMessage:
							parameterType = 55
						case protocol.CommandArgTypeRawText:
							parameterType = 58
						case protocol.CommandArgTypeJSON:
							parameterType = 62
						case protocol.CommandArgTypeBlockStates:
							parameterType = 71
						case protocol.CommandArgTypeCommand:
							parameterType = 74
						}
						parameter.Type = parameterType | protocol.CommandArgValid
						pk.Commands[ind1].Overloads[ind2].Parameters[ind3] = parameter
					}
				}
			}
			result[i] = pk
		case *packet.LevelEvent:
			if pk.EventType&packet.LevelEventParticleLegacyEvent != 0 {
				particleId := pk.EventType ^ packet.LevelEventParticleLegacyEvent
				if particleId >= 91 { // wind explosion
					particleId--
				}
				pk.EventType = packet.LevelEventParticleLegacyEvent | particleId
			}
		}
	}

	return result
}
