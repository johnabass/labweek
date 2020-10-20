package main

import (
	"fmt"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"plugin"
	"runtime/pprof"

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
		),
		arrangehttp.Server().
			ServerFactory(arrangehttp.ServerConfig{
				Address: ":8080", // default
			}).
			ProvideKey("servers.main"),
		fx.Invoke(
			func(in struct {
				fx.In
				Router        *mux.Router `name:"servers.main"`
				PluginHandler PluginHandler
				ScriptHandler ScriptHandler
			}) {
				in.Router.Handle("/plugin", in.PluginHandler)
				in.Router.Handle("/script", in.ScriptHandler)

				// TODO: arrangehttp should really provide a pprof integration

				in.Router.HandleFunc("/debug/pprof/", httppprof.Index)
				in.Router.HandleFunc("/debug/pprof/cmdline", httppprof.Cmdline)
				in.Router.HandleFunc("/debug/pprof/profile", httppprof.Profile)
				in.Router.HandleFunc("/debug/pprof/symbol", httppprof.Symbol)
				in.Router.HandleFunc("/debug/pprof/trace", httppprof.Trace)

				for _, p := range pprof.Profiles() {
					in.Router.HandleFunc("/debug/pprof/"+p.Name(), httppprof.Index)
				}
			},
		),
	)

	app.Run()
}
