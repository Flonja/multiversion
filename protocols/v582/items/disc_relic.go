package items

import (
	"github.com/df-mc/dragonfly/server/item/category"
	"golang.org/x/image/colornames"
	"image"
)

type DiscRelic struct {
}

func (d DiscRelic) EncodeItem() (name string, meta int16) {
	return "multiversion:mv_record_relic", 0
}

func (d DiscRelic) Name() string {
	return "Music Disc - Relic"
}

func (d DiscRelic) Texture() image.Image {
	im := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for x := 0; x < im.Bounds().Dx(); x++ {
		for y := 0; y < im.Bounds().Dy(); y++ {
			im.SetRGBA(x, y, colornames.Goldenrod)
		}
	}
	return im
}

func (d DiscRelic) Category() category.Category {
	return category.Items()
}
