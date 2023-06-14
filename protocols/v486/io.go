package v486

import (
	"fmt"
	"github.com/flonja/multiversion/protocols/v486/types"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// NewReader creates a new initialised Reader with an underlying protocol.Reader to write to.
func NewReader(r *protocol.Reader) *Reader {
	return &Reader{r}
}

type Reader struct {
	*protocol.Reader
}

func (r *Reader) Reads() bool {
	return true
}

func (r *Reader) StackRequestAction(x *protocol.StackRequestAction) {
	var id uint8
	r.Uint8(&id)
	if !lookupStackRequestAction(id, x) {
		r.UnknownEnumOption(id, "stack request action type")
		return
	}
	(*x).Marshal(r)
}

// lookupStackRequestAction looks up the StackRequestAction matching an ID.
func lookupStackRequestAction(id uint8, x *protocol.StackRequestAction) bool {
	switch id {
	case protocol.StackRequestActionTake:
		*x = &types.TakeStackRequestAction{TakeStackRequestAction: protocol.TakeStackRequestAction{}}
	case protocol.StackRequestActionPlace:
		*x = &types.PlaceStackRequestAction{PlaceStackRequestAction: protocol.PlaceStackRequestAction{}}
	case protocol.StackRequestActionSwap:
		*x = &types.SwapStackRequestAction{SwapStackRequestAction: protocol.SwapStackRequestAction{}}
	case protocol.StackRequestActionDrop:
		*x = &types.DropStackRequestAction{DropStackRequestAction: protocol.DropStackRequestAction{}}
	case protocol.StackRequestActionDestroy:
		*x = &types.DestroyStackRequestAction{DestroyStackRequestAction: protocol.DestroyStackRequestAction{}}
	case protocol.StackRequestActionConsume:
		*x = &types.ConsumeStackRequestAction{DestroyStackRequestAction: protocol.DestroyStackRequestAction{}}
	case protocol.StackRequestActionCreate:
		*x = &protocol.CreateStackRequestAction{}
	case protocol.StackRequestActionPlaceInContainer:
		*x = &types.PlaceInContainerStackRequestAction{PlaceInContainerStackRequestAction: protocol.PlaceInContainerStackRequestAction{}}
	case protocol.StackRequestActionTakeOutContainer:
		*x = &types.TakeOutContainerStackRequestAction{TakeOutContainerStackRequestAction: protocol.TakeOutContainerStackRequestAction{}}
	case protocol.StackRequestActionLabTableCombine:
		*x = &protocol.LabTableCombineStackRequestAction{}
	case protocol.StackRequestActionBeaconPayment:
		*x = &protocol.BeaconPaymentStackRequestAction{}
	case protocol.StackRequestActionMineBlock:
		*x = &protocol.MineBlockStackRequestAction{}
	case protocol.StackRequestActionCraftRecipe:
		*x = &protocol.CraftRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftRecipeAuto:
		*x = &types.AutoCraftRecipeStackRequestAction{AutoCraftRecipeStackRequestAction: protocol.AutoCraftRecipeStackRequestAction{}}
	case protocol.StackRequestActionCraftCreative:
		*x = &protocol.CraftCreativeStackRequestAction{}
	case protocol.StackRequestActionCraftRecipeOptional:
		*x = &protocol.CraftRecipeOptionalStackRequestAction{}
	case protocol.StackRequestActionCraftGrindstone:
		*x = &protocol.CraftGrindstoneRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftLoom:
		*x = &protocol.CraftLoomRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftNonImplementedDeprecated:
		*x = &protocol.CraftNonImplementedStackRequestAction{}
	case protocol.StackRequestActionCraftResultsDeprecated:
		*x = &protocol.CraftResultsDeprecatedStackRequestAction{}
	default:
		return false
	}
	return true
}

// NewWriter creates a new initialised Writer with an underlying protocol.Writer to write to.
func NewWriter(w *protocol.Writer) *Writer {
	return &Writer{w}
}

type Writer struct {
	*protocol.Writer
}

// ItemDescriptorCount writes an ItemDescriptorCount i to the underlying buffer.
func (w *Writer) ItemDescriptorCount(i *protocol.ItemDescriptorCount) {
	var id byte
	switch descriptor := i.Descriptor.(type) {
	case *protocol.InvalidItemDescriptor:
		id = protocol.ItemDescriptorInvalid
	case *protocol.DefaultItemDescriptor:
		id = protocol.ItemDescriptorDefault
	case *protocol.MoLangItemDescriptor:
		id = protocol.ItemDescriptorMoLang
	case *protocol.ItemTagItemDescriptor:
		id = protocol.ItemDescriptorItemTag
	case *protocol.DeferredItemDescriptor:
		id = protocol.ItemDescriptorDeferred
	case *protocol.ComplexAliasItemDescriptor:
		id = protocol.ItemDescriptorComplexAlias
	case *types.DefaultItemDescriptor:
		descriptor.Marshal(w)
		if descriptor.NetworkID != 0 {
			w.Varint32(&i.Count)
		}
		return
	default:
		w.UnknownEnumOption(fmt.Sprintf("%T", i.Descriptor), "item descriptor type")
		return
	}
	w.Uint8(&id)

	i.Descriptor.Marshal(w)
	w.Varint32(&i.Count)
}

// Recipe writes a Recipe to the writer.
func (w *Writer) Recipe(x *protocol.Recipe) {
	var recipeType int32
	if !lookupRecipeType(*x, &recipeType) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "crafting recipe type")
	}
	w.Varint32(&recipeType)
	(*x).Marshal(w)
}

// lookupRecipeType looks up the recipe type for a Recipe. False is returned if
// none was found.
func lookupRecipeType(x protocol.Recipe, recipeType *int32) bool {
	switch x.(type) {
	case *protocol.ShapelessRecipe:
		*recipeType = protocol.RecipeShapeless
	case *protocol.ShapedRecipe:
		*recipeType = protocol.RecipeShaped
	case *protocol.FurnaceRecipe:
		*recipeType = protocol.RecipeFurnace
	case *protocol.FurnaceDataRecipe:
		*recipeType = protocol.RecipeFurnaceData
	case *protocol.MultiRecipe:
		*recipeType = protocol.RecipeMulti
	case *protocol.ShulkerBoxRecipe:
		*recipeType = protocol.RecipeShulkerBox
	case *protocol.ShapelessChemistryRecipe:
		*recipeType = protocol.RecipeShapelessChemistry
	case *protocol.ShapedChemistryRecipe:
		*recipeType = protocol.RecipeShapedChemistry
	case *protocol.SmithingTransformRecipe:
		*recipeType = protocol.RecipeSmithingTransform
	case *protocol.SmithingTrimRecipe:
		*recipeType = protocol.RecipeSmithingTrim
	default:
		return false
	}
	return true
}
