package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/flonja/multiversion/packbuilder"
	_ "github.com/flonja/multiversion/protocols" // VERY IMPORTANT
	v486 "github.com/flonja/multiversion/protocols/v486"
	v582 "github.com/flonja/multiversion/protocols/v582"
	v589 "github.com/flonja/multiversion/protocols/v589"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sirupsen/logrus"
)

func runServer() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	i := testItem{}
	world.RegisterItem(i)
	creative.RegisterItem(item.NewStack(i, 1))

	uc := server.DefaultConfig()
	conf, err := uc.Config(log)
	if err != nil {
		log.Fatalln(err)
	}
	pv486 := v486.New()

	resources := conf.Resources
	if autoGen, ok := packbuilder.BuildResourcePack(world.CustomItems(), pv486.Ver()); ok && !conf.DisableResourceBuilding {
		resources = append(resources, autoGen)
	}
	conf.DisableResourceBuilding = true

	conf.Listeners = []func(conf server.Config) (server.Listener, error){
		func(conf server.Config) (server.Listener, error) {
			cfg := minecraft.ListenConfig{
				MaximumPlayers:         conf.MaxPlayers,
				StatusProvider:         statusProvider{name: conf.Name},
				AuthenticationDisabled: conf.AuthDisabled,
				ResourcePacks:          resources,
				Biomes:                 biomes(),
				TexturePacksRequired:   conf.ResourcesRequired,
				AcceptedProtocols:      []minecraft.Protocol{pv486, v582.New(), v589.New()},
			}
			l, err := cfg.Listen("raknet", uc.Network.Address)
			if err != nil {
				return nil, fmt.Errorf("create minecraft listener: %w", err)
			}
			conf.Log.Infof("Server running on %v.\n", l.Addr())
			return listener{l}, nil
		},
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()
	srv.World().StopTime()
	srv.World().SetTime(1000)

	srv.Listen()
	for srv.Accept(func(p *player.Player) {
		p.ShowCoordinates()
		p.SetGameMode(world.GameModeCreative)
		p.Inventory().Clear()
		_, _ = p.Inventory().AddItem(item.NewStack(item.MusicDisc{DiscType: sound.DiscRelic()}, 1))
		_, _ = p.Inventory().AddItem(item.NewStack(i, 1))
	}) {
	}
}

// listener is a Listener implementation that wraps around a minecraft.Listener so that it can be listened on by
// Server.
type listener struct {
	*minecraft.Listener
}

// Accept blocks until the next connection is established and returns it. An error is returned if the Listener was
// closed using Close.
func (l listener) Accept() (session.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn.(session.Conn), err
}

// Disconnect disconnects a connection from the Listener with a reason.
func (l listener) Disconnect(conn session.Conn, reason string) error {
	return l.Listener.Disconnect(conn.(*minecraft.Conn), reason)
}

// statusProvider handles the way the server shows up in the server list. The
// online players and maximum players are not changeable from outside the
// server, but the server name may be changed at any time.
type statusProvider struct {
	name string
}

// ServerStatus returns the player count, max players and the server's name as
// a minecraft.ServerStatus.
func (s statusProvider) ServerStatus(playerCount, maxPlayers int) minecraft.ServerStatus {
	return minecraft.ServerStatus{
		ServerName:  s.name,
		PlayerCount: playerCount,
		MaxPlayers:  maxPlayers,
	}
}

// ashyBiome represents a biome that has any form of ash.
type ashyBiome interface {
	// Ash returns the ash and white ash of the biome.
	Ash() (ash float64, whiteAsh float64)
}

// sporingBiome represents a biome that has blue or red spores.
type sporingBiome interface {
	// Spores returns the blue and red spores of the biome.
	Spores() (blueSpores float64, redSpores float64)
}

// biomes builds a mapping of all biome definitions of the server, ready to be set in the biomes field of the server
// listener.
func biomes() map[string]any {
	definitions := make(map[string]any)
	for _, b := range world.Biomes() {
		definition := map[string]any{
			"name_hash":   b.String(), // This isn't actually a hash despite what the field name may suggest.
			"temperature": float32(b.Temperature()),
			"downfall":    float32(b.Rainfall()),
			"rain":        b.Rainfall() > 0,
		}
		if a, ok := b.(ashyBiome); ok {
			ash, whiteAsh := a.Ash()
			definition["ash"], definition["white_ash"] = float32(ash), float32(whiteAsh)
		}
		if s, ok := b.(sporingBiome); ok {
			blueSpores, redSpores := s.Spores()
			definition["blue_spores"], definition["red_spores"] = float32(blueSpores), float32(redSpores)
		}
		definitions[b.String()] = definition
	}
	return definitions
}
