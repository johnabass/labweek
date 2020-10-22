package main

import (
	"net"

	"go.uber.org/zap"
)

type listener struct {
	net.Listener
	l *zap.Logger
}

func (l listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err == nil {
		l.l.Info(
			"accepted new connection",
			zap.Stringer("remoteAddr", c.RemoteAddr()),
		)
	}

	return c, err
}

func DecorateListener(l *zap.Logger, next net.Listener) net.Listener {
	return listener{
		Listener: next,
		l:        l,
	}
}
