package whatsapp

import (
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/channels"
	"picoclaw/agent/pkg/config"
)

func init() {
	channels.RegisterFactory("whatsapp", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewWhatsAppChannel(cfg.Channels.WhatsApp, b)
	})
}
