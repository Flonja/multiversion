package v649

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v649/packet"
	"github.com/flonja/multiversion/translator"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte
)

type Protocol struct {
	blockMapping    mapping.Block
	blockTranslator translator.BlockTranslator
}

func New() *Protocol {
	blockMapping := mapping.NewBlockMapping(blockStateData)
	latestBlockMapping := latest.NewBlockMapping()
	return &Protocol{blockMapping: blockMapping, blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ID() int32 {
	return 649
}

func (p Protocol) Ver() string {
	return "1.20.62"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDLecternUpdate] = func() packet.Packet { return &legacypacket.LecternUpdate{} }
	pool[packet.IDMobEffect] = func() packet.Packet { return &legacypacket.MobEffect{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDResourcePacksInfo] = func() packet.Packet { return &legacypacket.ResourcePacksInfo{} }
	pool[packet.IDActorEvent] = func() packet.Packet { return &legacypacket.SetActorMotion{} }
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
	case *legacypacket.LecternUpdate:
		if pk.DropBook {
			break
		}
		newPks = append(newPks, &packet.LecternUpdate{
			Page:      pk.Page,
			PageCount: pk.PageCount,
			Position:  pk.Position,
		})
	case *legacypacket.MobEffect:
		newPks = append(newPks, &packet.MobEffect{
			EntityRuntimeID: pk.EntityRuntimeID,
			Operation:       pk.Operation,
			EffectType:      pk.EffectType,
			Amplifier:       pk.Amplifier,
			Particles:       pk.Particles,
			Duration:        pk.Duration,
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
			InteractionModel:       pk.InteractionModel,
			GazeDirection:          pk.GazeDirection,
			Tick:                   pk.Tick,
			Delta:                  pk.Delta,
			ItemInteractionData:    pk.ItemInteractionData,
			ItemStackRequest:       pk.ItemStackRequest,
			BlockActions:           pk.BlockActions,
			ClientPredictedVehicle: pk.ClientPredictedVehicle,
			AnalogueMoveVector:     pk.AnalogueMoveVector,
		})
	case *legacypacket.ResourcePacksInfo:
		newPks = append(newPks, &packet.ResourcePacksInfo{
			TexturePackRequired: pk.TexturePackRequired,
			HasScripts:          pk.HasScripts,
			BehaviourPacks:      pk.BehaviourPacks,
			TexturePacks:        pk.TexturePacks,
			ForcingServerPacks:  pk.ForcingServerPacks,
			PackURLs:            pk.PackURLs,
		})
	case *legacypacket.SetActorMotion:
		newPks = append(newPks, &packet.SetActorMotion{
			EntityRuntimeID: pk.EntityRuntimeID,
			Velocity:        pk.Velocity,
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
	default:
		newPks = append(newPks, pk)
	}

	return p.blockTranslator.UpgradeBlockPackets(newPks, conn)
}

func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) (result []packet.Packet) {
	result = p.blockTranslator.DowngradeBlockPackets([]packet.Packet{pk}, conn)

	for i, pk := range result {
		switch pk := pk.(type) {
		case *packet.LecternUpdate:
			result[i] = &legacypacket.LecternUpdate{
				Page:      pk.Page,
				PageCount: pk.PageCount,
				Position:  pk.Position,
			}
		case *packet.MobEffect:
			result[i] = &legacypacket.MobEffect{
				EntityRuntimeID: pk.EntityRuntimeID,
				Operation:       pk.Operation,
				EffectType:      pk.EffectType,
				Amplifier:       pk.Amplifier,
				Particles:       pk.Particles,
				Duration:        pk.Duration,
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
				InteractionModel:       pk.InteractionModel,
				GazeDirection:          pk.GazeDirection,
				Tick:                   pk.Tick,
				Delta:                  pk.Delta,
				ItemInteractionData:    pk.ItemInteractionData,
				ItemStackRequest:       pk.ItemStackRequest,
				BlockActions:           pk.BlockActions,
				ClientPredictedVehicle: pk.ClientPredictedVehicle,
				AnalogueMoveVector:     pk.AnalogueMoveVector,
			}
		case *packet.ResourcePacksInfo:
			result[i] = &legacypacket.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				BehaviourPacks:      pk.BehaviourPacks,
				TexturePacks:        pk.TexturePacks,
				ForcingServerPacks:  pk.ForcingServerPacks,
				PackURLs:            pk.PackURLs,
			}
		case *packet.SetActorMotion:
			result[i] = &legacypacket.SetActorMotion{
				EntityRuntimeID: pk.EntityRuntimeID,
				Velocity:        pk.Velocity,
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
		}
	}

	return result
}
