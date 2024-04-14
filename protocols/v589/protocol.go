package v486

import (
	_ "embed"
	"github.com/flonja/multiversion/mapping"
	"github.com/flonja/multiversion/protocols/latest"
	legacypacket "github.com/flonja/multiversion/protocols/v589/packet"
	"github.com/flonja/multiversion/protocols/v589/types"
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
	itemMapping := mapping.NewItemMapping(itemRuntimeIDData)
	blockMapping := mapping.NewBlockMapping(blockStateData)
	latestBlockMapping := latest.NewBlockMapping()
	return &Protocol{itemMapping: itemMapping, blockMapping: blockMapping,
		itemTranslator:  translator.NewItemTranslator(itemMapping, latest.NewItemMapping(), blockMapping, latestBlockMapping),
		blockTranslator: translator.NewBlockTranslator(blockMapping, latestBlockMapping)}
}

func (p Protocol) ID() int32 {
	return 589
}

func (p Protocol) Ver() string {
	return "1.20.1"
}

func (Protocol) Packets(_ bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDAvailableCommands] = func() packet.Packet { return &legacypacket.AvailableCommands{} }
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
	case *legacypacket.AvailableCommands:
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
		case *packet.AvailableCommands:
			result[i] = &legacypacket.AvailableCommands{
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
