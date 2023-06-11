package packbuilder

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/rogpeppe/go-internal/dirhash"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"os"
)

// BuildResourcePack builds a resource pack based on custom features that have been registered to the server.
// It creates a UUID based on the hash of the directory so the client will only be prompted to download it
// once it is changed.
func BuildResourcePack(customItems []world.CustomItem, version string) (*resource.Pack, bool) {
	dir, err := os.MkdirTemp("", "dragonfly_resource_pack-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	var assets int
	var lang []string

	itemCount, itemLang := buildItems(dir, customItems)
	assets += itemCount
	lang = append(lang, itemLang...)

	if assets > 0 {
		buildLanguageFile(dir, lang)
		hash, err := dirhash.HashDir(dir, "", dirhash.Hash1)
		if err != nil {
			panic(err)
		}
		var header, module [16]byte
		copy(header[:], hash)
		copy(module[:], hash[16:])
		buildManifest(dir, version, header, module)
		return resource.MustCompile(dir), true
	}
	return nil, false
}
