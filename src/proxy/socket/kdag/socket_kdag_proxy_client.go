package kdag

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

// SocketKdagProxyClient is the client component of the KdagProxy that sends
// RPC requests to Kdag
type SocketKdagProxyClient struct {
	nodeAddr string
	timeout  time.Duration
	rpc      *rpc.Client
}

// NewSocketKdagProxyClient implements a new SocketKdagProxyClient
func NewSocketKdagProxyClient(nodeAddr string, timeout time.Duration) *SocketKdagProxyClient {
	return &SocketKdagProxyClient{
		nodeAddr: nodeAddr,
		timeout:  timeout,
	}
}

func (p *SocketKdagProxyClient) getConnection() error {
	if p.rpc == nil {
		conn, err := net.DialTimeout("tcp", p.nodeAddr, p.timeout)

		if err != nil {
			return err
		}

		p.rpc = jsonrpc.NewClient(conn)
	}

	return nil
}

// SubmitTx submits a transaction to Kdag
func (p *SocketKdagProxyClient) SubmitTx(tx []byte) (*bool, error) {
	if err := p.getConnection(); err != nil {
		return nil, err
	}

	var ack bool

	err := p.rpc.Call("Kdag.SubmitTx", tx, &ack)

	if err != nil {
		p.rpc = nil

		return nil, err
	}

	return &ack, nil
}
