// Package kdag implements a component that the TCP AppProxy can connect to.
package kdag

import (
	"fmt"
	"time"

	"github.com/Kdag-K/kdag/src/proxy"
	"github.com/sirupsen/logrus"
)

// SocketKdagProxy is a Golang implementation of a service that binds to a
// remote Kdag over an RPC/TCP connection. It implements handlers for the RPC
// requests sent by the SocketAppProxy, and submits transactions to Kdag via
// an RPC request. A SocketKdagProxy can be implemented in any programming
// language as long as it implements the AppProxy interface over RPC.
type SocketKdagProxy struct {
	nodeAddress string
	bindAddress string

	handler proxy.ProxyHandler

	client *SocketKdagProxyClient
	server *SocketKdagProxyServer
}

// NewSocketKdagProxy creates a new SocketKdagProxy
func NewSocketKdagProxy(
	nodeAddr string,
	bindAddr string,
	handler proxy.ProxyHandler,
	timeout time.Duration,
	logger *logrus.Entry,
) (*SocketKdagProxy, error) {

	if logger == nil {
		log := logrus.New()
		log.Level = logrus.DebugLevel
		logger = logrus.NewEntry(log)
	}

	client := NewSocketKdagProxyClient(nodeAddr, timeout)

	server, err := NewSocketKdagProxyServer(bindAddr, handler, timeout, logger)

	if err != nil {
		return nil, err
	}

	proxy := &SocketKdagProxy{
		nodeAddress: nodeAddr,
		bindAddress: bindAddr,
		handler:     handler,
		client:      client,
		server:      server,
	}

	go proxy.server.listen()

	return proxy, nil
}

// SubmitTx submits a transaction to Kdag
func (p *SocketKdagProxy) SubmitTx(tx []byte) error {
	ack, err := p.client.SubmitTx(tx)

	if err != nil {
		return err
	}

	if !*ack {
		return fmt.Errorf("Failed to deliver transaction to Kdag")
	}

	return nil
}
