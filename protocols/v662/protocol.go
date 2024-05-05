package v662

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v662/packet"
	"github.com/flonja/multiversion/translator"
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
	return 662
}

func (p Protocol) Ver() string {
	return "1.20.73"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDClientBoundDebugRenderer] = func() packet.Packet { return &legacypacket.ClientBoundDebugRenderer{} }
	pool[packet.IDCorrectPlayerMovePrediction] = func() packet.Packet { return &legacypacket.CorrectPlayerMovePrediction{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDResourcePackStack] = func() packet.Packet { return &legacypacket.ResourcePackStack{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDUpdateBlockSynced] = func() packet.Packet { return &legacypacket.UpdateBlockSynced{} }
	pool[packet.IDUpdatePlayerGameType] = func() packet.Packet { return &legacypacket.UpdatePlayerGameType{} }
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
	case *legacypacket.ClientBoundDebugRenderer:
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
	case *legacypacket.CorrectPlayerMovePrediction:
		newPks = append(newPks, &packet.CorrectPlayerMovePrediction{
			PredictionType: pk.PredictionType,
			Position:       pk.Position,
			Delta:          pk.Delta,
			OnGround:       pk.OnGround,
			Tick:           pk.Tick,
		})
	case *legacypacket.PlayerAuthInput:
		newPks = append(newPks, &packet.PlayerAuthInput{
			Pitch:                  pk.Pitch,
			Yaw:                    pk.Yaw,
			Position:               pk.Position,
			MoveVector:             pk.MoveVector,
			HeadYaw:                pk.HeadYaw,
			InputData:              pk.InputData,
			InputMode:              pk.InputMode,
			PlayMode:               pk.PlayMode,
			InteractionModel:       uint32(pk.InteractionModel),
			GazeDirection:          pk.GazeDirection,
			Tick:                   pk.Tick,
			Delta:                  pk.Delta,
			ItemInteractionData:    pk.ItemInteractionData,
			ItemStackRequest:       pk.ItemStackRequest,
			BlockActions:           pk.BlockActions,
			ClientPredictedVehicle: pk.ClientPredictedVehicle,
			VehicleRotation:        pk.VehicleRotation,
			AnalogueMoveVector:     pk.AnalogueMoveVector,
		})
	case *legacypacket.ResourcePackStack:
		newPks = append(newPks, &packet.ResourcePackStack{
			TexturePackRequired:          pk.TexturePackRequired,
			BehaviourPacks:               pk.BehaviourPacks,
			TexturePacks:                 pk.TexturePacks,
			BaseGameVersion:              pk.BaseGameVersion,
			Experiments:                  pk.Experiments,
			ExperimentsPreviouslyToggled: pk.ExperimentsPreviouslyToggled,
		})
	case *legacypacket.StartGame:
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
	case *legacypacket.UpdateBlockSynced:
		newPks = append(newPks, &packet.UpdateBlockSynced{
			Position:          pk.Position,
			NewBlockRuntimeID: pk.NewBlockRuntimeID,
			Flags:             pk.Flags,
			Layer:             pk.Layer,
			EntityUniqueID:    uint64(pk.EntityUniqueID),
			TransitionType:    pk.TransitionType,
		})
	case *legacypacket.UpdatePlayerGameType:
		newPks = append(newPks, &packet.UpdatePlayerGameType{
			GameType:       pk.GameType,
			PlayerUniqueID: pk.PlayerUniqueID,
		})
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
		case *packet.ClientBoundDebugRenderer:
			result[i] = &legacypacket.ClientBoundDebugRenderer{
				Type:     pk.Type,
				Text:     pk.Text,
				Position: pk.Position,
				Red:      pk.Red,
				Green:    pk.Green,
				Blue:     pk.Blue,
				Alpha:    pk.Alpha,
				Duration: pk.Duration,
			}
		case *packet.CorrectPlayerMovePrediction:
			result[i] = &legacypacket.CorrectPlayerMovePrediction{
				PredictionType: pk.PredictionType,
				Position:       pk.Position,
				Delta:          pk.Delta,
				OnGround:       pk.OnGround,
				Tick:           pk.Tick,
			}
		case *packet.PlayerAuthInput:
			result[i] = &legacypacket.PlayerAuthInput{
				Pitch:                  pk.Pitch,
				Yaw:                    pk.Yaw,
				Position:               pk.Position,
				MoveVector:             pk.MoveVector,
				HeadYaw:                pk.HeadYaw,
				InputData:              pk.InputData,
				InputMode:              pk.InputMode,
				PlayMode:               pk.PlayMode,
				InteractionModel:       int32(pk.InteractionModel),
				GazeDirection:          pk.GazeDirection,
				Tick:                   pk.Tick,
				Delta:                  pk.Delta,
				ItemInteractionData:    pk.ItemInteractionData,
				ItemStackRequest:       pk.ItemStackRequest,
				BlockActions:           pk.BlockActions,
				ClientPredictedVehicle: pk.ClientPredictedVehicle,
				VehicleRotation:        pk.VehicleRotation,
				AnalogueMoveVector:     pk.AnalogueMoveVector,
			}
		case *packet.ResourcePackStack:
			result[i] = &legacypacket.ResourcePackStack{
				TexturePackRequired:          pk.TexturePackRequired,
				BehaviourPacks:               pk.BehaviourPacks,
				TexturePacks:                 pk.TexturePacks,
				BaseGameVersion:              pk.BaseGameVersion,
				Experiments:                  pk.Experiments,
				ExperimentsPreviouslyToggled: pk.ExperimentsPreviouslyToggled,
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
			result[i] = &legacypacket.UpdateBlockSynced{
				Position:          pk.Position,
				NewBlockRuntimeID: pk.NewBlockRuntimeID,
				Flags:             pk.Flags,
				Layer:             pk.Layer,
				EntityUniqueID:    int64(pk.EntityUniqueID),
				TransitionType:    pk.TransitionType,
			}
		case *packet.UpdatePlayerGameType:
			result[i] = &legacypacket.UpdatePlayerGameType{
				GameType:       pk.GameType,
				PlayerUniqueID: pk.PlayerUniqueID,
			}
		}
	}

	return result
}
