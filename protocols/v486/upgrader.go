package v486

import (
	"github.com/flonja/multiversion/protocols/latest"
	"github.com/flonja/multiversion/protocols/v486/mappings"
)

// upgradeBlockRuntimeID upgrades legacy block runtime IDs to a latest block runtime ID.
func upgradeBlockRuntimeID(input uint32) uint32 {
	state, ok := mappings.RuntimeIDToState(input)
	if !ok {
		return latestAirRID
	}
	runtimeID, ok := latest.StateToRuntimeID(state)
	if !ok {
		return latestAirRID
	}
	return runtimeID
}
