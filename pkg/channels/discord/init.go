package discord

import (
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/channels"
	"picoclaw/agent/pkg/config"
)

func init() {
	channels.RegisterFactory("discord", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewDiscordChannel(cfg.Channels.Discord, b)
	})
}
