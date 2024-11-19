//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/mugglemath/go-dew/internal/handler"
)

func InitializeApp(config *Config) (*handler.Handler, error) {
	wire.Build(
		ProvideDB,
		ProvideWeatherClient,
		ProvideDiscordClient,
		ProvideHandler,
	)
	return nil, nil
}
