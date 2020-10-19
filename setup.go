package main

import (
	"github.com/spf13/viper"
	"github.com/xmidt-org/arrange"
	"github.com/xmidt-org/sallust"
	"go.uber.org/zap"
)

func setup(args []string) (l *zap.Logger, v *viper.Viper, err error) {
	var configLocations = []string{
		".",
		"$HOME/.labweek",
		"/etc/labweek",
	}

	v = viper.New()
	v.SetConfigName("labweek")
	for _, location := range configLocations {
		v.AddConfigPath(location)
	}

	err = v.ReadInConfig()
	if err == nil {
		var c sallust.Config
		if err = v.UnmarshalKey("logging", &c, arrange.DefaultDecodeHooks); err == nil {
			l, err = c.Build()
		}
	}

	if l != nil {
		l.Info(
			"bootstrap successful",
			zap.String("configFile", v.ConfigFileUsed()),
		)
	}

	return
}
