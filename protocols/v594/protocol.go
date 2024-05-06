package v594

import (
	_ "embed"
	"fmt"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v594/packet"
	"github.com/flonja/multiversion/protocols/v594/types"
	legacypacket_v618 "github.com/flonja/multiversion/protocols/v618/packet"
	legacypacket_v622 "github.com/flonja/multiversion/protocols/v622/packet"
	legacypacket_v630 "github.com/flonja/multiversion/protocols/v630/packet"
	types_v630 "github.com/flonja/multiversion/protocols/v630/types"
	legacypacket_v649 "github.com/flonja/multiversion/protocols/v649/packet"
	v662 "github.com/flonja/multiversion/protocols/v662"
	legacypacket_v662 "github.com/flonja/multiversion/protocols/v662/packet"
	"github.com/flonja/multiversion/translator"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"image/color"
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
	return 594
}

func (p Protocol) Ver() string {
	return "1.20.15"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDCameraInstruction] = func() packet.Packet { return &legacypacket.CameraInstruction{} }
	pool[packet.IDCameraPresets] = func() packet.Packet { return &legacypacket.CameraPresets{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }

	// v618
	pool[packet.IDShowStoreOffer] = func() packet.Packet { return &legacypacket_v618.Disconnect{} }

	// v622
	pool[packet.IDShowStoreOffer] = func() packet.Packet { return &legacypacket_v622.ShowStoreOffer{} }

	// v630
	pool[packet.IDCorrectPlayerMovePrediction] = func() packet.Packet { return &legacypacket_v630.CorrectPlayerMovePrediction{} }
	pool[packet.IDLevelChunk] = func() packet.Packet { return &legacypacket_v630.LevelChunk{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket_v630.PlayerAuthInput{} }
	pool[packet.IDPlayerList] = func() packet.Packet { return &legacypacket_v630.PlayerList{} }

	// v649
	pool[packet.IDLecternUpdate] = func() packet.Packet { return &legacypacket_v649.LecternUpdate{} }
	pool[packet.IDMobEffect] = func() packet.Packet { return &legacypacket_v649.MobEffect{} }
	pool[packet.IDResourcePacksInfo] = func() packet.Packet { return &legacypacket_v649.ResourcePacksInfo{} }
	pool[packet.IDActorEvent] = func() packet.Packet { return &legacypacket_v649.SetActorMotion{} }

	// v662
	pool[packet.IDClientBoundDebugRenderer] = func() packet.Packet { return &legacypacket_v662.ClientBoundDebugRenderer{} }
	pool[packet.IDResourcePackStack] = func() packet.Packet { return &legacypacket_v662.ResourcePackStack{} }
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
	case *legacypacket.CameraInstruction:
		newPks = append(newPks, &packet.CameraInstruction{
			Set: mapOptional(getOptionalValueFromMap[map[string]any](pk.Data, "set"), func(v map[string]any) protocol.CameraInstructionSet {
				return protocol.CameraInstructionSet{
					Preset: getValueFromMap[uint32](v, "preset"),
					Ease: mapOptional(getOptionalValueFromMap[map[string]any](v, "ease"), func(v map[string]any) protocol.CameraEase {
						return protocol.CameraEase{
							Type:     types.CameraEase(getValueFromMap[string](v, "type")),
							Duration: getValueFromMap[float32](v, "time"),
						}
					}),
					Position: mapOptional(getOptionalValueFromMap[map[string]any](v, "pos"), func(m map[string]any) mgl32.Vec3 {
						v, ok := m["pos"]
						if !ok {
							return mgl32.Vec3{}
						}
						value, ok := v.(mgl32.Vec3)
						if !ok {
							panic(fmt.Errorf("`pos` has the incorrect type (got %T, expected mgl32.Vec3)", v))
						}
						return value
					}),
					Rotation: mapOptional(getOptionalValueFromMap[map[string]any](v, "rot"), func(v map[string]any) mgl32.Vec2 {
						return mgl32.Vec2{
							getValueFromMap[float32](v, "x"),
							getValueFromMap[float32](v, "y"),
						}
					}),
					Facing: mapOptional(getOptionalValueFromMap[map[string]any](v, "facing"), func(m map[string]any) mgl32.Vec3 {
						v, ok := m["facing"]
						if !ok {
							return mgl32.Vec3{}
						}
						value, ok := v.(mgl32.Vec3)
						if !ok {
							panic(fmt.Errorf("`facing` has the incorrect type (got %T, expected mgl32.Vec3)", v))
						}
						return value
					}),
					Default: mapOptional(getOptionalValueFromMap[byte](pk.Data, "default"), mapByteToBool),
				}
			}),
			Clear: mapOptional(getOptionalValueFromMap[byte](pk.Data, "clear"), mapByteToBool),
			Fade: mapOptional(getOptionalValueFromMap[map[string]any](pk.Data, "fade"), func(v map[string]any) protocol.CameraInstructionFade {
				rawTime := getValueFromMap[map[string]any](v, "time")
				rawColor := getValueFromMap[map[string]any](v, "color")
				return protocol.CameraInstructionFade{
					TimeData: protocol.Option(protocol.CameraFadeTimeData{
						FadeInDuration:  getValueFromMap[float32](rawTime, "fadeIn"),
						WaitDuration:    getValueFromMap[float32](rawTime, "hold"),
						FadeOutDuration: getValueFromMap[float32](rawTime, "fadeOut"),
					}),
					Colour: protocol.Option(color.RGBA{
						R: uint8(getValueFromMap[float32](rawColor, "red") * 255),
						G: uint8(getValueFromMap[float32](rawColor, "green") * 255),
						B: uint8(getValueFromMap[float32](rawColor, "blue") * 255),
					}),
				}
			}),
		})
	case *legacypacket.CameraPresets:
		var presets []protocol.CameraPreset
		rawPresets := getValueFromMap[[]map[string]any](pk.Data, "presets")
		for _, preset := range rawPresets {
			presets = append(presets, protocol.CameraPreset{
				Name:   getValueFromMap[string](preset, "identifier"),
				Parent: getValueFromMap[string](preset, "inherit_from"),
				PosX:   getOptionalValueFromMap[float32](preset, "pos_x"),
				PosY:   getOptionalValueFromMap[float32](preset, "pos_y"),
				PosZ:   getOptionalValueFromMap[float32](preset, "pos_z"),
				RotX:   getOptionalValueFromMap[float32](preset, "rot_x"),
				RotY:   getOptionalValueFromMap[float32](preset, "rot_y"),
				AudioListener: mapOptional(getOptionalValueFromMap[string](preset, "audio_listener_type"), func(v string) byte {
					switch v {
					case "camera":
						return byte(protocol.AudioListenerCamera)
					case "player":
						return byte(protocol.AudioListenerPlayer)
					default:
						panic(fmt.Errorf("invalid audio listener type: %v", v))
					}
				}),
				PlayerEffects: mapOptional(getOptionalValueFromMap[byte](preset, "player_effects"), mapByteToBool),
			})
		}
		newPks = append(newPks, &packet.CameraPresets{
			Presets: presets,
		})
	case *legacypacket.ResourcePacksInfo:
		newPks = append(newPks, &packet.ResourcePacksInfo{
			TexturePackRequired: pk.TexturePackRequired,
			HasScripts:          pk.HasScripts,
			BehaviourPacks:      pk.BehaviourPacks,
			TexturePacks:        pk.TexturePacks,
			ForcingServerPacks:  pk.ForcingServerPacks,
		})
	case *legacypacket.StartGame:
		editorWorldType := packet.EditorWorldTypeNotEditor
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
			EditorWorldType:                int32(editorWorldType),
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
	case *legacypacket_v618.Disconnect:
		newPks = append(newPks, &packet.Disconnect{
			HideDisconnectionScreen: pk.HideDisconnectionScreen,
			Message:                 pk.Message,
		})
	case *legacypacket_v622.ShowStoreOffer:
		newPks = append(newPks, &packet.ShowStoreOffer{
			OfferID: pk.OfferID,
			Type:    packet.StoreOfferTypeServerPage,
		})
	case *legacypacket_v630.CorrectPlayerMovePrediction:
		newPks = append(newPks, &packet.CorrectPlayerMovePrediction{
			Position: pk.Position,
			Delta:    pk.Delta,
			OnGround: pk.OnGround,
			Tick:     pk.Tick,
		})
	case *legacypacket_v630.LevelChunk:
		newPks = append(newPks, &packet.LevelChunk{
			Position:        pk.Position,
			HighestSubChunk: pk.HighestSubChunk,
			SubChunkCount:   pk.SubChunkCount,
			CacheEnabled:    pk.CacheEnabled,
			BlobHashes:      pk.BlobHashes,
			RawPayload:      pk.RawPayload,
		})
	case *legacypacket_v630.PlayerAuthInput:
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
	case *legacypacket_v630.PlayerList:
		newPks = append(newPks, &packet.PlayerList{
			ActionType: pk.ActionType,
			Entries: lo.Map(pk.Entries, func(item types_v630.PlayerListEntry, _ int) protocol.PlayerListEntry {
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
		case *packet.CameraInstruction:
			data := make(map[string]any)
			if v, ok := pk.Set.Value(); ok {
				setInstruction := map[string]any{
					"preset": v.Preset,
				}

				if v, ok := v.Ease.Value(); ok {
					setInstruction["ease"] = map[string]any{
						"type": types.CameraEaseString(v.Type),
						"time": v.Duration,
					}
				}
				if v, ok := v.Position.Value(); ok {
					setInstruction["pos"] = map[string]mgl32.Vec3{
						"pos": v,
					}
				}
				if v, ok := v.Rotation.Value(); ok {
					setInstruction["rot"] = map[string]any{
						"x": v[0],
						"y": v[1],
					}
				}
				if v, ok := v.Facing.Value(); ok {
					setInstruction["facing"] = map[string]mgl32.Vec3{
						"facing": v,
					}
				}
				if v, ok := v.Default.Value(); ok {
					setInstruction["default"] = mapBoolToByte(v)
				}
				data["set"] = setInstruction

			}
			if v, ok := pk.Clear.Value(); ok {
				data["clear"] = mapBoolToByte(v)
			}
			if v, ok := pk.Fade.Value(); ok {
				t, _ := v.TimeData.Value()
				c, _ := v.Colour.Value()
				data["fade"] = map[string]any{
					"time": map[string]float32{
						"fadeIn":  t.FadeInDuration,
						"hold":    t.WaitDuration,
						"fadeOut": t.FadeOutDuration,
					},
					"color": map[string]float32{
						"red":   float32(c.R) / 255,
						"green": float32(c.G) / 255,
						"blue":  float32(c.B) / 255,
					},
				}
			}

			result[i] = &legacypacket.CameraInstruction{Data: data}
		case *packet.CameraPresets:
			var data []map[string]any
			for _, preset := range pk.Presets {
				nbtPreset := map[string]any{
					"identifier":   preset.Name,
					"inherit_from": preset.Parent,
				}

				if v, ok := preset.PosX.Value(); ok {
					nbtPreset["pos_x"] = v
				}
				if v, ok := preset.PosY.Value(); ok {
					nbtPreset["pos_y"] = v
				}
				if v, ok := preset.PosZ.Value(); ok {
					nbtPreset["pos_z"] = v
				}
				if v, ok := preset.RotX.Value(); ok {
					nbtPreset["rot_x"] = v
				}
				if v, ok := preset.RotY.Value(); ok {
					nbtPreset["rot_y"] = v
				}
				if v, ok := preset.AudioListener.Value(); ok {
					switch v {
					case protocol.AudioListenerCamera:
						nbtPreset["audio_listener_type"] = "camera"
					case protocol.AudioListenerPlayer:
						nbtPreset["audio_listener_type"] = "player"
					}
				}
				if v, ok := preset.PlayerEffects.Value(); ok {
					nbtPreset["player_effects"] = mapBoolToByte(v)
				}
				data = append(data, nbtPreset)
			}
			result[i] = &legacypacket.CameraPresets{
				Data: map[string]any{
					"presets": data,
				},
			}
		case *packet.ResourcePacksInfo:
			// TODOnt: pack urls
			result[i] = &legacypacket.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				BehaviourPacks:      pk.BehaviourPacks,
				TexturePacks:        pk.TexturePacks,
				ForcingServerPacks:  pk.ForcingServerPacks,
			}
		case *packet.StartGame:
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
				EditorWorld:                    pk.EditorWorldType != packet.EditorWorldTypeNotEditor,
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
		case *packet.Disconnect:
			result[i] = &legacypacket_v618.Disconnect{
				HideDisconnectionScreen: pk.HideDisconnectionScreen,
				Message:                 pk.Message,
			}
		case *packet.ShowStoreOffer:
			result[i] = &legacypacket_v622.ShowStoreOffer{
				OfferID: pk.OfferID,
			}
		case *packet.CorrectPlayerMovePrediction:
			result[i] = &legacypacket_v630.CorrectPlayerMovePrediction{
				Position: pk.Position,
				Delta:    pk.Delta,
				OnGround: pk.OnGround,
				Tick:     pk.Tick,
			}
		case *packet.LevelChunk:
			result[i] = &legacypacket_v630.LevelChunk{
				Position:        pk.Position,
				HighestSubChunk: pk.HighestSubChunk,
				SubChunkCount:   pk.SubChunkCount,
				CacheEnabled:    pk.CacheEnabled,
				BlobHashes:      pk.BlobHashes,
				RawPayload:      pk.RawPayload,
			}
		case *packet.PlayerAuthInput:
			result[i] = &legacypacket_v630.PlayerAuthInput{
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
			result[i] = &legacypacket_v630.PlayerList{
				ActionType: pk.ActionType,
				Entries: lo.Map(pk.Entries, func(item protocol.PlayerListEntry, _ int) types_v630.PlayerListEntry {
					return types_v630.PlayerListEntry{
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

func getValueFromMap[V any](m map[string]any, key string) V {
	v, ok := m[key]
	if !ok {
		panic(fmt.Errorf("missing `%v` in map", key))
	}
	value, ok := v.(V)
	if !ok {
		panic(fmt.Errorf("`%v` has the incorrect type (got %T, expected %T)", key, v, new(V)))
	}
	return value
}

func getOptionalValueFromMap[V any](m map[string]any, key string) protocol.Optional[V] {
	v, ok := m[key]
	if !ok {
		return protocol.Optional[V]{}
	}
	value, ok := v.(V)
	if !ok {
		panic(fmt.Errorf("`%v` has the incorrect type (got %T, expected %T)", key, v, new(V)))
	}
	return protocol.Option(value)
}

func mapOptional[V1 any, V2 any](o protocol.Optional[V1], f func(v V1) V2) protocol.Optional[V2] {
	v, ok := o.Value()
	if !ok {
		return protocol.Optional[V2]{}
	}
	return protocol.Option(f(v))
}

func mapByteToBool(rawByte byte) bool {
	return rawByte != 0
}

func mapBoolToByte(rawBool bool) byte {
	if rawBool {
		return 1
	}
	return 0
}
