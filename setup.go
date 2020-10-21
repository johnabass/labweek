package main

import (
	"context"
	"os"
	"runtime/pprof"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xmidt-org/arrange"
	"github.com/xmidt-org/sallust"
	"go.uber.org/zap"
)

type Profiling struct {
	CPU    string
	Memory string

	cpuFile *os.File
}

func (p *Profiling) Start(context.Context) (err error) {
	if len(p.CPU) > 0 {
		p.cpuFile, err = os.Create(p.CPU)
		if err != nil {
			return
		}

		if err = pprof.StartCPUProfile(p.cpuFile); err != nil {
			p.cpuFile.Close()
			return
		}
	}

	return
}

func (p *Profiling) Stop(context.Context) (err error) {
	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		p.cpuFile.Close()
		p.cpuFile = nil
	}

	if len(p.Memory) > 0 {
		var memFile *os.File
		if memFile, err = os.Create(p.Memory); err != nil {
			return
		}

		defer memFile.Close()
		err = pprof.WriteHeapProfile(memFile)
	}

	return
}

func newViper() (*viper.Viper, error) {
	var configLocations = []string{
		".",
		"$HOME/.labweek",
		"/etc/labweek",
	}

	v := viper.New()
	v.SetConfigName("labweek")
	for _, location := range configLocations {
		v.AddConfigPath(location)
	}

	return v, v.ReadInConfig()
}

func parseCmdLine(args []string, v *viper.Viper) (l *zap.Logger, p *Profiling, err error) {
	p = new(Profiling)
	fs := pflag.NewFlagSet("labweek", pflag.ExitOnError)
	fs.StringVar(&p.CPU, "cpuprofile", "", "turns on cpu profiling and writes it to the given file")
	fs.StringVar(&p.Memory, "memprofile", "", "turns on memory profiling and writes it to the given file")

	fs.Parse(args)
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

		if len(p.CPU) > 0 {
			l.Info(
				"writing CPU profile",
				zap.String("file", p.CPU),
			)
		}

		if len(p.Memory) > 0 {
			l.Info(
				"writing memory profile",
				zap.String("file", p.Memory),
			)
		}
	}

	return
}
