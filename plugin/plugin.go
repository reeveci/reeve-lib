package plugin

import (
	"encoding/gob"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/reeveci/reeve-lib/schema"
)

type PluginConfig struct {
	Plugin Plugin
	Logger hclog.Logger
}

func Serve(config *PluginConfig) {
	RegisterSharedTypes()

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: Handshake,

		Plugins: goplugin.PluginSet{
			"plugin": &ReevePlugin{Impl: config.Plugin},
		},

		Logger: config.Logger,
	})
}

func RegisterSharedTypes() {
	// params
	gob.Register(schema.LiteralParam(""))
	gob.Register(schema.EnvParam{})
	gob.Register(schema.VarParam{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	// plugin arguments
	gob.Register(map[string]string{})
	gob.Register(schema.PipelineStatus{})
}
