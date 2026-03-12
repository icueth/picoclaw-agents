package line

import (
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/channels"
	"picoclaw/agent/pkg/config"
)

func init() {
	channels.RegisterFactory("line", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewLINEChannel(cfg.Channels.LINE, b)
	})
}
