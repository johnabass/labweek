package main

import (
	"fmt"
	"net/http"
	"os"
	"plugin"

	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
	"github.com/spf13/viper"
	"github.com/xmidt-org/arrange"
	"github.com/xmidt-org/arrange/arrangehttp"
	"go.uber.org/fx"
)

func main() {
	l, v, err := setup(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app := fx.New(
		arrange.LoggerFunc(l.Sugar().Infof),
		arrange.ForViper(v),
		fx.Supply(v),
		fx.Provide(
			func(v *viper.Viper) (*plugin.Plugin, error) {
				return plugin.Open(
					v.GetString("plugin.path"),
				)
			},
			func(v *viper.Viper, p *plugin.Plugin) (PluginHandler, error) {
				name := v.GetString("pluginHandler.symbol")
				s, err := p.Lookup(name)
				if err == nil {
					if handle, ok := s.(func(http.ResponseWriter, *http.Request)); ok {
						return PluginHandler{H: handle}, nil
					}

					err = fmt.Errorf("Symbol %s is not a handler function", name)
				}

				return PluginHandler{}, err
			},
			func(v *viper.Viper) (*otto.Script, error) {
				vm := otto.New()
				return vm.Compile(
					"script",
					v.Get("script"),
				)
			},
			func(s *otto.Script) ScriptHandler {
				return ScriptHandler{S: s}
			},
			arrangehttp.Server().
				ServerFactory(arrangehttp.ServerConfig{
					Address: ":8080", // default
				}).
				UnmarshalKey("servers.main"),
		),
		fx.Invoke(
			func(r *mux.Router, ph PluginHandler, sh ScriptHandler) {
				r.Handle("/plugin", ph)
				r.Handle("/script", sh)
			},
		),
	)

	app.Run()
}
