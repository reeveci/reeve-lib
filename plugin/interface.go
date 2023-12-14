package plugin

import (
	"io"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/reeveci/reeve-lib/schema"
)

type Capabilities struct {
	Message    bool
	Discover   bool
	Resolve    bool
	Notify     bool
	CLIMethods map[string]string
}

type ReeveAPI interface {
	NotifyMessages(messages []schema.Message) error
	NotifyTriggers(triggers []schema.Trigger) error
	io.Closer
}

type Plugin interface {
	Name() (string, error)
	Register(settings map[string]string, api ReeveAPI) (Capabilities, error)
	Unregister() error

	Message(source string, message schema.Message) error
	Discover(trigger schema.Trigger) ([]schema.Pipeline, error)
	Resolve(env []string) (map[string]schema.Env, error)
	Notify(status schema.PipelineStatus) error
	CLIMethod(method string, args []string) (string, error)
}

var Handshake = goplugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "REEVE_PLUGIN",
	MagicCookieValue: "reeveci",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]goplugin.Plugin{
	"plugin": &ReevePlugin{},
}
