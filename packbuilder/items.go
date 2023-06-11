package packbuilder

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/image/colornames"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	_ "unsafe" // Imported for compiler directives.
)

// buildItems builds all the item-related files for the resource pack. This includes textures, language
// entries and item atlas.
func buildItems(dir string, customItems []world.CustomItem) (count int, lang []string) {
	if err := os.Mkdir(filepath.Join(dir, "items"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "textures/items"), os.ModePerm); err != nil {
		panic(err)
	}

	textureData := make(map[string]any)
	for _, item := range customItems {
		identifier, _ := item.EncodeItem()
		lang = append(lang, fmt.Sprintf("item.%s.name=%s", identifier, item.Name()))

		name := strings.Split(identifier, ":")[1]
		textureData[name] = map[string]string{"textures": fmt.Sprintf("textures/items/%s.png", name)}

		buildItemTexture(dir, name, item.Texture())

		count++
	}

	buildItemAtlas(dir, map[string]any{
		"resource_pack_name": "vanilla",
		"texture_name":       "atlas.items",
		"texture_data":       textureData,
	})
	return
}

// buildItemTexture creates a PNG file for the item from the provided image and name and writes it to the pack.
func buildItemTexture(dir, name string, img image.Image) {
	texture, err := os.Create(filepath.Join(dir, "textures/items", name+".png"))
	if err != nil {
		panic(err)
	}
	if img == nil {
		im := image.NewAlpha(image.Rect(0, 0, 64, 64))
		for x := 0; x < im.Bounds().Dx(); x++ {
			for y := 0; y < im.Bounds().Dy(); y++ {
				im.Set(x, y, colornames.Goldenrod)
			}
		}
		img = im
	}

	if err := png.Encode(texture, img); err != nil {
		_ = texture.Close()
		panic(err)
	}
	if err := texture.Close(); err != nil {
		panic(err)
	}
}

// buildItemAtlas creates the identifier to texture mapping and writes it to the pack.
func buildItemAtlas(dir string, atlas map[string]any) {
	b, err := json.Marshal(atlas)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "textures/item_texture.json"), b, 0666); err != nil {
		panic(err)
	}
}
