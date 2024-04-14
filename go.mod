module github.com/flonja/multiversion

go 1.21

require (
	github.com/df-mc/dragonfly v0.9.15
	github.com/df-mc/worldupgrader v1.0.13
	github.com/go-gl/mathgl v1.1.0
	github.com/google/uuid v1.6.0
	github.com/rogpeppe/go-internal v1.12.0
	github.com/samber/lo v1.39.0
	github.com/sandertv/go-raknet v1.12.0
	github.com/sandertv/gophertunnel v1.36.0
	github.com/segmentio/fasthash v1.0.3
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/exp v0.0.0-20240409090435-93d18d7e34b8
	golang.org/x/image v0.15.0
	golang.org/x/oauth2 v0.19.0
)

require (
	github.com/brentp/intintmap v0.0.0-20190211203843-30dc0ade9af9 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/df-mc/atomic v1.10.0 // indirect
	github.com/df-mc/goleveldb v1.1.9 // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/muhammadmuzzammil1998/jsonc v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/sandertv/go-raknet => github.com/tedacmc/tedac-raknet v0.0.4

replace github.com/sandertv/gophertunnel => github.com/flonja/tedac-gophertunnel v0.0.12-0.20240414174355-f004af9b69ac
