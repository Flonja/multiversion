package v662

import (
	"fmt"
	"github.com/flonja/multiversion/protocols/v662/types"
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

func (r *Reader) LimitsEnabled() bool {
	return r.Reader.LimitsEnabled()
}

// Recipe reads a Recipe from the reader.
func (r *Reader) Recipe(x *protocol.Recipe) {
	var recipeType int32
	r.Varint32(&recipeType)
	if !lookupRecipe(recipeType, x) {
		r.UnknownEnumOption(recipeType, "crafting data recipe type")
		return
	}
	(*x).Marshal(r)
}

// lookupRecipe looks up the Recipe for a recipe type. False is returned if not
// found.
func lookupRecipe(recipeType int32, x *protocol.Recipe) bool {
	switch recipeType {
	case protocol.RecipeShapeless:
		*x = &protocol.ShapelessRecipe{}
	case protocol.RecipeShaped:
		*x = &types.ShapedRecipe{}
	case protocol.RecipeFurnace:
		*x = &protocol.FurnaceRecipe{}
	case protocol.RecipeFurnaceData:
		*x = &protocol.FurnaceDataRecipe{}
	case protocol.RecipeMulti:
		*x = &protocol.MultiRecipe{}
	case protocol.RecipeShulkerBox:
		*x = &protocol.ShulkerBoxRecipe{}
	case protocol.RecipeShapelessChemistry:
		*x = &protocol.ShapelessChemistryRecipe{}
	case protocol.RecipeShapedChemistry:
		*x = &types.ShapedChemistryRecipe{}
	case protocol.RecipeSmithingTransform:
		*x = &protocol.SmithingTransformRecipe{}
	case protocol.RecipeSmithingTrim:
		*x = &protocol.SmithingTrimRecipe{}
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

// Recipe writes a Recipe to the writer.
func (w *Writer) Recipe(x *protocol.Recipe) {
	var recipeType int32
	if !lookupRecipeType(*x, &recipeType) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "crafting recipe type")
	}
	w.Varint32(&recipeType)
	switch r := (*x).(type) {
	case *protocol.ShapedRecipe:
		if r.AssumeSymmetry {
			w.InvalidValue(r.AssumeSymmetry, "assume symmetry", "doesn't exist for <=1.20.70")
		}
		*x = &types.ShapedRecipe{
			RecipeID:        r.RecipeID,
			Width:           r.Width,
			Height:          r.Height,
			Input:           r.Input,
			Output:          r.Output,
			UUID:            r.UUID,
			Block:           r.Block,
			Priority:        r.Priority,
			RecipeNetworkID: r.RecipeNetworkID,
		}
	case *protocol.ShapedChemistryRecipe:
		if r.AssumeSymmetry {
			w.InvalidValue(r.AssumeSymmetry, "assume symmetry", "doesn't exist for <=1.20.70")
		}
		*x = &types.ShapedChemistryRecipe{
			ShapedRecipe: types.ShapedRecipe{
				RecipeID:        r.RecipeID,
				Width:           r.Width,
				Height:          r.Height,
				Input:           r.Input,
				Output:          r.Output,
				UUID:            r.UUID,
				Block:           r.Block,
				Priority:        r.Priority,
				RecipeNetworkID: r.RecipeNetworkID,
			},
		}
	}
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
