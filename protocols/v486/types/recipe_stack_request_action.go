package types

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type TakeStackRequestAction struct {
	protocol.TakeStackRequestAction
}

func (a *TakeStackRequestAction) Marshal(r protocol.IO) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
}

type PlaceStackRequestAction struct {
	protocol.PlaceStackRequestAction
}

func (a *PlaceStackRequestAction) Marshal(r protocol.IO) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
}

type SwapStackRequestAction struct {
	protocol.SwapStackRequestAction
}

func (a *SwapStackRequestAction) Marshal(r protocol.IO) {
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
}

type DropStackRequestAction struct {
	protocol.DropStackRequestAction
}

func (a *DropStackRequestAction) Marshal(r protocol.IO) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	r.Bool(&a.Randomly)
}

type DestroyStackRequestAction struct {
	protocol.DestroyStackRequestAction
}

func (a *DestroyStackRequestAction) Marshal(r protocol.IO) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
}

type ConsumeStackRequestAction struct {
	protocol.DestroyStackRequestAction
}

type PlaceInContainerStackRequestAction struct {
	protocol.PlaceInContainerStackRequestAction
}

func (a *PlaceInContainerStackRequestAction) Marshal(r protocol.IO) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
}

type TakeOutContainerStackRequestAction struct {
	protocol.TakeOutContainerStackRequestAction
}

func (a *TakeOutContainerStackRequestAction) Marshal(r protocol.IO) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
}

type AutoCraftRecipeStackRequestAction struct {
	protocol.AutoCraftRecipeStackRequestAction
}

func (a *AutoCraftRecipeStackRequestAction) Marshal(r protocol.IO) {
	r.Varuint32(&a.RecipeNetworkID)
	r.Uint8(&a.TimesCrafted)
}

func StackReqSlotInfo(r protocol.IO, x *protocol.StackRequestSlotInfo) {
	if _, ok := r.(interface{ Reads() bool }); !ok && x.ContainerID > 21 {
		x.ContainerID -= 1
	}
	r.Uint8(&x.ContainerID)
	if _, ok := r.(interface{ Reads() bool }); ok && x.ContainerID >= 21 { // RECIPE_BOOK
		x.ContainerID += 1
	}
	r.Uint8(&x.Slot)
	r.Varint32(&x.StackNetworkID)
}
