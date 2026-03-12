package pico

import (
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/channels"
	"picoclaw/agent/pkg/config"
)

func init() {
	channels.RegisterFactory("pico", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewPicoChannel(cfg.Channels.Pico, b)
	})
}
