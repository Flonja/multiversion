package mappings

import (
	"bytes"
	_ "embed"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/flonja/multiversion/internal"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte

	// states holds a list of all possible vanilla block states.
	states []blockupgrader.BlockState
	// stateRuntimeIDs holds a map for looking up the runtime ID of a block by the stateHash it produces.
	stateRuntimeIDs = map[internal.StateHash]uint32{}
	// runtimeIDToState holds a map for looking up the blockState of a block by its runtime ID.
	runtimeIDToState = map[uint32]blockupgrader.BlockState{}
)

// init initializes the state mappings.
func init() {
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))

	var s blockupgrader.BlockState
	for {
		if err := dec.Decode(&s); err != nil {
			break
		}

		rid := uint32(len(states))
		states = append(states, s)

		stateRuntimeIDs[internal.HashState(s)] = rid
		runtimeIDToState[rid] = s
	}
}

// StateToRuntimeID converts a block state to a runtime ID.
func StateToRuntimeID(state blockupgrader.BlockState) (runtimeID uint32, found bool) {
	rid, ok := stateRuntimeIDs[internal.HashState(blockupgrader.Upgrade(state))]
	return rid, ok
}

// RuntimeIDToState converts a runtime ID to a name and its state properties.
func RuntimeIDToState(runtimeID uint32) (state blockupgrader.BlockState, found bool) {
	state, found = runtimeIDToState[runtimeID]
	return state, found
}
