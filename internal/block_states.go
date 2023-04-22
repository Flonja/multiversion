package internal

import (
	"fmt"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"sort"
	"strings"
	"unsafe"
)

// StateHash is a struct that may be used as a map key for block states. It contains the name of the block state
// and an encoded version of the properties.
type StateHash struct {
	Name, Properties string
}

// HashState produces a hash for the block properties held by the blockState.
func HashState(state blockupgrader.BlockState) StateHash {
	if state.Properties == nil {
		return StateHash{Name: state.Name}
	}
	keys := make([]string, 0, len(state.Properties))
	for k := range state.Properties {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var b strings.Builder
	for _, k := range keys {
		switch v := state.Properties[k].(type) {
		case bool:
			if v {
				b.WriteByte(1)
			} else {
				b.WriteByte(0)
			}
		case uint8:
			b.WriteByte(v)
		case int32:
			a := *(*[4]byte)(unsafe.Pointer(&v))
			b.Write(a[:])
		case string:
			b.WriteString(v)
		default:
			// If block encoding is broken, we want to find out as soon as possible. This saves a lot of time
			// debugging in-game.
			panic(fmt.Sprintf("invalid block property type %T for property %v", v, k))
		}
	}
	return StateHash{Name: state.Name, Properties: b.String()}
}
